import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/demo-auth.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { DEMO_REFRESH_KEY, validateDemoCredentials, demoAuthenticate, demoRefresh, demoLogout } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);
function storage() {
	const values = new Map();
	return { getItem: key => values.get(key) ?? null, setItem: (key, value) => values.set(key, value), removeItem: key => values.delete(key), values };
}

test('demo login and registration credentials use exact input and UTF-8 byte limit', () => {
	assert.equal(validateDemoCredentials('user', '123'), null);
	for (const [user, pass] of [['', '123'], ['user', ''], ['USER', '123'], ['user', 'wrong'], ['user', 'я'.repeat(33)]]) assert.ok(validateDemoCredentials(user, pass));
	assert.match(validateDemoCredentials('user', 'я'.repeat(33)), /64 байта/);
});
test('demo refresh persists a marker and logout removes it without storing passwords', async () => {
	const store = storage();
	assert.equal(await demoRefresh(store), false);
	await assert.rejects(demoAuthenticate('user', 'wrong', store));
	assert.equal(store.values.size, 0);
	await demoAuthenticate('user', '123', store);
	assert.equal(await demoRefresh(store), true);
	assert.equal([...store.values.values()].includes('123'), false);
	demoLogout(store);
	assert.equal(await demoRefresh(store), false);
	store.setItem(DEMO_REFRESH_KEY, 'invalid');
	assert.equal(await demoRefresh(store), false);
	assert.equal(store.getItem(DEMO_REFRESH_KEY), null);
});
test('storage failures do not claim a successful login or refresh', async () => {
	const blocked = { getItem() { throw Error(); }, setItem() { throw Error(); }, removeItem() { throw Error(); } };
	await assert.rejects(demoAuthenticate('user', '123', blocked), /не удалось сохранить/i);
	await assert.rejects(demoRefresh(blocked));
	assert.throws(() => demoLogout(blocked));
});
