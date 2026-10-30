import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import { compile } from 'svelte/compiler';
import { render } from 'svelte/server';

const source = await readFile(new URL('../src/lib/components/tabs/bond-groups.svelte', import.meta.url), 'utf8');
const { js } = compile(source, { generate: 'server', filename: 'bond-groups.svelte' });
const code = js.code.replace(/(['"])svelte\/internal\/server\1/g, JSON.stringify(import.meta.resolve('svelte/internal/server')));
const { default: BondGroups } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });

async function fields(patch = {}) {
	const bond = {
		code: 'BOND', name: 'Тестовая облигация', price: 950, redemptionPrice: 1000,
		maturityDate: '2030-05-01', couponAmount: 30, couponSchedule: [], quantity: 5,
		totalValue: 4750, remainingCoupons: 4, totalCoupons: 12,
		remainingCouponValue: 120, holdingRemainingCouponValue: 600,
		targetPercent: 100, actualPercent: 100, ...patch
	};
	const { body } = await render(BondGroups, { props: {
		profileName: 'Тестовый профиль',
		portfolio: { totalBondsValue: bond.totalValue, asOf: '2026-10-30T12:00:00Z', groups: [
			{ id: 'bonds', name: 'Облигации', totalValue: bond.totalValue, targetPercent: 100, actualPercent: 100, bonds: [bond] }
		] }
	} });
	return Object.fromEntries([...body.matchAll(/<dt[^>]*>([^<]+)<\/dt>\s*<dd[^>]*>(.*?)<\/dd>/gs)]
		.map((match) => [match[1], match[2].replace(/<!--.*?-->/gs, '').trim()]));
}

test('bond details show current/redemption prices and remaining/lifetime coupon pairs', async () => {
	const values = await fields();
	assert.equal(values['Цена / погашение'], `${currency.format(950)} / ${currency.format(1000)}`);
	assert.equal(values['Купоны, осталось / всего'], '4 / 12');
	assert.equal(values['Количество облигаций'], '5');
	assert.equal(values['Общая стоимость'], currency.format(4750));
	assert.equal(values['Дата погашения'], '01.05.2030');
	assert.equal(values['Размер лота, акций'], undefined);
});

test('coupon cash fields show per-payment, remaining per-bond and whole-holding values separately', async () => {
	const values = await fields();
	assert.equal(values['Купон на 1 облигацию'], currency.format(30));
	assert.equal(values['Оставшиеся купоны на 1 облигацию'], currency.format(120));
	assert.equal(values['Оставшиеся купоны на все облигации'], currency.format(600));
});

test('zero prices and coupon values are rendered as zero rather than placeholders', async () => {
	const values = await fields({ price: 0, redemptionPrice: 0, couponAmount: 0, totalValue: 0,
		remainingCoupons: 0, totalCoupons: 0, remainingCouponValue: 0, holdingRemainingCouponValue: 0 });
	assert.equal(values['Цена / погашение'], `${currency.format(0)} / ${currency.format(0)}`);
	assert.equal(values['Купоны, осталось / всего'], '0 / 0');
	for (const label of ['Купон на 1 облигацию', 'Общая стоимость', 'Оставшиеся купоны на 1 облигацию', 'Оставшиеся купоны на все облигации']) {
		assert.equal(values[label], currency.format(0));
	}
});

test('missing bond metadata is not represented by invented zero values', async () => {
	const values = await fields({ redemptionPrice: null, maturityDate: null, couponAmount: null,
		remainingCoupons: null, totalCoupons: null, remainingCouponValue: null, holdingRemainingCouponValue: null });
	assert.equal(values['Цена / погашение'], `${currency.format(950)} / Нет данных`);
	assert.equal(values['Купоны, осталось / всего'], 'Нет данных / Нет данных');
	for (const label of ['Дата погашения', 'Купон на 1 облигацию', 'Оставшиеся купоны на 1 облигацию', 'Оставшиеся купоны на все облигации']) {
		assert.equal(values[label], 'Нет данных');
	}
});
