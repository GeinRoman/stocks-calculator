import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/tabs/demo-backup.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
const { parseDemoBackup } = await import(`data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`);
const entry = (id) => ({
	profile: { id, name: id },
	stocks: { totalStocksValue: 0, groups: [], assetValues: { stocks: 0, bonds: 0, deposits: 0 } },
	bonds: { totalBondsValueRub: 0, groups: [], asOf: '2026-01-01T00:00:00.000Z' },
	crypto: { groups: [] }, personal: { deposits: [], insurances: [] }, creditCards: { cards: [] }
});
const backup = () => ({ format: 'stocks-calculator-demo-backup', version: 1, selectedId: 'one', profiles: [entry('one'), entry('two')] });

test('demo backup round-trips every profile and preserves zero and null values', () => {
	const data = backup();
	data.profiles[0].crypto = { priceCurrency: 'BTC', groups: [{ id: 'g', name: 'group', weight: 0, assets: [{ id: 'a', name: 'BTC', platform: 'wallet', url: '', annualInterest: 0, amount: 0.001, price: null }] }] };
	assert.deepEqual(parseDemoBackup(JSON.stringify(data)), data);
});
test('rejects malformed, incomplete, unsafe, or non-demo backups before replacement', () => {
	assert.throws(() => parseDemoBackup('{'));
	assert.throws(() => parseDemoBackup('{"version":1,"profiles":[]}'));
	for (const change of [
		v => v.version = 2,
		v => v.profiles = [],
		v => v.selectedId = 'missing',
		v => v.profiles[1].profile.id = 'one',
		v => v.profiles[0].profile.id = '__proto__',
		v => v.profiles[0].bonds.groups = [{}],
		v => v.profiles[0].personal.deposits = [{ id: 1 }],
		v => v.profiles[0].creditCards.cards = [{ id: 1 }],
		v => v.profiles[0].stocks.assetValues = null
	]) { const value = backup(); change(value); assert.throws(() => parseDemoBackup(JSON.stringify(value))); }
});
