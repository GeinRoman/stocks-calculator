import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/portfolio-back-state.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { isBackPastLayer } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('Back acts only when navigation passes the active tab, editor or drawer layer', () => {
	assert.equal(isBackPastLayer([], []), false);
	assert.equal(isBackPastLayer(['tab'], ['tab']), false);
	assert.equal(isBackPastLayer(['tab'], []), true);
	assert.equal(isBackPastLayer(['tab', 'editor'], ['tab']), true);
	assert.equal(isBackPastLayer(['tab', 'editor', 'drawer'], ['tab', 'editor']), true);
	assert.equal(isBackPastLayer(['tab', 'editor'], ['tab', 'editor', 'forward']), false);
});
