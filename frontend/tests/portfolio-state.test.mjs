import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const url = code => `data:text/javascript;base64,${Buffer.from(code).toString('base64')}`;
let code = compile(await readFile(new URL('../src/lib/logic/portfolio-state.ts', import.meta.url), 'utf8'));
for (const name of ['bond-rates', 'bond-view', 'stock-view', 'personal-assets-view']) {
	const source = await readFile(new URL(`../src/lib/api/${name}.ts`, import.meta.url), 'utf8');
	code = code.replace(`#lib/api/${name}.js`, url(compile(source)));
}
const { loadPortfolioState } = await import(url(code));
const data = {
	profiles: [{ name: 'Main', default: true }],
	stocks: [{ name: 'Stock', code: 'S', group_name: 'Shares', price: 10, lot_amount: 2, lot_size: 5 }],
	stock_groups: [{ name: 'Shares', weight: 100 }],
	bonds: [{ name: 'Bond', code: 'B', isin: '', group_name: 'Debt', clean_price: { currency: 'USD', price: 10 }, accrued_interest: { currency: 'RUB', price: 1 }, maturity_date: '', maturity_payment: null, lot_amount: 1, lot_size: 2 }],
	bond_groups: [{ name: 'Debt', weight: 100 }],
	crypto_assets: [{ id: 1, name: 'Coin', group_name: 'Crypto', platform: 'Wallet', url: '', annual_interest: 0, amount: 2 }],
	crypto_groups: [{ name: 'Crypto', weight: 100 }],
	personal_assets: { deposits: [{ amount_rub: 25 }], insurances: [{ amount_rub: 30 }], credit_cards: [] }
};

test('portfolio loading gets all data once, then fetches only required rates', async () => {
	const calls = [];
	const state = await loadPortfolioState({
		getAll: async () => { calls.push('getall'); return data; },
		currencyRate: async (from, to) => { calls.push(`${from}-${to}`); return { from, to, rate: 90 }; }
	});
	assert.deepEqual(calls, ['getall', 'USD-RUB']);
	assert.equal(state.stocks.total, 100);
	assert.equal(state.stocks.groups[0].targetPercent, 100);
	assert.equal(state.bonds.totalBondsValueRub, 1802);
	assert.equal(state.bonds.groups[0].targetPercent, 100);
	assert.deepEqual(state.personalTotals, { deposits: 25, insurances: 30 });
	assert.deepEqual(state.data.crypto_groups, data.crypto_groups);
	assert.deepEqual(state.data.crypto_assets, data.crypto_assets);
	assert.deepEqual(state.data.profiles, data.profiles);
});

test('missing rates do not substitute demo prices or block other portfolio data', async () => {
	const state = await loadPortfolioState({ getAll: async () => data, currencyRate: async () => { throw new Error('offline'); } });
	assert.deepEqual(state.unavailableRates, ['USD']);
	assert.equal(state.bonds.totalBondsValueRub, null);
	assert.equal(state.stocks.total, 100);
});

test('an empty portfolio needs no currency calls, and getall failures propagate', async () => {
	const empty = { profiles: [], stocks: [], stock_groups: [], bonds: [], bond_groups: [], crypto_assets: [], crypto_groups: [], personal_assets: { deposits: [], insurances: [], credit_cards: [] } };
	const state = await loadPortfolioState({ getAll: async () => empty, currencyRate: async () => { throw new Error('unexpected'); } });
	assert.equal(state.stocks.total, 0);
	assert.equal(state.bonds.totalBondsValueRub, 0);
	assert.deepEqual(state.unavailableRates, []);
	await assert.rejects(loadPortfolioState({ getAll: async () => { throw new Error('offline'); }, currencyRate: async () => { throw new Error('unexpected'); } }), /offline/);
});
