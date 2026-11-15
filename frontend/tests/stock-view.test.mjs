import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/api/stock-view.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { stockView } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('stock view derives lot values and percentages while retaining empty groups', () => {
	const view = stockView({
		stock_groups: [{ name: 'A', weight: 60 }, { name: 'Empty', weight: 0 }],
		stocks: [
			{ name: 'One', code: 'ONE', group_name: 'A', price: 10, lot_amount: 2, lot_size: 5 },
			{ name: 'Two', code: 'TWO', group_name: 'A', price: 50, lot_amount: 1, lot_size: 2 }
		]
	});
	assert.equal(view.total, 200);
	assert.equal(view.groups[0].actualPercent, 100);
	assert.deepEqual(view.groups[0].stocks.map(stock => [stock.totalValue, stock.actualPercent, stock.targetPercent]), [[100, 50, 30], [100, 50, 30]]);
	assert.deepEqual(view.groups[1], { id: 'Empty', name: 'Empty', totalValue: 0, targetPercent: 0, actualPercent: 0, stocks: [] });
});

test('stock view does not invent percentages for a zero-valued portfolio', () => {
	const view = stockView({ stock_groups: [{ name: 'A', weight: 0 }], stocks: [] });
	assert.equal(view.total, 0);
	assert.equal(view.groups[0].actualPercent, null);
});
