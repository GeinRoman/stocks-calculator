import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/live-rebalance.ts', import.meta.url), 'utf8');
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const { requestLiveRebalance, applyLiveRebalance } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
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
