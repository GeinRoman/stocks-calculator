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
const logicUrl = await moduleUrl('personal-assets');
const demoUrl = await moduleUrl('personal-assets-demo', { 'personal-assets': logicUrl });
const logic = await import(logicUrl);
const { createDemoPersonalAssets } = await import(demoUrl);
const service = await import(await moduleUrl('personal-assets-service', { 'personal-assets': logicUrl, 'personal-assets-demo': demoUrl }));
const input = { name: ' Вклад ', provider: ' Банк ', amountRub: 1000, annualInterest: 5, startDate: '2024-01-01', endDate: '2024-12-31' };
const asset = { id: 1, type: 'deposit', ...input };

test('validation requires real ISO dates, positive principal and nonnegative finite interest', () => {
	assert.equal(logic.validatePersonalAsset(input), null);
	for (const change of [
		{ name: '  ' }, { provider: '' }, { amountRub: 0 }, { amountRub: -1 }, { amountRub: Infinity },
		{ annualInterest: -1 }, { annualInterest: NaN }, { startDate: '0000-01-01' },
		{ startDate: '2023-02-29' }, { startDate: '2024-02-30' }, { startDate: '2024-2-01' },
		{ startDate: '2024-13-01' }, { startDate: '2025-01-01' }, { endDate: '10000-01-01' }
	]) assert.ok(logic.validatePersonalAsset({ ...input, ...change }), JSON.stringify(change));
	assert.equal(logic.validatePersonalAsset({ ...input, annualInterest: 0, startDate: '0001-01-01', endDate: '9999-12-31' }), null);
	assert.equal(logic.validatePersonalAsset({ ...input, startDate: '2024-02-29', endDate: '2024-02-29' }), null);
});

test('status uses calendar days including leap days and DST boundaries', () => {
	for (const [end, expected] of [['2024-02-28', 'expired'], ['2024-02-29', 'ending'], ['2024-03-14', 'ending'], ['2024-03-15', 'normal']]) {
		assert.equal(logic.getPersonalAssetStatus(end, '2024-02-29'), expected);
	}
	assert.equal(logic.getPersonalAssetStatus('2024-03-24', '2024-03-10'), 'ending');
	assert.equal(logic.getPersonalAssetStatus('2024-03-25', '2024-03-10'), 'normal');
	assert.throws(() => logic.getPersonalAssetStatus('2023-02-29', '2023-01-01'));
});

test('simple interest estimates the whole term using actual days over 365', () => {
	assert.equal(logic.estimatedMaturityAmount(asset), 1050);
	assert.equal(logic.estimatedMaturityAmount({ ...asset, annualInterest: 0 }), 1000);
	assert.equal(logic.estimatedMaturityAmount({ ...asset, endDate: asset.startDate }), 1000);
	assert.equal(logic.estimatedMaturityAmount({ ...asset, endDate: '2025-01-01' }), 1000 * (1 + 0.05 * 366 / 365));
	for (const change of [{ startDate: '' }, { startDate: undefined }, { endDate: '2023-01-01' }, { annualInterest: Infinity }, { amountRub: Infinity }, { amountRub: Number.MAX_VALUE, annualInterest: Number.MAX_VALUE }]) {
		assert.equal(logic.estimatedMaturityAmount({ ...asset, ...change }), null);
	}
});

test('sample dates are local, dynamic, and independently allocated', async () => {
	assert.equal(logic.localCalendarDate(new Date(2024, 1, 29, 23)), '2024-02-29');
	const demo = createDemoPersonalAssets(new Date(2024, 1, 29, 23));
	assert.deepEqual([...demo.main.deposits, ...demo.main.insurances].map(a => a.endDate), ['2024-02-28', '2024-03-07', '2024-03-14', '2024-03-15']);
	assert.deepEqual([...demo.main.deposits, ...demo.main.insurances].map(a => logic.getPersonalAssetStatus(a.endDate, '2024-02-29')), ['expired', 'ending', 'ending', 'normal']);
	const first = await service.getPersonalAssets('main');
	const second = await service.getPersonalAssets('main');
	first.deposits[0].name = 'changed';
	assert.notEqual(second.deposits[0].name, 'changed');
	assert.deepEqual(await service.getPersonalAssets('other'), logic.emptyPersonalAssets());
	assert.deepEqual(await service.getPersonalAssets('__proto__'), logic.emptyPersonalAssets());
});

test('create and update are immutable, scoped by type, and cannot change principal', async () => {
	const original = logic.emptyPersonalAssets();
	const first = await service.createPersonalAsset(original, 'deposit', input);
	const both = await service.createPersonalAsset(first, 'insurance', input);
	const next = await service.createPersonalAsset(both, 'deposit', input);
	assert.deepEqual(original, { deposits: [], insurances: [] });
	assert.deepEqual(next.deposits.map(a => a.id), [1, 2]);
	assert.equal(next.insurances[0].id, 1);
	assert.equal(next.deposits[0].name, 'Вклад');
	const changed = await service.updatePersonalAsset(next, 'deposit', 1, { ...input, name: ' Новое ', amountRub: 42, annualInterest: 0 });
	assert.equal(changed.deposits[0].amountRub, 1000);
	assert.equal(changed.deposits[0].name, 'Новое');
	assert.equal(changed.deposits[0].annualInterest, 0);
	assert.equal((await service.updatePersonalAsset(next, 'deposit', 1, { ...input, startDate: '2024-06-01' })).deposits[0].startDate, input.startDate);
	assert.equal(next.deposits[0].name, 'Вклад');
	changed.insurances[0].name = 'independent';
	assert.equal(next.insurances[0].name, 'Вклад');
	await assert.rejects(service.updatePersonalAsset(next, 'deposit', 99, input), /не найден/);
	await assert.rejects(service.createPersonalAsset(next, 'deposit', { ...input, amountRub: 0 }));
	await assert.rejects(service.updatePersonalAsset(next, 'deposit', 1, { ...input, endDate: 'bad' }));
});

test('close validates withdrawal amount while delete simply erases an entry', async () => {
	const snapshot = await service.createPersonalAsset(logic.emptyPersonalAssets(), 'deposit', input);
	for (const amount of [0, -1, NaN, Infinity]) await assert.rejects(service.closePersonalAsset(snapshot, 'deposit', 1, amount));
	assert.deepEqual(await service.closePersonalAsset(snapshot, 'deposit', 1, 1500), logic.emptyPersonalAssets());
	assert.deepEqual(await service.deletePersonalAsset(snapshot, 'deposit', 1), logic.emptyPersonalAssets());
	assert.equal(snapshot.deposits.length, 1);
	await assert.rejects(service.closePersonalAsset(snapshot, 'insurance', 1, 1), /не найден/);
	await assert.rejects(service.deletePersonalAsset(snapshot, 'deposit', 2), /не найден/);
});

test('duplicate IDs and overflow are rejected independently per type', async () => {
	const duplicated = { deposits: [asset, { ...asset }], insurances: [] };
	await assert.rejects(service.createPersonalAsset(duplicated, 'insurance', input), /идентификатор/);
	await assert.rejects(service.deletePersonalAsset(duplicated, 'deposit', 1), /идентификатор/);
	await assert.rejects(service.createPersonalAsset({ deposits: [{ ...asset, id: Number.MAX_SAFE_INTEGER }], insurances: [] }, 'deposit', input), /идентификатор/);
	await assert.rejects(service.createPersonalAsset({ deposits: [{ ...asset, id: 0 }], insurances: [] }, 'deposit', input), /идентификатор/);
});

test('principal summaries include expired assets and isolate invalid totals', () => {
	const snapshot = { deposits: [asset, { ...asset, id: 2, amountRub: 200 }], insurances: [{ ...asset, type: 'insurance', amountRub: 300 }] };
	assert.deepEqual(logic.summarizePersonalAssets(snapshot), { deposits: 1200, insurances: 300 });
	assert.deepEqual(logic.summarizePersonalAssets(logic.emptyPersonalAssets()), { deposits: 0, insurances: 0 });
	assert.deepEqual(logic.summarizePersonalAssets({ ...snapshot, deposits: [{ ...asset, amountRub: NaN }] }), { deposits: null, insurances: 300 });
	assert.deepEqual(logic.summarizePersonalAssets({ ...snapshot, deposits: [{ ...asset, amountRub: Number.MAX_VALUE }, { ...asset, amountRub: Number.MAX_VALUE }] }), { deposits: null, insurances: 300 });
});
