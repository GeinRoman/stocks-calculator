import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/api/coupons.ts', import.meta.url), 'utf8');
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const { couponSchedule } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);

test('coupon mapping retains unknown and zero amounts and rejects mismatched instruments', () => {
	const payload = { code: 'B', coupons: [{ payment_date: '2030-01-01', amount: null }, { payment_date: '2031-01-01', amount: { price: 0, currency: 'RUB' } }] };
	assert.deepEqual(couponSchedule('B', payload), { code: 'B', coupons: [{ paymentDate: '2030-01-01', amount: null }, { paymentDate: '2031-01-01', amount: { price: 0, currency: 'RUB' } }] });
	assert.throws(() => couponSchedule('A', payload));
});

test('live portfolio attaches all history drawers and live coupon loader', async () => {
	const shell = await readFile(new URL('../src/lib/components/portfolio-app.svelte', import.meta.url), 'utf8');
	for (const kind of ['stock', 'bond', 'crypto', 'deposit', 'insurance']) assert.match(shell, new RegExp(`kind="${kind}"[^>]*loadPage=\\{loadHistory\\}`));
	assert.match(shell, /<Bonds[^>]*\{loadCoupons\}/);
	const dialog = await readFile(new URL('../src/lib/components/tabs/bond-coupons-dialog.svelte', import.meta.url), 'utf8');
	assert.match(dialog, /\(loadCoupons \?\? getBondCoupons\)\(bondCode\)/);
	assert.match(dialog, /summarizeCoupons\(schedule\?\.coupons \?\? \[\], today, quantity, rates\)/);
});
