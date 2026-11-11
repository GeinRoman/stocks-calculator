import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

async function load(name) {
	const source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
	const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
	return import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);
}
const { formatMoney, convertMoney, addDirtyPrice, quoteBond, multiplyMoney } = await load('currency-service');
const { getBondCoupons } = await load('coupon-service');
const usd = { currency: 'USD', price: 10 }, rub = { currency: 'RUB', price: 90 };
test('conversion, missing rates and zeros are distinct', () => {
	assert.deepEqual(convertMoney(usd, 'RUB'), { currency: 'RUB', price: 900 });
	assert.deepEqual(convertMoney({ currency: 'USD', price: 0 }, 'RUB'), { currency: 'RUB', price: 0 });
	assert.equal(convertMoney({ currency: 'XYZ', price: 0 }, 'RUB'), null);
	assert.equal(convertMoney(usd, 'XYZ'), null);
	assert.equal(convertMoney({ currency: '', price: 0 }, 'RUB'), null);
	assert.equal(convertMoney(null, 'RUB'), null);
	assert.match(formatMoney(usd), /USD/);
	assert.notEqual(formatMoney(null), formatMoney({ currency: 'RUB', price: 0 }));
});
test('dirty price and quote modes use native or RUB amounts without altering inputs', () => {
	assert.deepEqual(addDirtyPrice(usd, rub), { currency: 'RUB', price: 990 });
	assert.deepEqual(addDirtyPrice(usd, { currency: 'USD', price: 1 }), { currency: 'USD', price: 11 });
	assert.equal(addDirtyPrice(usd, { currency: 'XYZ', price: 0 }), null);
	const bond = { cleanPrice: usd, accruedInterest: rub };
	assert.deepEqual(quoteBond(bond, false, false), usd);
	assert.deepEqual(quoteBond(bond, true, true), { currency: 'RUB', price: 990 });
	assert.deepEqual(multiplyMoney(usd, 0), { currency: 'USD', price: 0 });
	assert.equal(multiplyMoney(usd, Infinity), null);
	assert.deepEqual(usd, { currency: 'USD', price: 10 });
});
test('coupon schedules are date-only, independent clones and unknown codes reject', async () => {
	const schedule = await getBondCoupons(' ofz-demo-01 ');
	assert.equal(schedule.code, 'OFZ-DEMO-01');
	assert.ok(schedule.coupons.some(c => c.paymentDate < '2026-01-01'));
	assert.ok(schedule.coupons.some(c => c.paymentDate > '2026-01-01'));
	assert.ok(schedule.coupons.every(c => /^\d{4}-\d{2}-\d{2}$/.test(c.paymentDate)));
	schedule.coupons[0].amount.price = 999;
	assert.equal((await getBondCoupons('OFZ-DEMO-01')).coupons[0].amount.price, 40);
	assert.equal((await getBondCoupons('CORP-DEMO-01')).coupons[1].amount, null);
	assert.deepEqual((await getBondCoupons('ZERO-DEMO')).coupons, []);
	await assert.rejects(getBondCoupons('missing'));
});
