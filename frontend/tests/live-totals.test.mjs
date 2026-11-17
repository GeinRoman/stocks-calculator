import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

async function load(file) {
	const source = await readFile(new URL(`../src/lib/api/${file}.ts`, import.meta.url), 'utf8');
	const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
	return import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);
}
const { bondView } = await load('bond-view');
const { personalAssetTotals } = await load('personal-assets-view');
const { loadBondRates } = await load('bond-rates');
const bond = (code, clean_price, accrued_interest) => ({ code, name: code, isin: '', group_name: 'A', clean_price, accrued_interest, maturity_date: '', maturity_payment: null, lot_amount: 2, lot_size: 5 });
const data = { bond_groups: [{ name: 'A', weight: 100 }, { name: 'Empty', weight: 0 }], bonds: [bond('B', { currency: 'USD', price: 10 }, { currency: 'RUB', price: 5 })] };

test('bond values include accrued interest and lot size using provided rates', () => {
	const view = bondView(data, { USD: 90 }, '2026-01-01T00:00:00Z');
	assert.equal(view.totalBondsValueRub, 9050);
	assert.equal(view.groups[0].actualPercent, 100);
	assert.equal(view.groups[0].bonds[0].actualPercent, 100);
	assert.equal(view.groups[0].bonds[0].targetPercent, 100);
	assert.equal(view.groups[1].totalValueRub, 0);
});

test('missing exchange rates make totals and percentages unavailable', () => {
	const view = bondView(data, {}, '2026-01-01T00:00:00Z');
	assert.equal(view.totalBondsValueRub, null);
	assert.equal(view.groups[0].totalValueRub, null);
	assert.equal(view.groups[0].actualPercent, null);
	assert.equal(view.groups[0].bonds[0].actualPercent, null);
});

test('empty and zero-priced bond portfolios retain zero and unavailable percentages', () => {
	const empty = bondView({ bond_groups: [], bonds: [] }, {}, '2026-01-01T00:00:00Z');
	assert.equal(empty.totalBondsValueRub, 0);
	const zero = bondView({ bond_groups: [{ name: 'A', weight: 0 }], bonds: [bond('B', { currency: 'RUB', price: 0 }, { currency: 'RUB', price: 0 })] }, {}, '2026-01-01T00:00:00Z');
	assert.equal(zero.totalBondsValueRub, 0);
	assert.equal(zero.groups[0].actualPercent, null);
});

test('deposit and insurance totals are independent and reject invalid values', () => {
	assert.deepEqual(personalAssetTotals({ deposits: [{ amount_rub: 100 }, { amount_rub: 25 }], insurances: [{ amount_rub: 50 }] }), { deposits: 125, insurances: 50 });
	assert.deepEqual(personalAssetTotals({ deposits: [], insurances: [{ amount_rub: NaN }] }), { deposits: 0, insurances: null });
});

test('bond exchange rates are fetched once per required non-RUB currency', async () => {
	const calls = [];
	const bonds = [...data.bonds, bond('C', { currency: 'USD', price: 1 }, { currency: 'EUR', price: 1 })];
	const result = await loadBondRates(bonds, async (from, to) => {
		calls.push([from, to]);
		if (from === 'EUR') throw new Error('unavailable');
		return { from, to, rate: 90 };
	});
	assert.deepEqual(calls, [['USD', 'RUB'], ['EUR', 'RUB']]);
	assert.equal(result.rates.USD, 90);
	assert.deepEqual(result.unavailable, ['EUR']);
	assert.equal(bondView({ ...data, bonds }, result.rates, '2026-01-01T00:00:00Z').totalBondsValueRub, null);
});

test('RUB-only bonds do not request rates and invalid responses stay unavailable', async () => {
	let calls = 0;
	const onlyRub = [bond('R', { currency: 'RUB', price: 1 }, { currency: 'RUB', price: 0 })];
	assert.deepEqual((await loadBondRates(onlyRub, async () => { calls++; })).unavailable, []);
	assert.equal(calls, 0);
	const result = await loadBondRates(data.bonds, async () => ({ from: 'USD', to: 'RUB', rate: 0 }));
	assert.deepEqual(result.unavailable, ['USD']);
});
