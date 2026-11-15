import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/sidebar-viewport.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { observeSidebarViewport } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('sidebar viewport initializes immediately, follows changes, and removes its listener', () => {
	let listener;
	const media = { matches: true, addEventListener(type, fn) { assert.equal(type, 'change'); listener = fn; }, removeEventListener(type, fn) { assert.equal(type, 'change'); assert.equal(listener, fn); listener = undefined; } };
	const states = [];
	const stop = observeSidebarViewport(media, narrow => states.push(narrow));
	assert.deepEqual(states, [true]);
	media.matches = false; listener();
	media.matches = true; listener();
	assert.deepEqual(states, [true, false, true]);
	stop(); assert.equal(listener, undefined);
});
