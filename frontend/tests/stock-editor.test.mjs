import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

async function compileModule(name, imports = {}) {
	const source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
	let { outputText } = ts.transpileModule(source, {
		compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext }
	});
	for (const [dependency, url] of Object.entries(imports)) {
		outputText = outputText.replaceAll(`'#lib/logic/tabs/${dependency}.js'`, JSON.stringify(url))
			.replaceAll(`'./${dependency}.js'`, JSON.stringify(url));
	}
	return `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`;
}

const editorUrl = await compileModule('stock-editor');
const { createDraftGroups, validateDraft, addStockToGroup, validateWeights } = await import(editorUrl);
const { searchStocks, saveComposition, saveWeights } = await import(
	await compileModule('portfolio-service', { 'stock-editor': editorUrl })
);
const { demoPortfolios, emptyPortfolio } = await import(await compileModule('demo-portfolio'));
const sber = { code: 'SBER', name: 'Сбербанк', price: 300, lotSize: 10 };
const newGroup = (id = 'new') => ({ id, name: `Новая ${id}`, targetPercent: null, stocks: [] });

function freeze(value) {
	if (value && typeof value === 'object') {
		Object.freeze(value);
		Object.values(value).forEach(freeze);
	}
	return value;
}

freeze(demoPortfolios);
freeze(emptyPortfolio);

test('drafts are deeply isolated and discarding edits leaves the fixture unchanged', () => {
	const original = structuredClone(demoPortfolios.main);
	const draft = createDraftGroups(demoPortfolios.main);
	draft[0].name = 'Изменено';
	draft[0].stocks[0].lotCount = 99;
	draft[0].stocks.push({ ...sber, code: 'OTHER', lotCount: 2 });
	draft.splice(1, 1);
	assert.deepEqual(demoPortfolios.main, original);
	assert.deepEqual(createDraftGroups(demoPortfolios.main)[0].stocks[0].lotCount, 10);
});

test('draft creation preserves zero and normalizes missing targets to null', () => {
	const snapshot = structuredClone(demoPortfolios.main);
	snapshot.groups[0].targetPercent = 0;
	delete snapshot.groups[1].targetPercent;
	snapshot.groups[2].targetPercent = null;
	assert.deepEqual(createDraftGroups(snapshot).map((group) => group.targetPercent), [0, null, null]);
});

test('lots must be positive safe integers at validation, addition, and service boundaries', async () => {
	for (const lotCount of [undefined, 0, -1, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
		const groups = [{ ...newGroup(), stocks: [{ ...sber, lotCount }] }];
		assert.match(validateDraft(groups), /Количество лотов/);
		assert.throws(() => addStockToGroup([newGroup()], 'new', sber, lotCount), /Количество лотов/);
		assert.throws(() => addStockToGroup(groups, 'new', sber, 2), /Количество лотов/);
		await assert.rejects(saveComposition(emptyPortfolio, groups), /Количество лотов/);
	}
	assert.equal(validateDraft([{ ...newGroup(), stocks: [{ ...sber, lotCount: 1 }] }]), null);
});

test('same-group additions merge normalized codes immutably; cross-group ownership rejects', () => {
	const groups = freeze(createDraftGroups(demoPortfolios.main));
	const added = addStockToGroup(groups, 'banks', { ...sber, code: ' sber ' }, 2);
	assert.equal(added[0].stocks.length, 2);
	assert.equal(added[0].stocks[0].lotCount, 12);
	assert.equal(groups[0].stocks[0].lotCount, 10);
	assert.notEqual(added[1].stocks[0], groups[1].stocks[0]);
	assert.throws(() => addStockToGroup(groups, 'energy', { ...sber, code: 'sBeR' }, 2), /другой группе/);
	assert.throws(() => addStockToGroup(groups, 'missing', sber, 1), /не найдена/);
	const huge = [{ ...newGroup(), stocks: [{ ...sber, lotCount: Number.MAX_SAFE_INTEGER }] }];
	assert.throws(() => addStockToGroup(huge, 'new', sber, 1), /Количество лотов/);
});

test('new holdings normalize their code and clone quote data', () => {
	const quote = freeze({ ...sber, code: ' sber ' });
	const groups = freeze([newGroup()]);
	const added = addStockToGroup(groups, 'new', quote, 3);
	assert.deepEqual(added[0].stocks, [{ ...sber, lotCount: 3 }]);
	assert.deepEqual(groups[0].stocks, []);
});

test('draft duplicate codes are rejected both within and across groups', () => {
	const stock = { ...sber, lotCount: 1 };
	const duplicate = { ...stock, code: ' sber ' };
	assert.match(validateDraft([{ ...newGroup(), stocks: [stock, duplicate] }]), /не должна повторяться/);
	assert.match(validateDraft([
		{ ...newGroup(), stocks: [stock] }, { ...newGroup('second'), stocks: [duplicate] }
	]), /не должна повторяться/);
});

test('group names are trimmed, case-insensitively unique, and limited to 64 UTF-8 bytes', () => {
	assert.equal(validateDraft([{ ...newGroup(), name: '  Банки  ' }]), null);
	assert.equal(validateDraft([{ ...newGroup(), name: 'Я'.repeat(32) }]), null);
	for (const name of ['  ', 'a'.repeat(65), 'Я'.repeat(33)]) {
		assert.match(validateDraft([{ ...newGroup(), name }]), /Название группы/);
	}
	assert.match(validateDraft([
		{ ...newGroup(), name: 'Банки' }, { ...newGroup('other'), name: '  бАнКи ' }
	]), /уникальными/);
	assert.match(validateDraft([newGroup(), { ...newGroup(), name: 'Другое' }]), /Идентификаторы/);
});

test('invalid quote metadata is rejected, including during same-group merges', async () => {
	for (const patch of [{ price: NaN }, { price: 0 }, { price: -1 }, { price: Infinity },
		{ lotSize: 0 }, { lotSize: 1.5 }, { lotSize: NaN }, { lotSize: Infinity },
		{ code: ' ' }, { name: ' ' }]) {
		const stock = { ...sber, ...patch };
		assert.ok(validateDraft([{ ...newGroup(), stocks: [{ ...stock, lotCount: 1 }] }]));
		assert.throws(() => addStockToGroup(createDraftGroups(demoPortfolios.main), 'banks', stock, 1));
		await assert.rejects(saveComposition(emptyPortfolio, [{ ...newGroup(), stocks: [{ ...stock, lotCount: 1 }] }]));
	}
});

test('weights distinguish null, zero, and integers and allow totals up to 100', () => {
	assert.equal(validateWeights([]), null);
	assert.equal(validateWeights([{ id: 'a', value: null }, { id: 'b', value: 0 }]), null);
	assert.equal(validateWeights([{ id: 'a', value: 40 }, { id: 'b', value: 60 }]), null);
	for (const value of [undefined, -1, 101, 0.5, NaN, Infinity]) {
		assert.match(validateWeights([{ id: 'a', value }]), /целым числом/);
	}
	assert.match(validateWeights([{ id: 'a', value: 41 }, { id: 'b', value: 60 }]), /Сумма долей/);
	assert.match(validateWeights([{ id: 'a', value: 0 }, { id: 'a', value: null }]), /Идентификаторы/);
	assert.match(validateWeights([{ id: '', value: 0 }]), /Идентификаторы/);
	const groups = createDraftGroups(demoPortfolios.main);
	groups[0].targetPercent = 31;
	assert.match(validateDraft(groups), /Сумма долей/);
});

test('weight submissions require exactly 100% and explicit values for every group', () => {
	for (const weights of [[], [{ id: 'a', value: 60 }, { id: 'b', value: 0 }],
		[{ id: 'a', value: 100 }, { id: 'b', value: null }]]) {
		assert.equal(validateWeights(weights), null, 'incomplete drafts can be edited');
		assert.match(validateWeights(weights, true), /100%/);
	}
	assert.equal(validateWeights([{ id: 'a', value: 100 }, { id: 'b', value: 0 }], true), null);
	assert.equal(validateWeights([{ id: 'a', value: 30 }, { id: 'b', value: 30 }, { id: 'c', value: 40 }], true), null);
});

test('weight saves reject the unassigned remainder without redistributing it', async () => {
	const weights = [
		{ id: 'banks', value: 30 }, { id: 'energy', value: 30 }, { id: 'technology', value: 40 }
	];
	const snapshot = await saveWeights(demoPortfolios.main, weights);
	weights[2].value = 0;
	await assert.rejects(saveWeights(snapshot, weights), /100%/);
	assert.deepEqual(weights.map(({ value }) => value), [30, 30, 0]);
	assert.deepEqual(snapshot.groups.map(({ targetPercent }) => targetPercent), [30, 30, 40]);
	weights[0].value = 50;
	weights[1].value = 50;
	const saved = await saveWeights(snapshot, weights);
	assert.deepEqual(saved.groups.map(({ targetPercent }) => targetPercent), [50, 50, 0]);
});

test('search matches Russian names and codes, returns 14 Россети samples, and never shares quotes', async () => {
	const rosseti = await searchStocks('  рОсСеТи  ');
	assert.equal(rosseti.length, 14);
	assert.equal(new Set(rosseti.map(({ code }) => code)).size, 14);
	for (const stock of rosseti) {
		assert.ok(Number.isFinite(stock.price) && stock.price > 0);
		assert.ok(Number.isSafeInteger(stock.lotSize) && stock.lotSize > 0);
	}
	assert.equal((await searchStocks('сБеР'))[0].code, 'SBER');
	for (const code of ['SBER', 'VTBR', 'LKOH', 'GAZP', 'YDEX', 'POSI']) {
		assert.ok((await searchStocks(code.toLowerCase())).some((stock) => stock.code === code));
	}
	assert.deepEqual(await searchStocks('no-such-stock'), []);
	assert.deepEqual(await searchStocks('  '), []);
	const first = await searchStocks('SBER');
	first[0].price = -1;
	assert.equal((await searchStocks('SBER'))[0].price, 300);
});

test('composition saves support group addition/deletion, normalize names, and preserve non-stock assets', async () => {
	const original = structuredClone(demoPortfolios.main);
	const draft = createDraftGroups(demoPortfolios.main);
	draft.splice(1, 1);
	draft.push({ ...newGroup(), name: '  Новая группа  ', stocks: [{ code: ' fees ', name: 'Россети', price: 0.08, lotSize: 10000, lotCount: 2 }] });
	draft[0].stocks[0].lotCount = 20;
	const updated = await saveComposition(demoPortfolios.main, freeze(draft));
	assert.deepEqual(updated.groups.map(({ id }) => id), ['banks', 'technology', 'new']);
	assert.equal(updated.groups[2].name, 'Новая группа');
	assert.equal(updated.groups[2].targetPercent, null);
	assert.equal(updated.groups[2].stocks[0].targetPercent, null);
	assert.equal(updated.groups[2].stocks[0].code, 'FEES');
	assert.equal(updated.groups[2].totalValue, 1600);
	assert.equal(updated.groups[0].totalValue, 70000);
	assert.equal(updated.totalStocksValue, 96600);
	assert.deepEqual(updated.assetValues, { stocks: 96600, bonds: 250000, deposits: 150000 });
	assert.equal(updated.groups[0].actualPercent, 70000 / 96600 * 100);
	assert.equal(updated.groups[0].stocks[0].actualPercent, 60000 / 96600 * 100);
	assert.equal(updated.groups[0].stocks[0].targetPercent, 15);
	assert.deepEqual(demoPortfolios.main, original);
	updated.groups[0].stocks[0].lotCount = 999;
	updated.assetValues.bonds = 0;
	assert.equal(draft[0].stocks[0].lotCount, 20);
	assert.deepEqual(demoPortfolios.main, original);
	const unchanged = await saveComposition(demoPortfolios.main, createDraftGroups(demoPortfolios.main));
	assert.equal(unchanged.totalStocksValue, original.totalStocksValue);
	for (let i = 0; i < unchanged.groups.length; i++) {
		assert.equal(unchanged.groups[i].totalValue, original.groups[i].totalValue);
		for (let j = 0; j < unchanged.groups[i].stocks.length; j++) {
			assert.ok(Math.abs(unchanged.groups[i].stocks[j].actualPercent - original.groups[i].stocks[j].actualPercent) < 1e-10);
		}
	}
});

test('empty composition yields zero valuation and null actual percentages, preserving null and zero targets', async () => {
	const groups = [{ ...newGroup('unset') }, { ...newGroup('zero'), targetPercent: 0 }];
	const updated = await saveComposition(demoPortfolios.main, freeze(groups));
	assert.equal(updated.totalStocksValue, 0);
	assert.equal(updated.assetValues.stocks, 0);
	assert.deepEqual(updated.groups.map(({ actualPercent }) => actualPercent), [null, null]);
	assert.deepEqual(updated.groups.map(({ targetPercent }) => targetPercent), [null, 0]);
	const deleted = await saveComposition(demoPortfolios.main, []);
	assert.deepEqual(deleted.groups, []);
	assert.equal(deleted.assetValues.bonds, 250000);
});

test('weight saves match ids exactly, accept any order, recalculate equal targets, and remain isolated', async () => {
	const weights = freeze([{ id: 'technology', value: 0 }, { id: 'banks', value: 0 }, { id: 'energy', value: 100 }]);
	const updated = await saveWeights(demoPortfolios.main, weights);
	assert.deepEqual(updated.groups.map(({ targetPercent }) => targetPercent), [0, 100, 0]);
	assert.deepEqual(updated.groups.map(({ stocks }) => stocks.map(({ targetPercent }) => targetPercent)), [[0, 0], [50, 50], [0, 0]]);
	assert.equal(updated.totalStocksValue, 100000);
	assert.equal(updated.groups[0].actualPercent, 40);
	assert.equal(demoPortfolios.main.groups[0].targetPercent, 30);
	updated.groups[0].stocks[0].price = 1;
	assert.equal(demoPortfolios.main.groups[0].stocks[0].price, 300);
	for (const invalid of [[], weights.slice(1), [...weights, { id: 'extra', value: null }],
		weights.map((weight) => weight.id === 'banks' ? { ...weight, id: 'unknown' } : weight),
		weights.map((weight) => weight.id === 'banks' ? { ...weight, value: 1 } : weight),
		[...weights.slice(1), weights[1]]]) {
		await assert.rejects(saveWeights(demoPortfolios.main, invalid), /групп|долей/);
	}
	await assert.rejects(saveWeights(emptyPortfolio, []), /100%/);
	await assert.rejects(saveWeights(demoPortfolios.main, [
		{ id: 'banks', value: 0 }, { id: 'energy', value: 100 }, { id: 'technology', value: null }
	]), /100%/);
});

test('services revalidate names, duplicate ownership, and finite aggregate values', async () => {
	const duplicate = createDraftGroups(demoPortfolios.main);
	duplicate[1].stocks.push({ ...duplicate[0].stocks[0] });
	await assert.rejects(saveComposition(demoPortfolios.main, duplicate), /не должна повторяться/);
	const invalidName = [{ ...newGroup(), name: ' ' }];
	await assert.rejects(saveComposition(emptyPortfolio, invalidName), /Название группы/);
	const huge = [{ ...newGroup(), stocks: [
		{ ...sber, price: Number.MAX_VALUE, lotSize: 1, lotCount: 1 },
		{ ...sber, code: 'OTHER', price: Number.MAX_VALUE, lotSize: 1, lotCount: 1 }
	] }];
	await assert.rejects(saveComposition(emptyPortfolio, huge), /слишком велика/);
});
