import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/live-crypto.ts', import.meta.url), 'utf8');
const code = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
const { saveLiveCrypto } = await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`);
const asset = { id: 'new', name: 'BTC', platform: 'Wallet', url: '', annualInterest: 0, amount: 0, price: null };

test('crypto draft sends commands in order and resolves server IDs for subsequent trades', async () => {
	const calls = [];
	const api = {
		addGroups: async groups => calls.push(['add', groups]),
		renameGroup: async group => calls.push(['rename', group]),
		buy: async operations => calls.push(['buy', operations]),
		sell: async operations => calls.push(['sell', operations]),
		updateAsset: async value => calls.push(['edit', value]),
		assets: async () => [{ id: 41, name: 'BTC', group_name: 'Renamed' }]
	};
	await saveLiveCrypto(api, [
		{ kind: 'add-group', name: 'Group' },
		{ kind: 'rename-group', oldName: 'Group', name: 'Renamed' },
		{ kind: 'buy', asset, groupName: 'Renamed', amount: 2, price: 100, currency: 'USDT' },
		{ kind: 'edit-asset', asset: { ...asset, platform: 'Exchange' } },
		{ kind: 'sell', asset: { ...asset, amount: 2 }, groupName: 'Renamed', amount: 1, price: 120, currency: 'USDT' }
	]);
	assert.deepEqual(calls.map(call => call[0]), ['add', 'rename', 'buy', 'edit', 'sell']);
	assert.equal(calls[2][1][0].id, 0);
	assert.deepEqual(calls[2][1][0].price, { price: 100, currency: 'USDT' });
	assert.equal(calls[3][1].id, 41);
	assert.equal(calls[4][1][0].id, 41);
});

test('failure stops the draft and reports already confirmed operations', async () => {
	const calls = [];
	const api = {
		addGroups: async () => calls.push('add'),
		renameGroup: async () => { calls.push('rename'); throw new Error('network'); },
		removeGroups: async () => calls.push('remove')
	};
	await assert.rejects(saveLiveCrypto(api, [
		{ kind: 'add-group', name: 'A' },
		{ kind: 'rename-group', oldName: 'A', name: 'B' },
		{ kind: 'delete-group', name: 'B' }
	]), /подтверждено операций: 1/);
	assert.deepEqual(calls, ['add', 'rename']);
});

test('a fully sold asset is recreated on a later purchase', async () => {
	const buys = [];
	const api = {
		sell: async () => {},
		buy: async ops => buys.push(ops[0]),
		assets: async () => [{ id: 99, name: 'BTC', group_name: 'G' }]
	};
	await saveLiveCrypto(api, [
		{ kind: 'sell', asset: { ...asset, id: '7', amount: 1 }, groupName: 'G', amount: 1, price: 2, currency: 'USDT' },
		{ kind: 'buy', asset: { ...asset, id: '7' }, groupName: 'G', amount: 3, price: 4, currency: 'USDT' },
		{ kind: 'buy', asset: { ...asset, id: '7', amount: 3 }, groupName: 'G', amount: 1, price: 5, currency: 'USDT' }
	]);
	assert.deepEqual(buys.map(value => value.id), [0, 99]);
});

test('manual prices are local to the current profile and never sent as market data', async () => {
	const app = await readFile(new URL('../src/lib/components/portfolio-app.svelte', import.meta.url), 'utf8');
	assert.match(app, /name !== priceProfile[\s\S]*cryptoPrices = \{\}/);
	assert.match(app, /price: cryptoPrices\[asset\.id\] \?\? null/);
	assert.doesNotMatch(source, /asset\.price/);
});
