import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/tabs/personal-asset-date.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { parseDateText, formatDateISO, monthDays, dateISO } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('strictly parses and formats real calendar dates', () => {
	for (const [text, iso] of [['29/02/2024', '2024-02-29'], ['01/01/0001', '0001-01-01'], ['31/12/9999', '9999-12-31'], ['29/02/2000', '2000-02-29']]) {
		assert.equal(parseDateText(text), iso);
		assert.equal(formatDateISO(iso), text);
	}
	for (const text of ['', '1/01/2024', '01-01-2024', ' 01/01/2024', '29/02/1900', '31/04/2024', '00/01/2024', '01/13/2024', '01/01/0000']) assert.equal(parseDateText(text), null);
	for (const iso of ['2023-02-29', '0000-01-01', '2024-1-01']) assert.equal(formatDateISO(iso), '');
});

test('calendar starts on Monday and handles early years and leap months', () => {
	assert.equal(monthDays(1, 1)[0], 1);
	assert.deepEqual(monthDays(2024, 2).slice(0, 5), [null, null, null, 1, 2]);
	assert.equal(monthDays(2024, 2).at(-1), 29);
	assert.equal(monthDays(1900, 2).at(-1), 28);
	assert.equal(dateISO(1, 1, 2), '0001-01-02');
});
