import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';
import { compile } from 'svelte/compiler';
import { render } from 'svelte/server';
import { groupLayoutModuleUrl } from './group-layout-module.mjs';

const source = await readFile(new URL('../src/lib/components/tabs/bond-groups.svelte', import.meta.url), 'utf8');
const currencySource = await readFile(new URL('../src/lib/logic/tabs/currency-service.ts', import.meta.url), 'utf8');
const currencyCode = ts.transpileModule(currencySource, { compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022 } }).outputText;
const dataUrl = (text) => `data:text/javascript;base64,${Buffer.from(text).toString('base64')}`;
const currencyUrl = dataUrl(currencyCode);
const stubUrl = dataUrl('export function Button(renderer, props) { props.children?.(renderer); } export default function Dialog() {}');
const { js } = compile(source, { generate: 'server', filename: 'bond-groups.svelte' });
const code = js.code.replace(/(['"])svelte\/internal\/server\1/g, JSON.stringify(import.meta.resolve('svelte/internal/server')))
	.replace(/(['"])#lib\/logic\/tabs\/group-layout\.js\1/g, JSON.stringify(groupLayoutModuleUrl))
	.replace(/(['"])#lib\/logic\/tabs\/currency-service\.js\1/g, JSON.stringify(currencyUrl))
	.replace(/(['"])#lib\/components\/ui\/button\/index\.js\1/g, JSON.stringify(stubUrl))
	.replace(/(['"])\.\/bond-coupons-dialog\.svelte\1/g, JSON.stringify(stubUrl));
const { default: BondGroups } = await import(dataUrl(code));

async function fields(patch = {}, options = {}) {
	const bond = {
		code: 'BOND', name: 'Тестовая облигация', isin: '', cleanPrice: { currency: 'USD', price: 10 },
		accruedInterest: { currency: 'USD', price: 1 }, maturityPayment: { currency: 'USD', price: 12 },
		maturityDate: '2030-05-01', lotCount: 5, lotSize: 2, totalValueRub: 9900,
		targetPercent: 100, actualPercent: 100, ...patch
	};
	const { body } = await render(BondGroups, { props: {
		profileName: 'Тестовый профиль', ...options,
		portfolio: { totalBondsValueRub: 9900, asOf: '2026-10-30T12:00:00Z', groups: [
			{ id: 'bonds', name: 'Облигации', totalValueRub: 9900, targetPercent: 100, actualPercent: 100, bonds: [bond] }
		] }
	} });
	return { body, values: Object.fromEntries([...body.matchAll(/<dt[^>]*>([^<]+)<\/dt>\s*<dd[^>]*>(.*?)<\/dd>/gs)]
		.map((match) => [match[1], match[2].replace(/<[^>]*>/g, '').trim()])) };
}

test('bond row shows general bond fields, lots as individual bonds, and coupon button', async () => {
	const { body, values } = await fields();
	assert.equal(values['Цена с НКД'], '11 USD');
	assert.match(values['Количество облигаций'], /10.*Размер лота: 2/s);
	assert.equal(values['Общая стоимость'], '110 USD');
	assert.equal(values['Выплата при погашении'], '12 USD');
	assert.equal(values['Дата погашения'], '01.05.2030');
	assert.ok(body.includes('Купоны'));
	assert.ok(!body.includes('Купоны, осталось / всего'));
});

test('clean and RUB display modes alter bond values without changing RUB group totals', async () => {
	const { values } = await fields({}, { includeAccrued: false, inRub: true });
	assert.equal(values['Чистая цена'], '900 RUB');
	assert.equal(values['Общая стоимость'], '9 000 RUB');
	assert.equal(values['Выплата при погашении'], '1 080 RUB');
	assert.match(values['Текущая доля'], /^100\s%$/);
});

test('missing conversion and maturity are shown as unavailable while zero remains zero', async () => {
	const { values } = await fields({ cleanPrice: { currency: 'XYZ', price: 0 }, accruedInterest: { currency: 'XYZ', price: 0 }, maturityPayment: null, maturityDate: null }, { inRub: true });
	assert.equal(values['Цена с НКД'], 'Нет данных');
	assert.equal(values['Общая стоимость'], 'Нет данных');
	assert.equal(values['Выплата при погашении'], 'Нет данных');
	assert.equal(values['Дата погашения'], 'Нет данных');
	const zero = await fields({ cleanPrice: { currency: 'USD', price: 0 }, accruedInterest: { currency: 'USD', price: 0 } });
	assert.equal(zero.values['Цена с НКД'], '0 USD');
});
