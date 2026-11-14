import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

async function moduleUrl(name, dependencies = {}) {
	let source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
	for (const [path, url] of Object.entries(dependencies)) source = source.replaceAll(`'./${path}.js'`, `'${url}'`);
	const { outputText } = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } });
	return `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`;
}
const logicUrl = await moduleUrl('crypto');
const logic = await import(logicUrl);
const service = await import(await moduleUrl('crypto-service', { crypto: logicUrl }));

test('fractional crypto trades avoid binary accumulation and reject invalid sales', () => {
	assert.equal(logic.transactAmount(0.1, 0.2, false), 0.3);
	assert.equal(logic.transactAmount(0.3, 0.2, true), 0.1);
	assert.equal(logic.transactAmount(0.00000001, 0.00000001, true), 0);
	assert.throws(() => logic.transactAmount(0.1, 0.2, true));
	assert.throws(() => logic.transactAmount(1, 0, false));
	assert.equal(logic.decimal('0,125'), 0.125);
	assert.equal(logic.decimal('1abc'), null);
});

test('price inputs accept only digits and one decimal point, including pasted text', () => {
	assert.equal(logic.priceInput('1a2.3,4-5.6e7'), '12.34567');
	assert.equal(logic.priceInput('.01'), '.01');
	assert.equal(logic.priceInputValue(1e-8), '0.00000001');
	assert.equal(Number(logic.priceInputValue(1e+21)), 1e+21);
	assert.equal(Number(logic.priceInputValue(5e-324)), 5e-324);
});

test('sample valuations do not mistake transaction prices for market prices', () => {
	const asset = { id: 'a', name: 'BTC', platform: 'X', url: '', annualInterest: 0, amount: 1, price: null };
	service.trade(asset, 0.25, 100, 'USD', false);
	assert.equal(asset.amount, 1.25);
	assert.deepEqual(logic.cryptoValue([asset]), { value: null, partial: true });
	assert.deepEqual(logic.cryptoValue([]), { value: 0, partial: false });
	assert.deepEqual(logic.cryptoValue([asset, { ...asset, price: 10 }]), { value: 12.5, partial: true });
	assert.equal(logic.decimal('1e-8'), 1e-8);
	assert.throws(() => service.trade(asset, 1, 0, 'USD', false));
});

test('mock mutations are isolated and validate names and safe URLs', () => {
	const original = logic.demoCrypto().main;
	const changed = service.updateCrypto(original, draft => { draft.groups[0].name = 'New'; });
	assert.notEqual(original.groups[0].name, changed.groups[0].name);
	assert.throws(() => service.groupName(original, original.groups[0].name));
	assert.throws(() => service.groupName(original, 'я'.repeat(33)));
	assert.throws(() => service.checkMetadata(original, { name: 'Bitcoin', platform: 'X', annualInterest: 0, url: '' }));
	assert.ok(logic.validateMetadata({ name: 'X', platform: 'Y', annualInterest: 0, url: 'javascript:alert(1)' }));
	assert.equal(logic.validateMetadata({ name: 'X', platform: 'Y', annualInterest: 0, url: 'https://example.com' }), null);
});

test('crypto weight saves require valid complete group entries totaling 100', () => {
	const original = logic.demoCrypto().main;
	const weights = [{ id: 'coins', value: 40 }, { id: 'other', value: 60 }];
	const result = service.saveCryptoWeights(original, weights);
	assert.equal(result.groups[0].weight, 40);
	assert.equal(original.groups[0].weight, 70);
	assert.throws(() => service.saveCryptoWeights(original, [{ id: 'coins', value: 40 }, { id: 'other', value: 30 }]));
	assert.throws(() => service.saveCryptoWeights(original, [{ id: 'coins', value: 80 }, { id: 'other', value: 30 }]));
	assert.throws(() => service.saveCryptoWeights(original, [{ id: 'coins', value: null }, { id: 'other', value: 30 }]));
	assert.throws(() => service.saveCryptoWeights(original, [{ id: 'coins', value: 1 }, { id: 'coins', value: 2 }]));
});
