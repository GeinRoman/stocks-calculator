import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

async function load(path) {
	const source = await readFile(new URL(path, import.meta.url), 'utf8');
	const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
	return import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
}
const { ApiClient } = await load('../src/lib/api/client.ts');
const { portfolioApi } = await load('../src/lib/api/portfolio-api.ts');
const { historyLoader } = await load('../src/lib/api/transaction-history.ts');
const { couponSchedule } = await load('../src/lib/api/coupons.ts');
const { requestLiveRebalance, applyLiveRebalance } = await load('../src/lib/logic/live-rebalance.ts');
const stock = (code, lot_amount) => ({ name: code, code, group_name: 'G', price: 10, lot_amount, lot_size: 1 });
const json = value => new Response(JSON.stringify(value), { status: 200 });

test('mock HTTP flow authenticates, switches profile and loads history and coupons on demand', async () => {
	const calls = [], storage = new Map();
	let selected = 'A';
	const client = new ApiClient({ now: () => 0, storage: { getItem: key => storage.get(key) ?? null, setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) }, fetch: async (url, init) => {
		calls.push([url, init.method, init.body]);
		if (url === '/login') return json({ token: { access_token: 'test', expires_at: '2099-01-01T00:00:00Z' }, ref_token: 'refresh', profiles: [] });
		assert.equal(init.headers.get('Authorization'), 'Bearer test');
		if (url === '/setdefaultprofile') { selected = JSON.parse(init.body).name; return new Response(null, { status: 200 }); }
		if (url === '/getall') return json({ profiles: [{ name: selected, default: true }] });
		if (url.startsWith('/getstocktransactions')) return json([{ id: 7, instrument_code: selected, lot_amount: 1, lot_size: 1, buying: true, price: 10, currency: 'RUB', time: '2026-01-01T00:00:00Z' }]);
		if (url === '/bondcoupons?code=B') return json({ code: 'B', coupons: [{ payment_date: '2030-01-01', amount: null }] });
		return new Response('unexpected', { status: 500 });
	} });
	const api = portfolioApi(client);
	await client.authenticate('/login', 'user', 'password');
	assert.equal((await api.getAll()).profiles[0].name, 'A');
	assert.equal(calls.some(([path]) => path.includes('transactions') || path.includes('coupons')), false);
	await api.profiles.setDefault({ name: 'B', default: false });
	assert.equal((await api.getAll()).profiles[0].name, 'B');
	const history = await historyLoader(api)('stock', 'B', { before: '2026-02-01T00:00:00Z', beforeId: 10 });
	assert.equal(history.items[0].instrument_code, 'B');
	assert.match(calls.at(-1)[0], /before_id=10/);
	assert.equal(couponSchedule('B', await api.bonds.coupons('B')).coupons[0].amount, null);
});

test('mock HTTP rebalance keeps first successful trade and refuses retry after later failure', async () => {
	const calls = [], storage = new Map([['stocks-calculator-refresh-token', 'refresh']]);
	const client = new ApiClient({ now: () => 0, storage: { getItem: key => storage.get(key) ?? null, setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) }, fetch: async (url, init) => {
		calls.push([url, init.body]);
		if (url === '/reftoken') return json({ access_token: 'test', expires_at: '2099-01-01T00:00:00Z' });
		if (url.startsWith('/rebalance?')) return json({ stocks: [stock('A', 3), stock('B', 1)], total_prev_cost: 50, cost_diff: 0 });
		if (url === '/addstocks') return new Response(null, { status: 200 });
		if (url === '/removestocks') return new Response('failed', { status: 500 });
		return new Response('unexpected', { status: 500 });
	} });
	const api = portfolioApi(client);
	const trades = await requestLiveRebalance(api, { stocks: [stock('A', 2), stock('B', 3)], bonds: [] }, { stocks: 0, bonds: null, noSell: false });
	const committed = [];
	await assert.rejects(applyLiveRebalance(api, trades, {}, trade => committed.push(trade.holding.code)), /Предыдущие подтверждённые операции сохранены/);
	assert.deepEqual(committed, ['A']);
	await assert.rejects(applyLiveRebalance(api, trades, {}, () => assert.fail('repeated')), /неизвестен/);
	assert.equal(calls.filter(([url]) => url === '/addstocks').length, 1);
	assert.equal(calls.filter(([url]) => url === '/removestocks').length, 1);
});

test('profile rename and snapshot installation use the authenticated API contracts', async () => {
	const calls = [], storage = new Map([['stocks-calculator-refresh-token', 'refresh']]);
	const client = new ApiClient({ now: () => 0, storage: { getItem: key => storage.get(key) ?? null, setItem: (key, value) => storage.set(key, value), removeItem: key => storage.delete(key) }, fetch: async (url, init) => {
		if (url === '/reftoken') return json({ access_token: 'test', expires_at: '2099-01-01T00:00:00Z' });
		assert.equal(init.headers.get('Authorization'), 'Bearer test');
		calls.push([url, init.method, JSON.parse(init.body)]);
		return new Response(null, { status: 200 });
	} });
	const api = portfolioApi(client);
	const snapshot = { version: 1, profiles: [] };
	await api.profiles.rename('Old', 'New');
	await api.installSnapshot(snapshot);
	assert.deepEqual(calls, [
		['/renameprofile', 'PATCH', { name: 'Old', new_name: 'New' }],
		['/installsnapshot', 'POST', snapshot]
	]);
});
