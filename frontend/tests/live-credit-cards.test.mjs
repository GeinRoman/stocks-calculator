import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/live-credit-cards.ts', import.meta.url), 'utf8');
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const { saveLiveCreditCard } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
const input = { name: ' Card ', bankName: ' Bank ', paymentDate: '2026-12-01', paymentAmountRub: 0, overallLimitRub: 0 };

test('live card create and update preserve zero values and send backend field names', async () => {
	const calls = [];
	const api = { addCreditCard: async card => calls.push(['create', card]), updateCreditCard: async card => calls.push(['update', card]) };
	await saveLiveCreditCard(api, { kind: 'create', input });
	await saveLiveCreditCard(api, { kind: 'update', id: 7, input });
	const expected = { name: 'Card', bank_name: 'Bank', mandatory_payment_date: '2026-12-01', mandatory_payment_amount_rub: 0, overall_limit_rub: 0 };
	assert.deepEqual(calls, [['create', { ...expected, id: 0 }], ['update', { ...expected, id: 7 }]]);
});

test('live card removal sends only the selected ID and propagates errors', async () => {
	const calls = [];
	const api = { removeCreditCard: async id => { calls.push(id); throw new Error('failed'); } };
	await assert.rejects(saveLiveCreditCard(api, { kind: 'remove', id: 42 }), /failed/);
	assert.deepEqual(calls, [42]);
});

test('live card actions reload server data and block uncertain retries', async () => {
	const page = await readFile(new URL('../src/routes/+page.svelte', import.meta.url), 'utf8');
	const app = await readFile(new URL('../src/lib/components/portfolio-app.svelte', import.meta.url), 'utf8');
	const cards = await readFile(new URL('../src/lib/components/tabs/credit-cards.svelte', import.meta.url), 'utf8');
	const save = page.slice(page.indexOf('async function saveCreditCard'), page.indexOf('async function loadHistory'));
	assert.match(save, /await saveLiveCreditCard\(api\.personal, operation\)/);
	assert.match(save, /await loadPortfolioState\(api\)/);
	assert.match(save, /error\.status >= 500/);
	assert.match(app, /error\.uncertain\) cardBlocked = true/);
	assert.match(cards, /saving \|\| blocked/);
});
