import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/tabs/coupon-display.ts', import.meta.url), 'utf8');
const currencySource = await readFile(new URL('../src/lib/logic/tabs/currency-service.ts', import.meta.url), 'utf8');
const currencyCode = ts.transpileModule(currencySource, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const currencyUrl = `data:text/javascript;base64,${Buffer.from(currencyCode).toString('base64')}`;
let { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
outputText = outputText.replace(/(['"])#lib\/logic\/tabs\/currency-service\.js\1/g, JSON.stringify(currencyUrl));
const { summarizeCoupons, localCalendarDate } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);
const coupon = (paymentDate, price, currency = 'RUB') => ({ paymentDate, amount: price === null ? null : { currency, price } });
const today = '2026-10-30';

test('sorts dates and includes today with variable amounts and individual bond quantity', () => {
	const result = summarizeCoupons([coupon('2027-01-01', 2.25), coupon('2026-01-01', null), coupon(today, 1.5)], today, 3);
	assert.deepEqual(result.sorted.map((item) => item.paymentDate), ['2026-01-01', today, '2027-01-01']);
	assert.equal(result.remaining, 2);
	assert.deepEqual(result.perBond, { currency: 'RUB', price: 3.75 });
	assert.deepEqual(result.holding, { currency: 'RUB', price: 11.25 });
});

test('zero is known, but no future or unknown future is unavailable', () => {
	assert.deepEqual(summarizeCoupons([coupon(today, 0)], today, 2).holding, { currency: 'RUB', price: 0 });
	assert.equal(summarizeCoupons([coupon('2020-01-01', 1)], today, 2).perBond, null);
	assert.equal(summarizeCoupons([], today, 2).holding, null);
	assert.equal(summarizeCoupons([coupon(today, null)], today, 2).perBond, null);
});

test('mixed currencies use RUB, while missing rates and unsafe numbers remain unavailable', () => {
	assert.deepEqual(summarizeCoupons([coupon(today, 1), coupon(today, 2, 'USD')], today, 2).perBond, { currency: 'RUB', price: 181 });
	assert.equal(summarizeCoupons([coupon(today, 1), coupon(today, 2, 'XYZ')], today, 2).perBond, null);
	assert.equal(summarizeCoupons([coupon(today, Infinity)], today, 2).perBond, null);
	assert.equal(summarizeCoupons([coupon(today, Number.MAX_SAFE_INTEGER)], today, 2).holding, null);
	assert.equal(summarizeCoupons([coupon(today, 1)], today, 1.5).holding, null);
});

test('local date uses local calendar fields', () => {
	const date = new Date(2026, 0, 2, 23, 59);
	assert.equal(localCalendarDate(date), '2026-01-02');
});
