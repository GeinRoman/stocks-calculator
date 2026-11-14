import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/tabs/draft-lot-change.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { draftLotChange } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('draft sub-rows describe added and changed lots without inventing transactions', () => {
	assert.equal(draftLotChange(3, 3), null);
	assert.equal(draftLotChange(undefined, 2), 'Добавлено: 2 лот. (черновик)');
	assert.equal(draftLotChange(3, 5), 'Лоты: 3 → 5 (черновик)');
	assert.equal(draftLotChange(3, 1), 'Лоты: 3 → 1 (черновик)');
	assert.equal(draftLotChange(3, 0), 'Лоты: 3 → 0 (черновик)');
	assert.equal(draftLotChange(3, NaN), null);
});
