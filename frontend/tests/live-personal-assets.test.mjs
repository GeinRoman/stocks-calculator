import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/live-personal-assets.ts', import.meta.url), 'utf8');
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const { saveLivePersonalAsset } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
const input = { name: ' Name ', provider: ' Provider ', amountRub: 100, annualInterest: 0, startDate: '2026-01-01', endDate: '2027-01-01' };

test('deposit and insurance create and update use their respective backend contracts', async () => {
	const calls = [];
	const api = Object.fromEntries(['addDeposit', 'updateDeposit', 'addInsurance', 'updateInsurance'].map(key => [key, async body => calls.push([key, body])]));
	for (const type of ['deposit', 'insurance']) {
		await saveLivePersonalAsset(api, { kind: 'create', type, input });
		await saveLivePersonalAsset(api, { kind: 'update', type, id: 4, input });
	}
	assert.deepEqual(calls.map(([name]) => name), ['addDeposit', 'updateDeposit', 'addInsurance', 'updateInsurance']);
	assert.deepEqual(calls[0][1], { id: 0, name: 'Name', bank_name: 'Provider', amount_rub: 100, annual_interest: 0, start_date: '2026-01-01', end_date: '2027-01-01' });
	assert.equal(calls[2][1].company_name, 'Provider');
	for (const index of [1, 3]) {
		assert.equal(calls[index][1].id, 4);
		assert.equal('amount_rub' in calls[index][1], false);
		assert.equal(calls[index][1].annual_interest, 0);
	}
});

test('remove and close call the matching endpoint without inventing withdrawal amounts', async () => {
	const calls = [];
	const api = Object.fromEntries(['removeDeposit', 'closeDeposit', 'removeInsurance', 'closeInsurance'].map(key => [key, async (...args) => calls.push([key, ...args])]));
	for (const type of ['deposit', 'insurance']) {
		await saveLivePersonalAsset(api, { kind: 'remove', type, id: 1 });
		await saveLivePersonalAsset(api, { kind: 'close', type, id: 2, amount: 123.45 });
	}
	assert.deepEqual(calls, [['removeDeposit', 1], ['closeDeposit', 2, 123.45], ['removeInsurance', 1], ['closeInsurance', 2, 123.45]]);
});

test('live cards and closing form show the maturity estimate while keeping principal immutable', async () => {
	const editor = await readFile(new URL('../src/lib/components/tabs/personal-asset-dialog.svelte', import.meta.url), 'utf8');
	const close = await readFile(new URL('../src/lib/components/tabs/personal-asset-close-dialog.svelte', import.meta.url), 'utf8');
	const cards = await readFile(new URL('../src/lib/components/tabs/personal-assets.svelte', import.meta.url), 'utf8');
	assert.match(editor, /readonly=\{live && !!asset\}/);
	assert.match(close, /const estimate = untrack\(\(\) => estimatedMaturityAmount\(asset\)\)/);
	assert.match(close, /estimate === null \? ''/);
	assert.match(close, /Проверьте и при необходимости измените сумму на фактически полученную/);
	assert.match(cards, /@const maturity = estimatedMaturityAmount\(asset\)/);
	assert.match(cards, /<p class="text-sm">Оценка к окончанию:/);
});
