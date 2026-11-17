import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/live-rebalance.ts', import.meta.url), 'utf8');
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const { requestLiveRebalance, applyLiveRebalance, rebalanceTotals, rebalanceLotValue } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
const stock = (code, lots) => ({ name: code, code, group_name: 'G', price: 5, lot_amount: lots, lot_size: 2 });
const bond = { name: 'Bond', code: 'B', group_name: 'D', isin: '', clean_price: { currency: 'USD', price: 10 }, accrued_interest: { currency: 'RUB', price: 2 }, maturity_date: '', maturity_payment: null, lot_amount: 3, lot_size: 1 };

test('proposals use separate endpoints and translate final lot counts', async () => {
	const calls = [];
	const api = { stocks: { rebalance: async (...args) => { calls.push(['stock', ...args]); return { stocks: [stock('S', 4)] }; } }, bonds: { rebalance: async (...args) => { calls.push(['bond', ...args]); return { bonds: [{ ...bond, lot_amount: 1 }] }; } } };
	const trades = await requestLiveRebalance(api, { stocks: [stock('S', 2)], bonds: [bond] }, { stocks: 10, bonds: -20, noSell: false });
	assert.deepEqual(calls, [['stock', false, 10], ['bond', false, -20]]);
	assert.deepEqual(trades.map(trade => [trade.kind, trade.holding.lot_amount, trade.target]), [['stock', 2, 4], ['bond', 3, 1]]);
});

test('confirmed changes remain committed and uncertain failures cannot be retried', async () => {
	const calls = [], committed = [];
	const api = { stocks: { add: async value => calls.push(value), remove: async () => { throw Error('lost connection'); } }, bonds: {} };
	const trades = [{ kind: 'stock', holding: stock('A', 2), target: 4 }, { kind: 'stock', holding: stock('C', 3), target: 1 }];
	await assert.rejects(applyLiveRebalance(api, trades, {}, trade => committed.push(trade.holding.lot_amount)), /Предыдущие подтверждённые операции сохранены/);
	assert.deepEqual(committed, [4]);
	assert.equal(calls[0][0].lot_amount, 2);
	await assert.rejects(applyLiveRebalance(api, trades, {}, () => assert.fail()), /Результат предыдущего запроса неизвестен/);
	assert.equal(calls.length, 1);
});

test('bond transactions use the dirty price per bond in RUB when currencies differ', async () => {
	let operation;
	const api = { stocks: {}, bonds: { remove: async values => operation = values[0] } };
	await applyLiveRebalance(api, [{ kind: 'bond', holding: bond, target: 1 }], { USD: 90 }, () => {});
	assert.equal(operation.lot_amount, 2);
	assert.deepEqual(operation.price, { currency: 'RUB', price: 902 });
});

test('rejects stale or invalid server proposals before offering trades', async () => {
	const api = { stocks: { rebalance: async () => ({ stocks: [stock('S', -1)] }) }, bonds: {} };
	await assert.rejects(requestLiveRebalance(api, { stocks: [stock('S', 2)], bonds: [] }, { stocks: 0, bonds: null, noSell: false }), /не соответствует/);
	api.stocks.rebalance = async () => ({ stocks: [stock('S', 1)] });
	await assert.rejects(requestLiveRebalance(api, { stocks: [stock('S', 2)], bonds: [] }, { stocks: 10, bonds: null, noSell: true }), /не соответствует/);
});

test('transaction totals use lot sizes and live rates, including missing exchange rates', () => {
	const trades = [{ kind: 'stock', holding: stock('S', 2), target: 4 }, { kind: 'bond', holding: bond, target: 1 }];
	assert.equal(rebalanceLotValue(trades[1], { USD: 90 }), 902);
	assert.deepEqual(rebalanceTotals(trades, { USD: 90 }), { stocks: 20, bonds: -1804, total: -1784 });
	assert.deepEqual(rebalanceTotals(trades, {}), { stocks: 20, bonds: null, total: null });
	assert.deepEqual(rebalanceTotals([{ ...trades[0], target: undefined }], {}), { stocks: null, bonds: null, total: null });
});

test('live proposal groups editable transaction quantities by side and asset', async () => {
	const view = await readFile(new URL('../src/lib/components/tabs/live-rebalance.svelte', import.meta.url), 'utf8');
	assert.match(view, /label: 'Купить'/);
	assert.match(view, /label: 'Продать'/);
	assert.match(view, /bind:value=\{row\.quantity\}/);
	assert.match(view, /rebalanceTotals\(proposed, rates\)/);
});

test('live screen requests proposals and confirms recording without broker orders', async () => {
	const view = await readFile(new URL('../src/lib/components/tabs/live-rebalance.svelte', import.meta.url), 'utf8');
	const page = await readFile(new URL('../src/routes/+page.svelte', import.meta.url), 'utf8');
	const app = await readFile(new URL('../src/lib/components/portfolio-app.svelte', import.meta.url), 'utf8');
	assert.match(view, /не отправляет поручения брокеру и не меняет портфель/);
	assert.match(view, /только если сделки уже выполнены/);
	assert.match(view, /цены из загруженного портфеля/);
	assert.match(view, /blocked = true/);
	assert.match(page, /await applyLiveRebalance/);
	assert.match(page, /stockView\(data\).*bondView\(data/);
	assert.match(app, /<LiveRebalance/);
});
