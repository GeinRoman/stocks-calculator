import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const modules = new Map();
async function compileModule(name) {
	if (modules.has(name)) return modules.get(name);
	const source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
	let { outputText } = ts.transpileModule(source, {
		compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext }
	});
	for (const [, specifier, dependency] of [...outputText.matchAll(/['"]((?:#lib\/logic\/tabs\/|\.\/)([\w-]+)\.js)['"]/g)]) {
		outputText = outputText.replaceAll(`'${specifier}'`, JSON.stringify(await compileModule(dependency)))
			.replaceAll(`"${specifier}"`, JSON.stringify(await compileModule(dependency)));
	}
	const url = `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`;
	modules.set(name, url);
	return url;
}

const { createBondDraftGroups, validateBondDraft, addBondToGroup } = await import(await compileModule('bond-editor'));
const { computeBondSnapshot } = await import(await compileModule('bond-snapshot'));
const { sampleBonds, createEmptyBondPortfolio, createDemoBondPortfolios } = await import(await compileModule('demo-bonds'));
const { searchBonds, saveBondComposition, saveBondWeights } = await import(await compileModule('bond-service'));
const { demoPortfolios } = await import(await compileModule('demo-portfolio'));
const clock = new Date('2026-10-30T12:00:00Z');
const quote = {
	code: 'BOND', name: 'Облигация', price: 10, redemptionPrice: 1000,
	maturityDate: '2028-01-01', couponAmount: 999,
	couponSchedule: [
		{ date: '2025-01-01T12:00:00Z', amount: 7 },
		{ date: '2026-10-30T15:00:00+03:00', amount: 20 },
		{ date: '2026-10-30T12:00:00.001Z', amount: 3 },
		{ date: '2027-01-01T12:00:00Z', amount: 5 }
	]
};
const group = (id = 'a', bonds = [{ ...structuredClone(quote), quantity: 2 }]) => ({
	id, name: `Группа ${id}`, targetPercent: null, bonds
});
const snapshot = (groups = [group()]) => computeBondSnapshot(groups, clock);
const holding = (groups = [group()]) => snapshot(groups).groups[0].bonds[0];

function freeze(value) {
	if (value && typeof value === 'object') {
		Object.freeze(value);
		Object.values(value).forEach(freeze);
	}
	return value;
}

test('demo main matches the existing 250000 overview with 60/40 groups and an empty reserve', () => {
	const portfolios = createDemoBondPortfolios(clock);
	assert.equal(portfolios.main.totalBondsValue, 250000);
	assert.equal(portfolios.main.totalBondsValue, demoPortfolios.main.assetValues.bonds);
	assert.deepEqual(portfolios.main.groups.map(({ totalValue }) => totalValue), [150000, 100000]);
	assert.deepEqual(portfolios.main.groups.map(({ targetPercent }) => targetPercent), [60, 40]);
	assert.deepEqual(portfolios.main.groups.map(({ actualPercent }) => actualPercent), [60, 40]);
	assert.deepEqual(portfolios.main.groups.map(({ bonds }) => bonds.map(({ quantity }) => quantity)), [[100, 75], [60, 50, 10]]);
	assert.equal(portfolios.main.groups[1].bonds[2].totalValue, 0);
	assert.deepEqual(portfolios.reserve, { totalBondsValue: 0, groups: [], asOf: clock.toISOString() });
});

test('fixture factories and searches do not share schedules or mutable snapshot objects', async () => {
	const first = createDemoBondPortfolios(clock);
	first.main.groups[0].bonds[0].couponSchedule[0].amount = -1;
	first.reserve.groups.push(group());
	assert.equal(createDemoBondPortfolios(clock).main.groups[0].bonds[0].couponSchedule[0].amount, 40);
	assert.equal(sampleBonds[0].couponSchedule[0].amount, 40);
	assert.deepEqual(createEmptyBondPortfolio(clock).groups, []);
	assert.equal((await searchBonds('ofz'))[0].couponSchedule[0].amount, 40);
});

test('coupon cutoff is strictly after the timestamp including timezone offsets and milliseconds', () => {
	const bond = holding();
	assert.equal(bond.totalCoupons, 4);
	assert.equal(bond.remainingCoupons, 2);
	assert.equal(bond.remainingCouponValue, 8);
	assert.equal(bond.holdingRemainingCouponValue, 16);
	assert.equal(bond.couponSchedule.length, 4);
});

test('valuation excludes redemption principal and coupons and does not multiply couponAmount by count', () => {
	const saved = snapshot();
	assert.equal(saved.totalBondsValue, 20);
	assert.equal(saved.groups[0].totalValue, 20);
	assert.equal(saved.groups[0].bonds[0].totalValue, 20);
	assert.equal(saved.groups[0].bonds[0].remainingCouponValue, 8);
	assert.equal(saved.groups[0].bonds[0].actualPercent, 100);
});

test('unknown schedule preserves null counts and income without falling back to couponAmount', () => {
	const bond = holding([group('a', [{ ...quote, couponSchedule: null, quantity: 2 }])]);
	for (const key of ['totalCoupons', 'remainingCoupons', 'remainingCouponValue', 'holdingRemainingCouponValue']) assert.equal(bond[key], null);
});

test('unknown future amount keeps known counts but null income; unknown past amounts do not poison income', () => {
	const futureUnknown = structuredClone(quote);
	futureUnknown.couponSchedule[3].amount = null;
	const bond = holding([group('a', [{ ...futureUnknown, quantity: 2 }])]);
	assert.equal(bond.remainingCoupons, 2);
	assert.equal(bond.totalCoupons, 4);
	assert.equal(bond.remainingCouponValue, null);
	assert.equal(bond.holdingRemainingCouponValue, null);
	futureUnknown.couponSchedule[3].amount = 5;
	futureUnknown.couponSchedule[0].amount = null;
	assert.equal(holding([group('a', [{ ...futureUnknown, quantity: 2 }])]).remainingCouponValue, 8);
});

test('zero price, zero coupon, empty schedules and matured instruments are valid distinct states', () => {
	const bond = holding([group('a', [{ ...quote, price: 0, redemptionPrice: 0, couponAmount: 0,
		maturityDate: '2025-01-01', couponSchedule: [], quantity: 10 }])]);
	for (const key of ['price', 'redemptionPrice', 'couponAmount', 'totalValue', 'totalCoupons', 'remainingCoupons', 'remainingCouponValue', 'holdingRemainingCouponValue']) assert.equal(bond[key], 0);
	assert.equal(bond.actualPercent, null);
	const zeroFuture = holding([group('a', [{ ...quote, couponSchedule: [{ date: '2027-01-01T00:00:00Z', amount: 0 }], quantity: 2 }])]);
	assert.equal(zeroFuture.remainingCoupons, 1);
	assert.equal(zeroFuture.remainingCouponValue, 0);
	const matured = computeBondSnapshot([group()], new Date('2030-01-01T00:00:00Z')).groups[0].bonds[0];
	assert.equal(matured.totalCoupons, 4);
	assert.equal(matured.remainingCoupons, 0);
	assert.equal(matured.remainingCouponValue, 0);
});

test('maturity metadata never invents payments or overrides the authoritative demo schedule', () => {
	for (const maturityDate of [null, '2025-01-01', '2035-01-01']) {
		assert.equal(holding([group('a', [{ ...quote, maturityDate, quantity: 2 }])]).remainingCoupons, 2);
	}
});

test('draft editing and cancellation deeply isolate saved coupon schedules and holdings', () => {
	const saved = freeze(snapshot());
	const before = structuredClone(saved);
	const draft = createBondDraftGroups(saved);
	draft[0].bonds[0].couponSchedule[0].amount = 100;
	draft[0].bonds[0].quantity = 99;
	draft[0].name = 'Изменено';
	draft.splice(0, 1);
	assert.deepEqual(saved, before);
	assert.equal(createBondDraftGroups(saved)[0].bonds[0].quantity, 2);
});

test('draft creation preserves explicit zero targets and represents missing metadata as null', () => {
	const saved = snapshot([group(), group('b', [])]);
	saved.groups[0].targetPercent = 0;
	delete saved.groups[1].targetPercent;
	for (const field of ['redemptionPrice', 'maturityDate', 'couponAmount', 'couponSchedule']) delete saved.groups[0].bonds[0][field];
	const draft = createBondDraftGroups(saved);
	assert.deepEqual(draft.map(({ targetPercent }) => targetPercent), [0, null]);
	for (const field of ['redemptionPrice', 'maturityDate', 'couponAmount', 'couponSchedule']) assert.equal(draft[0].bonds[0][field], null);
});

test('same-group normalized additions merge quantities immutably and cross-group duplicates reject', () => {
	const groups = freeze([group(), group('b', [])]);
	const added = addBondToGroup(groups, 'a', { ...quote, code: ' bond ' }, 3);
	assert.equal(added[0].bonds.length, 1);
	assert.equal(added[0].bonds[0].quantity, 5);
	assert.equal(groups[0].bonds[0].quantity, 2);
	assert.notEqual(added[0].bonds[0].couponSchedule[0], groups[0].bonds[0].couponSchedule[0]);
	assert.throws(() => addBondToGroup(groups, 'b', { ...quote, code: ' bOnD ' }, 1), /другой группе/);
	assert.throws(() => addBondToGroup(groups, 'missing', quote, 1), /не найдена/);
	const newQuote = freeze({ ...structuredClone(quote), code: ' new ', name: '  Новая  ' });
	const inserted = addBondToGroup(groups, 'b', newQuote, 1)[1].bonds[0];
	assert.equal(inserted.code, 'NEW');
	assert.equal(inserted.name, 'Новая');
	assert.notEqual(inserted.couponSchedule[0], newQuote.couponSchedule[0]);
});

test('quantities must be positive safe integers at draft, add, merge and service boundaries', async () => {
	for (const quantity of [undefined, 0, -1, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
		const groups = [group('a', [{ ...quote, quantity }])];
		assert.match(validateBondDraft(groups), /Количество облигаций/);
		assert.throws(() => addBondToGroup([group('a', [])], 'a', quote, quantity), /Количество облигаций/);
		assert.throws(() => addBondToGroup(groups, 'a', quote, 1), /Количество облигаций/);
		await assert.rejects(saveBondComposition(snapshot(), groups, clock), /Количество облигаций/);
	}
	assert.throws(() => addBondToGroup([group('a', [{ ...quote, quantity: Number.MAX_SAFE_INTEGER }])], 'a', quote, 1), /Количество облигаций/);
});

test('duplicate normalized codes reject within and across bond groups', () => {
	const duplicate = { ...quote, code: ' bond ', quantity: 1 };
	assert.match(validateBondDraft([group('a', [{ ...quote, quantity: 1 }, duplicate])]), /не должна повторяться/);
	assert.match(validateBondDraft([group(), group('b', [duplicate])]), /не должна повторяться/);
});

test('bond group names are trimmed, case-insensitively unique and limited to 64 UTF-8 bytes', () => {
	for (const name of ['  ', 'a'.repeat(65), 'Я'.repeat(33)]) assert.match(validateBondDraft([{ ...group(), name }]), /Название группы/);
	assert.equal(validateBondDraft([{ ...group(), name: ' Я'.trim().repeat(32) }]), null);
	assert.match(validateBondDraft([{ ...group(), name: ' Банки ' }, { ...group('b', []), name: 'бАнКи' }]), /уникальными/);
	assert.match(validateBondDraft([group('a', []), { ...group('a', []), name: 'Другая' }]), /Идентификаторы/);
	assert.equal(validateBondDraft([{ ...group(), name: demoPortfolios.main.groups[0].name }]), null);
});

test('bad quote metadata rejects even when merging into an existing holding', async () => {
	for (const patch of [{ code: ' ' }, { name: ' ' }, { price: -1 }, { price: NaN }, { price: Infinity },
		{ redemptionPrice: -1 }, { redemptionPrice: Infinity }, { couponAmount: -1 }, { couponAmount: NaN }]) {
		const bond = { ...quote, ...patch };
		assert.ok(validateBondDraft([group('a', [{ ...bond, quantity: 1 }])]));
		assert.throws(() => addBondToGroup([group()], 'a', bond, 1));
		await assert.rejects(saveBondComposition(snapshot(), [group('a', [{ ...bond, quantity: 1 }])], clock));
	}
});

test('coupon schedules require real timezone-qualified ISO timestamps and valid nullable amounts', () => {
	for (const date of ['2027-01-01', '2027-01-01T12:00:00', '2027-02-30T12:00:00Z', '2027-13-01T00:00:00Z', '2027-01-01T24:00:00Z', 'bad']) {
		assert.match(validateBondDraft([group('a', [{ ...quote, couponSchedule: [{ date, amount: 0 }], quantity: 1 }])]), /Дата купона/);
	}
	for (const amount of [-1, NaN, Infinity, undefined]) {
		assert.match(validateBondDraft([group('a', [{ ...quote, couponSchedule: [{ date: '2027-01-01T00:00:00Z', amount }], quantity: 1 }])]), /Сумма купона/);
	}
	for (const maturityDate of ['bad', '2027-02-29', '2027-01-01T12:00:00']) assert.match(validateBondDraft([group('a', [{ ...quote, maturityDate, quantity: 1 }])]), /дату погашения/);
	assert.equal(validateBondDraft([group('a', [{ ...quote, maturityDate: '2028-02-29', quantity: 1 }])]), null);
});

test('overflow rejects holding values, aggregate values, coupon sums and holding income', async () => {
	const huge = { ...quote, price: Number.MAX_VALUE, couponSchedule: null, quantity: 2 };
	const cases = [
		[group('a', [huge])],
		[group('a', [{ ...huge, quantity: 1 }]), group('b', [{ ...huge, code: 'OTHER', quantity: 1 }])],
		[group('a', [{ ...quote, quantity: 2, couponSchedule: [{ date: '2027-01-01T00:00:00Z', amount: Number.MAX_VALUE }] }])],
		[group('a', [{ ...quote, quantity: 1, couponSchedule: [1, 2].map(() => ({ date: '2027-01-01T00:00:00Z', amount: Number.MAX_VALUE })) }])]
	];
	for (const groups of cases) {
		assert.match(validateBondDraft(groups), /слишком велика/);
		await assert.rejects(saveBondComposition(snapshot(), groups, clock), /слишком велика/);
	}
});

test('composition saves add/delete groups only on confirmation and isolate response schedules', async () => {
	const saved = freeze(createDemoBondPortfolios(clock).main);
	const before = structuredClone(saved);
	const draft = createBondDraftGroups(saved);
	draft.splice(1, 1);
	draft.push({ ...group('new'), name: '  Новая  ' });
	assert.deepEqual(saved, before);
	const updated = await saveBondComposition(saved, freeze(draft), clock);
	assert.deepEqual(updated.groups.map(({ id }) => id), ['government', 'new']);
	assert.equal(updated.groups[1].name, 'Новая');
	assert.equal(updated.totalBondsValue, 150020);
	assert.equal(updated.groups[1].bonds[0].targetPercent, null);
	updated.groups[0].bonds[0].couponSchedule[0].amount = 999;
	assert.deepEqual(saved, before);
	assert.equal(draft[0].bonds[0].couponSchedule[0].amount, 40);
	assert.deepEqual((await saveBondComposition(saved, [], clock)).groups, []);
});

test('weights allow incomplete drafts but saves require exact ids, initialized integer weights and exact 100', async () => {
	const saved = createDemoBondPortfolios(clock).main;
	assert.equal(validateBondDraft([group('a', []), { ...group('b', []), targetPercent: 0 }]), null);
	assert.match(validateBondDraft([{ ...group('a', []), targetPercent: 60 }, { ...group('b', []), targetPercent: 41 }]), /Сумма долей/);
	for (const weights of [[], [{ id: 'government', value: 100 }],
		[{ id: 'government', value: 100 }, { id: 'corporate', value: null }],
		[{ id: 'government', value: 60 }, { id: 'corporate', value: 30 }],
		[{ id: 'government', value: 60 }, { id: 'other', value: 40 }],
		[{ id: 'government', value: 50 }, { id: 'government', value: 50 }],
		[{ id: 'government', value: 60.5 }, { id: 'corporate', value: 39.5 }],
		[{ id: 'government', value: 100 }, { id: 'corporate', value: 0 }, { id: 'extra', value: 0 }]]) {
		await assert.rejects(saveBondWeights(saved, weights, clock));
	}
	await assert.rejects(saveBondWeights(createEmptyBondPortfolio(clock), [], clock), /100%/);
	const weights = freeze([{ id: 'corporate', value: 0 }, { id: 'government', value: 100 }]);
	const updated = await saveBondWeights(freeze(saved), weights, clock);
	assert.deepEqual(updated.groups.map(({ targetPercent }) => targetPercent), [100, 0]);
	assert.deepEqual(updated.groups.map(({ bonds }) => bonds.map(({ targetPercent }) => targetPercent)), [[50, 50], [0, 0, 0]]);
	assert.equal(updated.totalBondsValue, 250000);
	assert.equal(saved.groups[0].targetPercent, 60);
});

test('actual percentages use bonds only, target percentages are equal, and zero totals stay null', () => {
	const saved = snapshot([{ ...group(), targetPercent: 60 }, { ...group('b', [{ ...quote, code: 'OTHER', quantity: 6 }]), targetPercent: 40 }]);
	assert.deepEqual(saved.groups.map(({ actualPercent }) => actualPercent), [25, 75]);
	assert.deepEqual(saved.groups.map(({ bonds }) => bonds[0].targetPercent), [60, 40]);
	const zero = snapshot([{ ...group('a', [{ ...quote, price: 0, quantity: 1 }]), targetPercent: 0 }, group('b', [])]);
	assert.deepEqual(zero.groups.map(({ actualPercent }) => actualPercent), [null, null]);
	assert.equal(zero.groups[0].bonds[0].actualPercent, null);
	assert.equal(zero.groups[0].bonds[0].targetPercent, 0);
});

test('bond services never mutate or import stock holdings and asset snapshots', async () => {
	const before = structuredClone(demoPortfolios);
	freeze(demoPortfolios);
	const saved = createDemoBondPortfolios(clock).main;
	await saveBondComposition(saved, [group()], clock);
	await saveBondWeights(saved, [{ id: 'government', value: 60 }, { id: 'corporate', value: 40 }], clock);
	assert.deepEqual(demoPortfolios, before);
	assert.equal('assetValues' in saved, false);
	for (const name of ['bond-editor', 'bond-snapshot', 'bond-service', 'demo-bonds']) {
		const source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
		assert.doesNotMatch(source, /demo-portfolio|portfolio-service/);
	}
});

test('search matches Russian names and codes case-insensitively and deeply clones results', async () => {
	assert.equal((await searchBonds('  оФз  ')).length, 2);
	assert.equal((await searchBonds('corp-demo-01'))[0].code, 'CORP-DEMO-01');
	assert.deepEqual(await searchBonds('  '), []);
	assert.deepEqual(await searchBonds('missing-bond'), []);
	const first = await searchBonds('ofz-demo-01');
	first[0].couponSchedule[0].amount = -1;
	first[0].price = -1;
	const second = await searchBonds('OFZ-DEMO-01');
	assert.equal(second[0].couponSchedule[0].amount, 40);
	assert.equal(second[0].price, 750);
	assert.equal((await searchBonds('ZERO-DEMO'))[0].couponAmount, 0);
});

test('service saves revalidate draft metadata and duplicate ownership including weight-only saves', async () => {
	const saved = createDemoBondPortfolios(clock).main;
	const invalid = createBondDraftGroups(saved);
	invalid[1].bonds.push(structuredClone(invalid[0].bonds[0]));
	await assert.rejects(saveBondComposition(saved, invalid, clock), /не должна повторяться/);
	await assert.rejects(saveBondComposition(saved, [{ ...group(), name: ' ' }], clock), /Название группы/);
	const badSnapshot = structuredClone(saved);
	badSnapshot.groups[0].bonds[0].quantity = 0;
	await assert.rejects(saveBondWeights(badSnapshot, [{ id: 'government', value: 60 }, { id: 'corporate', value: 40 }], clock), /Количество облигаций/);
	const duplicateIds = structuredClone(saved);
	duplicateIds.groups[1].id = 'government';
	await assert.rejects(saveBondWeights(duplicateIds, [{ id: 'government', value: 60 }, { id: 'corporate', value: 40 }], clock), /точно соответствовать/);
});

test('snapshot asOf is valid ISO, supports clock refresh and rejects invalid clocks', async () => {
	assert.equal(snapshot().asOf, clock.toISOString());
	assert.ok(Number.isFinite(Date.parse(createEmptyBondPortfolio().asOf)));
	assert.throws(() => createEmptyBondPortfolio(new Date(NaN)), /дата оценки/);
	await assert.rejects(saveBondComposition(snapshot(), [group()], new Date(NaN)), /дата оценки/);
	const saved = await saveBondComposition(snapshot(), [group()], new Date('2030-01-01T00:00:00Z'));
	assert.equal(saved.groups[0].bonds[0].remainingCoupons, 0);
	assert.equal(saved.asOf, '2030-01-01T00:00:00.000Z');
});
