import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/tabs/transaction-history.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { demoHistoryPage, historyTime } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('history pages are newest first and use timestamp and ID at ties', () => {
	for (const kind of ['stock', 'bond', 'crypto', 'deposit', 'insurance']) {
		const first = demoHistoryPage(kind, 'main');
		assert.equal(first.items.length, 3);
		const last = first.items.at(-1);
		const second = demoHistoryPage(kind, 'main', { before: historyTime(last), beforeId: last.id });
		assert.equal(second.items.length, 1);
		const secondLast = second.items.at(-1);
		assert.deepEqual(demoHistoryPage(kind, 'main', { before: historyTime(secondLast), beforeId: secondLast.id }), { items: [] });
		const all = [...first.items, ...second.items];
		assert.equal(new Set(all.map(item => item.id)).size, all.length);
		assert.ok(all.every((item, index) => index === 0 || historyTime(all[index - 1]) >= historyTime(item)));
	}
	const firstStock = demoHistoryPage('stock', 'main').items;
	const tiePage = demoHistoryPage('stock', 'main', { before: historyTime(firstStock[0]), beforeId: firstStock[0].id });
	assert.equal(tiePage.items[0].id, firstStock[1].id);
});

test('empty profiles and results do not leak sample data or shared mutable records', () => {
	assert.deepEqual(demoHistoryPage('stock', 'reserve'), { items: [] });
	const a = demoHistoryPage('crypto', 'main'); a.items[0].asset_name = 'changed';
	assert.notEqual(demoHistoryPage('crypto', 'main').items[0].asset_name, 'changed');
	assert.deepEqual(demoHistoryPage('bond', 'main', { before: '2000-01-01T00:00:00Z', beforeId: 1 }), { items: [] });
});
