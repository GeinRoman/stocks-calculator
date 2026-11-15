import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const url = code => `data:text/javascript;base64,${Buffer.from(code).toString('base64')}`;
let code = compile(await readFile(new URL('../src/lib/logic/live-stocks.ts', import.meta.url), 'utf8'));
for (const path of ['api/stock-view', 'logic/tabs/stock-editor']) code = code.replace(`#lib/${path}.js`, url(compile(await readFile(new URL(`../src/lib/${path}.ts`, import.meta.url), 'utf8'))));
const { saveLiveStocks, searchLiveStocks, StockSaveError } = await import(url(code));
const stock = { code: 'A', name: 'A', group_name: 'Old', price: 12, lot_size: 2, lot_amount: 3 };
const state = () => ({ data: { stock_groups: [{ name: 'Old', weight: 100 }], stocks: [stock] }, stocks: {} });
const draft = [{ id: 'new', name: 'New', targetPercent: 0, stocks: [{ code: 'B', name: 'B', price: 20, lotSize: 5, lotCount: 2 }] }];

test('search uses server quotes and reports server errors', async () => {
	const api = { find: async names => { assert.deepEqual(names, ['query']); return [{ err_msg: '', search_results: [{ code: 'B', name: 'B', price: 20, lot_size: 5 }] }]; } };
	assert.deepEqual(await searchLiveStocks(api, ' query '), [{ code: 'B', name: 'B', price: 20, lotSize: 5 }]);
	await assert.rejects(searchLiveStocks({ find: async () => [{ err_msg: 'not found' }] }, 'x'), /not found/);
});

test('each 200 commits sequentially; retry does not repeat confirmed changes', async () => {
	const calls = []; let current = state(), fail = true;
	const api = {
		remove: async values => calls.push(['sell', values]),
		removeGroups: async values => calls.push(['delete', values]),
		addGroups: async values => calls.push(['group', values]),
		add: async values => { calls.push(['buy', values]); if (fail) throw { status: 400 }; }
	};
	await assert.rejects(saveLiveStocks(api, current, draft, value => current = value), error => error instanceof StockSaveError && !error.uncertain);
	assert.deepEqual(calls.map(([name]) => name), ['sell', 'delete', 'group', 'buy']);
	assert.equal(calls[0][1][0].price, 12);
	assert.deepEqual(current.data.stock_groups, [{ name: 'New', weight: 0 }]);
	assert.deepEqual(current.data.stocks, []);
	fail = false;
	await saveLiveStocks(api, current, draft, value => current = value);
	assert.deepEqual(calls.map(([name]) => name), ['sell', 'delete', 'group', 'buy', 'buy']);
	assert.equal(current.data.stocks[0].price, 20);
	assert.equal(current.data.stocks[0].lot_amount, 2);
});

test('unknown or server failure marks outcome uncertain and does not commit failed operation', async () => {
	let current = state();
	const api = { remove: async () => { throw new Error('network'); } };
	await assert.rejects(saveLiveStocks(api, current, draft, value => current = value), error => error instanceof StockSaveError && error.uncertain);
	assert.deepEqual(current.data.stocks, [stock]);
});
