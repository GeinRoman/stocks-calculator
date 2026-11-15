import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/tabs/draft-confirmation.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { hasDraftChanges, hasRebalanceChanges } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('confirmations distinguish unchanged drafts, edits and pending transactions', () => {
	const initial = JSON.stringify([{ name: 'group', count: 10 }]);
	assert.equal(hasDraftChanges(initial, [{ name: 'group', count: 10 }]), false);
	assert.equal(hasDraftChanges(initial, [{ name: 'group', count: 0 }]), true);
	assert.equal(hasDraftChanges(initial, [{ name: 'group', count: 10 }], 2), true);
	assert.equal(hasRebalanceChanges([{ quantity: 0 }]), false);
	assert.equal(hasRebalanceChanges([{ quantity: 1 }]), true);
	assert.equal(hasRebalanceChanges([{ quantity: undefined }]), true);
});
