import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import { compile } from 'svelte/compiler';
import { render } from 'svelte/server';

const source = await readFile(new URL('../src/lib/components/stocks.svelte', import.meta.url), 'utf8');
const { js } = compile(source, { generate: 'server', filename: 'stocks.svelte' });
// Resolve the compiler's runtime import before loading the component from memory.
const code = js.code.replace(/(['"])svelte\/internal\/server\1/g, JSON.stringify(import.meta.resolve('svelte/internal/server')));
const { default: Stocks } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
const percent = new Intl.NumberFormat('ru-RU', { style: 'unit', unit: 'percent', maximumFractionDigits: 2 });

async function allocations(groupPercentages, stockPercentages) {
	const { body } = await render(Stocks, {
		props: {
			profileName: 'Тестовый профиль',
			portfolio: {
				totalStocksValue: 1000,
				assetValues: { stocks: 1000, bonds: 0, deposits: 0 },
				groups: [{
					id: 'banks', name: 'Банки', totalValue: 1000, ...groupPercentages,
					stocks: [{ code: 'SBER', name: 'Сбербанк', price: 100, lotCount: 1, lotSize: 10, totalValue: 1000, ...stockPercentages }]
				}]
			}
		}
	});
	return [...body.matchAll(/<dt[^>]*>(Целевая доля|Текущая доля)<\/dt>\s*<dd[^>]*>(.*?)<\/dd>/gs)]
		.map((match) => match[2].replace(/<!--.*?-->/gs, '').trim());
}

test('omitted percentages display placeholders at group and stock level', async () => {
	assert.deepEqual(await allocations({}, {}), ['Не задана', 'Нет данных', 'Не задана', 'Нет данных']);
});

test('null percentages display placeholders rather than 0%', async () => {
	const missing = { targetPercent: null, actualPercent: null };
	assert.deepEqual(await allocations(missing, missing), ['Не задана', 'Нет данных', 'Не задана', 'Нет данных']);
});

test('actual percentages remain visible without a target', async () => {
	assert.deepEqual(await allocations({ actualPercent: 40 }, { actualPercent: 25 }),
		['Не задана', percent.format(40), 'Не задана', percent.format(25)]);
});

test('stock targets are not shown when the group target is unset', async () => {
	assert.deepEqual(await allocations({ actualPercent: 40 }, { targetPercent: 15, actualPercent: 25 }),
		['Не задана', percent.format(40), 'Не задана', percent.format(25)]);
});

test('an unset stock target has its own placeholder', async () => {
	assert.deepEqual(await allocations({ targetPercent: 30, actualPercent: 40 }, { actualPercent: 25 }),
		[percent.format(30), percent.format(40), 'Не задана', percent.format(25)]);
});

test('zero percentages are valid values, not placeholders', async () => {
	const zero = { targetPercent: 0, actualPercent: 0 };
	assert.deepEqual(await allocations(zero, zero), Array(4).fill(percent.format(0)));
});
