import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

async function load(path) {
	const source = await readFile(new URL(path, import.meta.url), 'utf8');
	const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
	return import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);
}
const { ApiClient, HttpError, REFRESH_TOKEN_KEY, SessionExpiredError } = await load('../src/lib/api/client.ts');
const { portfolioApi } = await load('../src/lib/api/portfolio-api.ts');
const token = (access_token = 'access') => ({ access_token, expires_at: '2030-01-01T00:00:00Z' });
const json = (value, status = 200) => new Response(JSON.stringify(value), { status });
function store() {
	const values = new Map();
	return { getItem: key => values.get(key) ?? null, setItem: (key, value) => values.set(key, value), removeItem: key => values.delete(key) };
}

test('login stores refresh token, uses bearer auth and accepts empty 200 responses', async () => {
	const storage = store(), calls = [];
	const client = new ApiClient({ storage, now: () => 0, fetch: async (url, init) => {
		calls.push([url, init]);
		return url === '/login' ? json({ token: token(), ref_token: 'refresh', profiles: [] }) : new Response(null, { status: 200 });
	} });
	await client.authenticate('/login', 'alice', 'password');
	assert.equal(storage.getItem(REFRESH_TOKEN_KEY), 'refresh');
	await client.request('/addstocks', { method: 'POST', body: [{ code: 'SBER' }] });
	assert.equal(calls[1][1].headers.get('Authorization'), 'Bearer access');
	assert.equal(calls[1][1].body, '[{"code":"SBER"}]');
});

test('registration uses the create-user endpoint and failed authentication does not create a session', async () => {
	const storage = store(); const calls = [];
	const client = new ApiClient({ storage, now: () => 0, baseUrl: '/api', fetch: async (url, init) => {
		if (url === '/api/test') { assert.equal(init.headers.get('Authorization'), 'Bearer access'); return new Response(null, { status: 200 }); }
		calls.push([url, JSON.parse(init.body)]);
		return url === '/api/createuser' ? json({ token: token(), ref_token: 'new-refresh', profiles: [] }) : new Response('wrong credentials', { status: 401 });
	} });
	await assert.rejects(client.authenticate('/login', 'alice', 'wrong'), HttpError);
	assert.equal(storage.getItem(REFRESH_TOKEN_KEY), null);
	await client.authenticate('/createuser', 'alice', 'password');
	assert.deepEqual(calls, [['/api/login', { user: 'alice', pass: 'wrong' }], ['/api/createuser', { user: 'alice', pass: 'password' }]]);
	assert.equal(storage.getItem(REFRESH_TOKEN_KEY), 'new-refresh');
	await client.request('/test');
	client.logout();
	assert.equal(storage.getItem(REFRESH_TOKEN_KEY), null);
});

test('authentication rejects missing or invalid nested access tokens without saving the refresh token', async () => {
	for (const value of [null, {}, { access_token: '', expires_at: '2030-01-01T00:00:00Z' }, { access_token: 'access', expires_at: 'invalid' }]) {
		const storage = store();
		const client = new ApiClient({ storage, now: () => 0, fetch: async () => json({ token: value, ref_token: 'refresh', profiles: [] }) });
		await assert.rejects(client.authenticate('/login', 'user', 'pass'), /Invalid authentication response/);
		assert.equal(storage.getItem(REFRESH_TOKEN_KEY), null);
	}
});

test('expired tokens refresh before requests, concurrent refreshes share one request', async () => {
	const storage = store(); storage.setItem(REFRESH_TOKEN_KEY, 'refresh');
	let refreshes = 0;
	const client = new ApiClient({ storage, now: () => 0, fetch: async (url, init) => {
		if (url === '/reftoken') { refreshes++; assert.equal(init.body, '"refresh"'); return json(token()); }
		assert.equal(init.headers.get('Authorization'), 'Bearer access'); return json({ ok: true });
	} });
	assert.deepEqual(await Promise.all([client.request('/getall'), client.request('/getall')]), [{ ok: true }, { ok: true }]);
	assert.equal(refreshes, 1);
});

test('expiration is checked again before each request', async () => {
	const storage = store(); storage.setItem(REFRESH_TOKEN_KEY, 'refresh');
	let now = 0, refreshes = 0, requests = 0;
	const client = new ApiClient({ storage, now: () => now, fetch: async (url, init) => {
		if (url === '/reftoken') return json({ access_token: `token${++refreshes}`, expires_at: '2030-01-01T00:00:00Z' });
		requests++;
		if (requests === 2) return new Response(null, { status: 401 });
		return json(init.headers.get('Authorization'));
	} });
	assert.equal(await client.request('/getall'), 'Bearer token1');
	now = Date.parse('2030-01-01T00:00:00Z');
	// A new token must expire after the current clock, not merely parse as a date.
	await assert.rejects(client.request('/getall'));
	assert.equal(requests, 1);
	assert.equal(storage.getItem(REFRESH_TOKEN_KEY), null);
});

test('a 401 refreshes and successfully retries with the new token', async () => {
	const storage = store(); storage.setItem(REFRESH_TOKEN_KEY, 'refresh');
	let refreshes = 0, requests = 0;
	const client = new ApiClient({ storage, now: () => 0, fetch: async (url, init) => {
		if (url === '/reftoken') return json(token(`token${++refreshes}`));
		if (++requests === 1) return new Response(null, { status: 401 });
		return json(init.headers.get('Authorization'));
	} });
	assert.equal(await client.request('/getall'), 'Bearer token2');
	assert.equal(requests, 2); assert.equal(refreshes, 2);
});

test('401 refreshes and retries once; repeated 401 expires session', async () => {
	const storage = store(); storage.setItem(REFRESH_TOKEN_KEY, 'refresh');
	let refreshes = 0, expired = 0, requests = 0;
	const client = new ApiClient({ storage, now: () => 0, onSessionExpired: () => expired++, fetch: async url => {
		if (url === '/reftoken') return json(token(`access${++refreshes}`));
		requests++; return new Response(null, { status: 401 });
	} });
	await assert.rejects(client.request('/getall'), SessionExpiredError);
	assert.equal(refreshes, 2); assert.equal(requests, 2); assert.equal(expired, 1);
	assert.equal(storage.getItem(REFRESH_TOKEN_KEY), null);
});

test('refresh failure expires session; non-200 write fails without pretending success', async () => {
	const storage = store(); storage.setItem(REFRESH_TOKEN_KEY, 'refresh');
	let expired = 0;
	const client = new ApiClient({ storage, now: () => 0, onSessionExpired: () => expired++, fetch: async url => url === '/reftoken' ? json(token()) : new Response('invalid', { status: 400 }) });
	await assert.rejects(client.request('/addstocks', { method: 'POST', body: [] }), error => error instanceof HttpError && error.status === 400 && error.message === 'invalid');
	assert.equal(expired, 0);
	const other = new ApiClient({ storage, onSessionExpired: () => expired++, fetch: async () => new Response('invalid', { status: 404 }) });
	await assert.rejects(other.refresh(), HttpError);
	assert.equal(expired, 1); assert.equal(storage.getItem(REFRESH_TOKEN_KEY), null);
});

test('logout during refresh cannot resurrect the session', async () => {
	const storage = store(); storage.setItem(REFRESH_TOKEN_KEY, 'refresh');
	let resolve;
	const client = new ApiClient({ storage, now: () => 0, fetch: () => new Promise(r => resolve = r) });
	const pending = client.refresh(); client.logout(); resolve(json(token()));
	await assert.rejects(pending, SessionExpiredError);
	assert.equal(storage.getItem(REFRESH_TOKEN_KEY), null);
});

test('endpoint wrappers send profile changes, separate rebalance and lazy history/coupons without reloading', async () => {
	const calls = [];
	const api = portfolioApi({ request: async (path, options) => { calls.push({ path, ...options }); } });
	await api.profiles.setDefault({ name: 'Main', default: true });
	await api.stocks.rebalance(true, 100);
	await api.bonds.rebalance(false, -20);
	await api.bonds.coupons('A&B');
	await api.currencyRate('USD', 'RUB');
	await api.stocks.transactions({ before: '2026-01-01T00:00:00Z', before_id: 5 });
	await api.stocks.add([{ code: 'SBER' }]);
	assert.deepEqual(calls.map(call => call.path), ['/setdefaultprofile', '/rebalance?nosell=true&valdiff=100', '/rebalancebonds?nosell=false&valdiff=-20', '/bondcoupons?code=A%26B', '/currencyrate?from=USD&to=RUB', '/getstocktransactions?before=2026-01-01T00%3A00%3A00Z&before_id=5', '/addstocks']);
	assert.equal(calls[0].method, 'PATCH'); assert.equal(calls[6].method, 'POST');
});

test('/getall preserves all group lists and weights, including empty groups', async () => {
	const storage = store(); storage.setItem(REFRESH_TOKEN_KEY, 'refresh');
	const payload = {
		stocks: [], stock_groups: [{ name: 'Stocks', weight: 50 }],
		bonds: [], bond_groups: [{ name: 'Bonds', weight: 0 }],
		crypto_assets: [], crypto_groups: [], profiles: [],
		personal_assets: { deposits: [], credit_cards: [], insurances: [] }
	};
	const calls = [];
	const client = new ApiClient({ storage, now: () => 0, fetch: async url => {
		calls.push(url);
		return json(url === '/reftoken' ? token() : payload);
	} });
	assert.deepEqual(await portfolioApi(client).getAll(), payload);
	assert.deepEqual(calls, ['/reftoken', '/getall']);
});

test('session flows load all data only after authentication, restoration, or confirmed profile switch', async () => {
	const apiSource = await readFile(new URL('../src/lib/api/portfolio-api.ts', import.meta.url), 'utf8');
	const sessionSource = await readFile(new URL('../src/lib/api/session-api.ts', import.meta.url), 'utf8');
	const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
	const apiUrl = `data:text/javascript;base64,${Buffer.from(compile(apiSource)).toString('base64')}`;
	const sessionCode = compile(sessionSource).replace('./portfolio-api.js', apiUrl);
	const { sessionApi } = await import(`data:text/javascript;base64,${Buffer.from(sessionCode).toString('base64')}`);
	const calls = [];
	const session = sessionApi({
		authenticate: async path => { calls.push(path); return {}; },
		refresh: async () => { calls.push('/reftoken'); },
		request: async path => { calls.push(path); return {}; }
	});
	await session.login('user', 'pass'); await session.restore(); await session.selectProfile({ name: 'Main', default: true });
	assert.deepEqual(calls, ['/login', '/getall', '/reftoken', '/getall', '/setdefaultprofile', '/getall']);
});
