import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const modules = new Map();
async function compile(name) {
	if (modules.has(name)) return modules.get(name);
	const source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
	let { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
	for (const [, specifier, dependency] of [...outputText.matchAll(/['"]((?:#lib\/logic\/tabs\/|\.\/)([\w-]+)\.js)['"]/g)]) {
		outputText = outputText.replaceAll(`'${specifier}'`, JSON.stringify(await compile(dependency))).replaceAll(`"${specifier}"`, JSON.stringify(await compile(dependency)));
	}
	const url = `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`;
	modules.set(name, url);
	return url;
}
const { createBondDraftGroups, validateBondDraft, addBondToGroup } = await import(await compile('bond-editor'));
const { computeBondSnapshot } = await import(await compile('bond-snapshot'));
const { createDemoBondPortfolios, createEmptyBondPortfolio, sampleBonds } = await import(await compile('demo-bonds'));
const { searchBonds, saveBondComposition, saveBondWeights } = await import(await compile('bond-service'));
const { demoPortfolios } = await import(await compile('demo-portfolio'));
const clock = new Date('2026-10-30T12:00:00Z');
const quote = { code: 'BOND', name: 'Bond', isin: 'TEST00000001', cleanPrice: { currency: 'USD', price: 10 }, accruedInterest: { currency: 'RUB', price: 90 }, maturityDate: '2028-02-29', maturityPayment: null, lotSize: 2 };
const group = (id = 'a', bonds = [{ ...structuredClone(quote), lotCount: 2 }]) => ({ id, name: `Group ${id}`, targetPercent: 50, bonds });

test('snapshot uses dirty RUB price and lots; targets independent of valuation', () => {
	const result = computeBondSnapshot([group(), { ...group('b', [{ ...quote, code: 'OTHER', lotCount: 1 }]), targetPercent: 50 }], clock);
	assert.equal(result.totalBondsValueRub, 5940);
	assert.deepEqual(result.groups.map(g => g.totalValueRub), [3960, 1980]);
	assert.ok(Math.abs(result.groups[0].actualPercent - 200 / 3) < 1e-10);
	assert.ok(Math.abs(result.groups[1].actualPercent - 100 / 3) < 1e-10);
	assert.equal(result.groups[0].bonds[0].targetPercent, 50);
	assert.equal(result.asOf, clock.toISOString());
});
test('unavailable rate never produces partial total or actual percentages', () => {
	const result = computeBondSnapshot([group(), group('b', [{ ...quote, code: 'OTHER', cleanPrice: { currency: 'XYZ', price: 1 }, accruedInterest: { currency: 'XYZ', price: 0 }, lotCount: 1 }])], clock);
	assert.equal(result.totalBondsValueRub, null);
	assert.equal(result.groups[0].totalValueRub, 3960);
	assert.equal(result.groups[1].totalValueRub, null);
	assert.deepEqual(result.groups.map(g => g.actualPercent), [null, null]);
	assert.equal(result.groups[0].bonds[0].actualPercent, null);
	assert.equal(result.groups[1].bonds[0].targetPercent, 50);
});
test('zero values and empty portfolios remain valid', () => {
	const result = computeBondSnapshot([group('a', [{ ...quote, cleanPrice: { currency: 'RUB', price: 0 }, accruedInterest: { currency: 'RUB', price: 0 }, lotCount: 1 }])], clock);
	assert.equal(result.totalBondsValueRub, 0);
	assert.equal(result.groups[0].actualPercent, null);
	assert.deepEqual(createEmptyBondPortfolio(clock), { totalBondsValueRub: 0, groups: [], asOf: clock.toISOString() });
});
test('validates safe integer lots, metadata, group names and duplicates', () => {
	for (const lotCount of [undefined, 0, -1, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.ok(validateBondDraft([group('a', [{ ...quote, lotCount }])]));
	for (const lotSize of [0, -1, 1.5, NaN, Number.MAX_SAFE_INTEGER + 1]) assert.ok(validateBondDraft([group('a', [{ ...quote, lotSize, lotCount: 1 }])]));
	for (const patch of [{ isin: null }, { maturityDate: '2028-02-30' }, { maturityPayment: undefined }, { cleanPrice: { currency: '', price: 0 } }, { accruedInterest: { currency: 'RUB', price: -1 } }]) assert.ok(validateBondDraft([group('a', [{ ...quote, ...patch, lotCount: 1 }])]));
	assert.equal(validateBondDraft([group('a', [{ ...quote, isin: '', lotCount: 1 }])]), null);
	assert.match(validateBondDraft([group('a', [{ ...quote, lotCount: Number.MAX_SAFE_INTEGER }])]), /Количество облигаций слишком велико/);
	assert.ok(validateBondDraft([group(), { ...group('b', []), name: ' group A ' }]));
	assert.ok(validateBondDraft([group(), group('b', [{ ...quote, code: ' bond ', lotCount: 1 }])]));
});
test('editing merges lots immutably and rejects duplicates across groups', async () => {
	const saved = createDemoBondPortfolios(clock).main, before = structuredClone(saved);
	const draft = createBondDraftGroups(saved);
	const updated = addBondToGroup(draft, 'government', { ...sampleBonds[0], code: ' ofz-demo-01 ' }, 2);
	assert.equal(updated[0].bonds[0].lotCount, 102);
	assert.equal(draft[0].bonds[0].lotCount, 100);
	assert.throws(() => addBondToGroup(draft, 'corporate', sampleBonds[0], 1), /другой группе/);
	updated[0].bonds[0].cleanPrice.price = 1;
	assert.deepEqual(saved, before);
	const response = await saveBondComposition(saved, draft, clock);
	response.groups[0].bonds[0].cleanPrice.price = 2;
	assert.deepEqual(saved, before);
	assert.notEqual(draft[0].bonds[0].cleanPrice.price, 2);
});
test('weights require exact group ids and preserve valuation', async () => {
	const saved = createDemoBondPortfolios(clock).main;
	await assert.rejects(saveBondWeights(saved, [{ id: 'government', value: 100 }], clock));
	const result = await saveBondWeights(saved, [{ id: 'government', value: 100 }, { id: 'corporate', value: 0 }], clock);
	assert.equal(result.totalBondsValueRub, saved.totalBondsValueRub);
	assert.deepEqual(result.groups.map(g => g.targetPercent), [100, 0]);
});
test('search clones metadata, stock fixtures remain independent', async () => {
	const before = structuredClone(demoPortfolios);
	assert.equal((await searchBonds(' оФз ')).length, 2);
	assert.equal((await searchBonds('corp-demo-01'))[0].code, 'CORP-DEMO-01');
	assert.deepEqual(await searchBonds(' '), []);
	const first = await searchBonds('OFZ-DEMO-01');
	first[0].cleanPrice.price = -1;
	assert.equal((await searchBonds('ofz-demo-01'))[0].cleanPrice.price, 750);
	assert.deepEqual(demoPortfolios, before);
	assert.equal('coupons' in first[0], false);
});
