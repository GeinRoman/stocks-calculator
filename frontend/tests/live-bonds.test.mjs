import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const url = code => `data:text/javascript;base64,${Buffer.from(code).toString('base64')}`;
const read = async path => compile(await readFile(new URL(`../src/lib/${path}.ts`, import.meta.url), 'utf8'));
let code = await read('logic/live-bonds');
const editor = (await read('logic/tabs/bond-editor')).replace('./stock-editor.js', url(await read('logic/tabs/stock-editor')));
code = code.replace('#lib/logic/tabs/bond-editor.js', url(editor)).replace('#lib/api/bond-view.js', url(await read('api/bond-view'))).replace('#lib/logic/tabs/currency-service.js', url(await read('logic/tabs/currency-service')));
const { searchLiveBonds, saveLiveBonds, BondSaveError } = await import(url(code));
const bond = { code: 'A', name: 'A', isin: '', group_name: 'Old', clean_price: { currency: 'USD', price: 10 }, accrued_interest: { currency: 'USD', price: 1 }, maturity_date: '', maturity_payment: null, lot_size: 2, lot_amount: 3 };
const state = () => ({ data: { bond_groups: [{ name: 'Old', weight: 100 }], bonds: [bond] }, rates: { USD: 90 }, unavailableRates: [], loadedAt: '2026-01-01T00:00:00Z' });
const draft = [{ id: 'new', name: 'New', targetPercent: 0, bonds: [{ code: 'B', name: 'B', isin: '', cleanPrice: { currency: 'RUB', price: 100 }, accruedInterest: { currency: 'RUB', price: 5 }, maturityDate: null, maturityPayment: null, lotSize: 1, lotCount: 2 }] }];

test('search maps backend quote and reports errors', async () => {
	const api = { find: async query => { assert.deepEqual(query, ['bond']); return [{ err_msg: '', search_results: [bond] }]; } };
	assert.deepEqual(await searchLiveBonds(api, ' bond '), [{ code: 'A', name: 'A', isin: '', cleanPrice: bond.clean_price, accruedInterest: bond.accrued_interest, maturityDate: null, maturityPayment: null, lotSize: 2 }]);
	await assert.rejects(searchLiveBonds({ find: async () => [{ err_msg: 'not found' }] }, 'x'), /not found/);
});

test('each 200 commits before the next operation and retry skips earlier successes', async () => {
	let current = state(), fail = true; const calls = [];
	const api = { remove: async value => calls.push(['sell', value]), removeGroups: async value => calls.push(['delete', value]), addGroups: async value => calls.push(['group', value]), add: async value => { calls.push(['buy', value]); if (fail) throw { status: 400 }; } };
	await assert.rejects(saveLiveBonds(api, current, draft, value => current = value), error => error instanceof BondSaveError && !error.uncertain);
	assert.deepEqual(calls.map(([name]) => name), ['sell', 'delete', 'group', 'buy']);
	assert.deepEqual(calls[0][1][0].price, { currency: 'USD', price: 11 });
	assert.equal(calls[0][1][0].lot_amount, 3);
	assert.deepEqual(current.data.bond_groups, [{ name: 'New', weight: 0 }]);
	assert.deepEqual(current.data.bonds, []);
	fail = false;
	await saveLiveBonds(api, current, draft, value => current = value);
	assert.equal(calls.filter(([name]) => name === 'sell').length, 1);
	assert.deepEqual(calls.at(-1)[1][0].price, { currency: 'RUB', price: 105 });
	assert.equal(current.data.bonds[0].lot_amount, 2);
});

test('mixed currencies require live rates and missing rates do not send a transaction', async () => {
	let current = state(), sent = 0;
	current.data.bonds = [{ ...bond, accrued_interest: { currency: 'RUB', price: 2 } }];
	current.rates = {};
	const api = { remove: async () => sent++ };
	await assert.rejects(saveLiveBonds(api, current, [], value => current = value), /цены с НКД или курса/);
	assert.equal(sent, 0);
	current.rates = { USD: 90 };
	let price;
	await assert.rejects(saveLiveBonds({ remove: async values => { price = values[0].price; throw { status: 500 }; } }, current, [], () => {}), error => error instanceof BondSaveError && error.uncertain);
	assert.deepEqual(price, { currency: 'RUB', price: 902 });
});

test('foreign quote without exchange rate remains unavailable after a successful purchase', async () => {
	let current = { ...state(), data: { bond_groups: [], bonds: [] }, rates: {}, unavailableRates: [] };
	const foreign = [{ id: 'new', name: 'New', targetPercent: 0, bonds: [{ code: 'A', name: 'A', isin: '', cleanPrice: bond.clean_price, accruedInterest: bond.accrued_interest, maturityDate: null, maturityPayment: null, lotSize: 2, lotCount: 1 }] }];
	await saveLiveBonds({ addGroups: async () => {}, add: async () => {} }, current, foreign, value => current = value);
	assert.deepEqual(current.unavailableRates, ['USD']);
	assert.equal(current.bonds.totalBondsValueRub, null);
});
