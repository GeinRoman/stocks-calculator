import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const compile = source => ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const url = code => `data:text/javascript;base64,${Buffer.from(code).toString('base64')}`;
let code = compile(await readFile(new URL('../src/lib/logic/live-weights.ts', import.meta.url), 'utf8'));
for (const [name, path] of [['stock-editor', 'logic/tabs/stock-editor'], ['stock-view', 'api/stock-view'], ['bond-view', 'api/bond-view']]) {
	const source = await readFile(new URL(`../src/lib/${path}.ts`, import.meta.url), 'utf8');
	code = code.replace(`#lib/${path}.js`, url(compile(source)));
}
const { saveLiveWeights } = await import(url(code));
const state = () => ({ data: { stock_groups: [{ name: 'A', weight: 0 }, { name: 'B', weight: 100 }], stocks: [], bond_groups: [{ name: 'Debt', weight: 0 }], bonds: [], crypto_groups: [{ name: 'Coins', weight: 0 }] }, rates: {}, loadedAt: '2026-01-01T00:00:00Z' });

test('weights commit only after the correct endpoint succeeds without reloading', async () => {
	let resolve;
	const calls = [];
	const api = { stocks: { updateWeights: groups => { calls.push(groups); return new Promise(done => resolve = done); } }, bonds: { updateWeights: () => assert.fail('wrong endpoint') } };
	const original = state();
	const pending = saveLiveWeights(api, original, 'stock', [{ id: 'A', value: 40 }, { id: 'B', value: 60 }]);
	assert.equal(original.data.stock_groups[0].weight, 0);
	resolve();
	const result = await pending;
	assert.deepEqual(calls, [[{ name: 'A', weight: 40 }, { name: 'B', weight: 60 }]]);
	assert.equal(result.stocks.groups[0].targetPercent, 40);
	assert.equal(original.data.stock_groups[0].weight, 0);
});

test('bond failures and invalid group lists preserve the original state', async () => {
	const original = state();
	let calls = 0;
	const api = { stocks: {}, bonds: { updateWeights: async () => { calls++; throw new Error('offline'); } } };
	await assert.rejects(saveLiveWeights(api, original, 'bond', [{ id: 'Debt', value: 100 }]), /Не удалось сохранить/);
	await assert.rejects(saveLiveWeights(api, original, 'bond', [{ id: 'Other', value: 100 }]), /Список групп/);
	assert.equal(calls, 1);
	assert.equal(original.data.bond_groups[0].weight, 0);
});

test('crypto weights use the crypto endpoint and preserve the snapshot on failure', async () => {
	const original = state();
	let fail = true;
	const api = { stocks: {}, bonds: {}, crypto: { updateWeights: async groups => { assert.deepEqual(groups, [{ name: 'Coins', weight: 100 }]); if (fail) throw new Error('offline'); } } };
	await assert.rejects(saveLiveWeights(api, original, 'crypto', [{ id: 'Coins', value: 100 }]), /Не удалось сохранить/);
	assert.equal(original.data.crypto_groups[0].weight, 0);
	fail = false;
	const result = await saveLiveWeights(api, original, 'crypto', [{ id: 'Coins', value: 100 }]);
	assert.equal(result.data.crypto_groups[0].weight, 100);
});
