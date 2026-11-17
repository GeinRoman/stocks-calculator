import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

async function moduleUrl(name, dependencies = {}) {
	let source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
	for (const [path, url] of Object.entries(dependencies)) source = source.replaceAll(`'./${path}.js'`, `'${url}'`);
	const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
	return `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`;
}
const dateUrl = await moduleUrl('personal-asset-date');
const assetsUrl = await moduleUrl('personal-assets');
const logicUrl = await moduleUrl('credit-cards', { 'personal-asset-date': dateUrl, 'personal-assets': assetsUrl });
const demoUrl = await moduleUrl('credit-card-demo', { 'personal-assets': assetsUrl, 'credit-cards': logicUrl });
const logic = await import(logicUrl);
const { createDemoCreditCards } = await import(demoUrl);
const service = await import(await moduleUrl('credit-card-service', { 'credit-cards': logicUrl, 'credit-card-demo': demoUrl }));
const input = { name: ' Карта ', bankName: ' Банк ', paymentDate: '2024-02-29', paymentAmountRub: 0, overallLimitRub: 0 };
const card = { id: 1, ...input };

test('validates names, strict Gregorian dates and finite nonnegative amounts', () => {
	for (const date of ['0001-01-01', '9999-12-31', '2000-02-29', '2024-02-29', '1900-02-28']) {
		assert.equal(logic.validateCreditCard({ ...input, paymentDate: date, paymentAmountRub: Number.MAX_VALUE }), null);
	}
	for (const change of [
		{ name: ' ' }, { bankName: '' }, { name: null }, { paymentDate: null }, { paymentDate: '0000-01-01' },
		{ paymentDate: '10000-01-01' }, { paymentDate: '1900-02-29' }, { paymentDate: '2023-02-29' },
		{ paymentDate: '2024-02-30' }, { paymentDate: '2024-2-01' }, { paymentDate: '2024-13-01' },
		{ paymentAmountRub: -1 }, { paymentAmountRub: NaN }, { paymentAmountRub: Infinity },
		{ overallLimitRub: -1 }, { overallLimitRub: -Infinity }, { overallLimitRub: '0' }
	]) assert.ok(logic.validateCreditCard({ ...input, ...change }), JSON.stringify(change));
	assert.equal(logic.validateCreditCard({ ...input, paymentAmountRub: 100, overallLimitRub: 1 }), null);
});

test('demo dates cross leap and year boundaries, sort immutably and reuse statuses', () => {
	const demo = createDemoCreditCards(new Date(2024, 1, 29, 23));
	assert.deepEqual(demo.reserve, { cards: [] });
	assert.deepEqual(demo.main.cards.map(c => c.paymentDate), ['2024-03-14', '2024-02-26', '2024-03-15', '2024-02-29', '2024-03-07']);
	const original = structuredClone(demo.main.cards);
	const sorted = logic.sortCreditCards(demo.main.cards);
	assert.deepEqual(demo.main.cards, original);
	assert.deepEqual(sorted.map(c => logic.getCreditCardStatus(c.paymentDate, '2024-02-29')), ['expired', 'ending', 'ending', 'ending', 'normal']);
	assert.deepEqual(logic.sortCreditCards([{ ...card, id: 5 }, { ...card, id: 2 }, { ...card, id: 3 }]).map(c => c.id), [2, 3, 5]);
	assert.deepEqual(createDemoCreditCards(new Date(2023, 11, 30)).main.cards.map(c => c.paymentDate), ['2024-01-13', '2023-12-27', '2024-01-14', '2023-12-30', '2024-01-06']);
});

test('profile loads are independent and do not expose prototype properties', async () => {
	const first = await service.getCreditCards('main');
	const second = await service.getCreditCards('main');
	first.cards[0].name = 'changed';
	assert.notEqual(second.cards[0].name, 'changed');
	for (const profile of ['reserve', 'other', '__proto__', 'constructor']) assert.deepEqual(await service.getCreditCards(profile), logic.emptyCreditCards());
	assert.notEqual(logic.emptyCreditCards().cards, logic.emptyCreditCards().cards);
});

test('CRUD validates inputs, preserves IDs and isolates every snapshot', async () => {
	const empty = logic.emptyCreditCards();
	const first = await service.createCreditCard(empty, input);
	const second = await service.createCreditCard(first, { ...input, name: 'Другая' });
	assert.deepEqual(empty, { cards: [] });
	assert.deepEqual(second.cards.map(c => c.id), [1, 2]);
	assert.equal(first.cards[0].name, 'Карта');
	assert.equal(first.cards[0].bankName, 'Банк');
	const changed = await service.updateCreditCard(second, 1, { ...input, id: 42, name: ' Новая ', paymentAmountRub: 42 });
	assert.equal(changed.cards[0].id, 1);
	assert.equal(changed.cards[0].name, 'Новая');
	assert.equal(second.cards[0].paymentAmountRub, 0);
	assert.deepEqual(changed.cards[1], second.cards[1]);
	changed.cards[1].name = 'independent';
	assert.equal(second.cards[1].name, 'Другая');
	const deleted = await service.deleteCreditCard(second, 1);
	assert.equal(deleted.cards.length, 1);
	deleted.cards[0].name = 'independent';
	assert.equal(second.cards[1].name, 'Другая');
	assert.equal((await service.createCreditCard(deleted, input)).cards[1].id, 3);
	for (const change of [{ name: ' ' }, { paymentDate: 'bad' }, { paymentAmountRub: NaN }, { overallLimitRub: -1 }]) {
		await assert.rejects(service.createCreditCard(second, { ...input, ...change }));
		await assert.rejects(service.updateCreditCard(second, 1, { ...input, ...change }));
	}
	await assert.rejects(service.updateCreditCard(second, 99, input), /не найдена/);
	await assert.rejects(service.deleteCreditCard(second, 99), /не найдена/);
});

test('rejects tampered and duplicate IDs and exhausted ID space', async () => {
	for (const cards of [[card, { ...card }], [{ ...card, id: 0 }], [{ ...card, id: -1 }], [{ ...card, id: 1.5 }], [{ ...card, id: NaN }], [{ ...card, id: Number.MAX_SAFE_INTEGER + 1 }]]) {
		for (const operation of [service.createCreditCard({ cards }, input), service.updateCreditCard({ cards }, 1, input), service.deleteCreditCard({ cards }, 1)]) {
			await assert.rejects(operation, /идентификатор/);
		}
	}
	await assert.rejects(service.createCreditCard({ cards: [{ ...card, id: Number.MAX_SAFE_INTEGER }] }, input), /идентификатор/);
});
