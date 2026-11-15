import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/live-snapshots.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { liveSnapshots } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);

test('live records are mapped to the existing tab models without inventing crypto prices', () => {
	const snapshot = liveSnapshots({ crypto_groups: [{ name: 'Coins', weight: 100 }, { name: 'Empty', weight: 0 }], crypto_assets: [{ id: 3, name: 'Coin', group_name: 'Coins', platform: 'Wallet', url: '', annual_interest: 1, amount: 2 }], personal_assets: { deposits: [{ id: 1, name: 'Deposit', bank_name: 'Bank', amount_rub: 50, annual_interest: 5, start_date: '2026-01-01', end_date: '2027-01-01' }], insurances: [{ id: 2, name: 'Insurance', company_name: 'Company', amount_rub: 20, annual_interest: 3, start_date: '2026-01-01', end_date: '2027-01-01' }], credit_cards: [{ id: 4, name: 'Card', bank_name: 'Bank', mandatory_payment_date: '2026-12-01', mandatory_payment_amount_rub: 10, overall_limit_rub: 100 }] } });
	assert.equal(snapshot.crypto.groups[0].assets[0].price, null);
	assert.equal(snapshot.crypto.groups[0].assets[0].id, '3');
	assert.deepEqual(snapshot.crypto.groups[1].assets, []);
	assert.equal(snapshot.personal.deposits[0].provider, 'Bank');
	assert.equal(snapshot.personal.insurances[0].provider, 'Company');
	assert.equal(snapshot.cards.cards[0].paymentAmountRub, 10);
});
