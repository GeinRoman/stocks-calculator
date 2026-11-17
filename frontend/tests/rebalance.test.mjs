import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import ts from 'typescript';

const modules = new Map();
async function compileModule(name) {
	if (modules.has(name)) return modules.get(name);
	const source = await readFile(new URL(`../src/lib/logic/tabs/${name}.ts`, import.meta.url), 'utf8');
	let { outputText } = ts.transpileModule(source, {
		compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext }
	});
	for (const [, specifier, dependency] of [...outputText.matchAll(/['"]((?:#lib\/logic\/tabs\/|\.\/)([\w-]+)\.js)['"]/g)]) {
		const url = await compileModule(dependency);
		outputText = outputText.replaceAll(`'${specifier}'`, JSON.stringify(url)).replaceAll(`"${specifier}"`, JSON.stringify(url));
	}
	const url = `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`;
	modules.set(name, url);
	return url;
}

const { requestRebalance, applyRebalance, sendRebalanceChanges } = await import(await compileModule('rebalance-service'));
const { validateRebalanceRequest, validateRebalanceProposal, getRebalanceEstimate, getRebalanceErrorMessage } = await import(await compileModule('rebalance-validation'));
const { demoPortfolios } = await import(await compileModule('demo-portfolio'));
const { createDemoBondPortfolios, sampleBonds } = await import(await compileModule('demo-bonds'));
const snapshots = () => ({ stocks: structuredClone(demoPortfolios.main), bonds: createDemoBondPortfolios(new Date('2026-10-30T12:00:00Z')).main });
const request = (overrides = {}) => ({ profileId: 'main', stocks: 100, bonds: 100, noSell: false, ...overrides });

test('mock proposals use selected existing holdings and no-sell requires positive deposits globally', async () => {
	const data = snapshots();
	const proposal = await requestRebalance(request({ bonds: null, noSell: true }), data);
	assert.equal(proposal.profileId, 'main');
	assert.ok(proposal.trades.length > 0);
	assert.ok(proposal.trades.every((trade) => trade.asset === 'stocks' && trade.side === 'buy' &&
		data.stocks.groups.some((group) => group.id === trade.groupId && group.stocks.some((stock) => stock.code === trade.code))));
	for (const amount of [0, -1]) {
		assert.match(validateRebalanceRequest(request({ bonds: amount, noSell: true }), data), /положительное пополнение/);
		await assert.rejects(requestRebalance(request({ stocks: amount, bonds: null, noSell: true }), data), /положительное пополнение/);
	}
	assert.match(validateRebalanceRequest(request({ stocks: null, bonds: null }), data), /хотя бы один/);
	assert.match(validateRebalanceRequest(request({ profileId: ' ' }), data), /профиль/);
	assert.match(validateRebalanceRequest(request({ stocks: NaN }), data), /Введите сумму/);
});

test('selected group targets must all be initialized and total exactly 100', async () => {
	for (const asset of ['stocks', 'bonds']) {
		for (const value of [null, 0, 101]) {
			const data = snapshots();
			data[asset].groups[0].targetPercent = value;
			assert.ok(validateRebalanceRequest(request(), data));
			await assert.rejects(requestRebalance(request(), data));
		}
	}
	const data = snapshots();
	data.bonds.groups[0].targetPercent = null;
	assert.equal(validateRebalanceRequest(request({ bonds: null }), data), null);
});

test('estimates reflect adjusted stock and bond lots without enforcing the requested budget', async () => {
	const data = snapshots();
	const proposal = await requestRebalance(request(), data);
	const stock = proposal.trades.find((trade) => trade.asset === 'stocks');
	const bond = proposal.trades.find((trade) => trade.asset === 'bonds');
	proposal.trades = [stock, bond];
	stock.quantity = 3;
	bond.quantity = 2;
	assert.equal(stock.unitPrice, data.stocks.groups.flatMap((g) => g.stocks).find((s) => s.code === stock.code).price *
		data.stocks.groups.flatMap((g) => g.stocks).find((s) => s.code === stock.code).lotSize);
	assert.deepEqual(getRebalanceEstimate(proposal), { stocks: stock.unitPrice * 3, bonds: bond.unitPrice * 2, total: stock.unitPrice * 3 + bond.unitPrice * 2 });
	assert.equal(validateRebalanceProposal(proposal, data), null);
	const result = await applyRebalance(proposal, data);
	assert.equal(result.stocks.groups[0].stocks[0].lotCount, data.stocks.groups[0].stocks[0].lotCount + 3);
	assert.equal(result.bonds.groups[0].bonds[0].lotCount, data.bonds.groups[0].bonds[0].lotCount + 2);
});

test('foreign dirty bond prices are per lot and estimates are converted to RUB', async () => {
	const data = snapshots();
	data.bonds.groups[0].bonds = [sampleBonds[2], sampleBonds[5]].map((bond) => ({ ...bond, lotCount: 4, totalValueRub: null, targetPercent: null, actualPercent: null }));
	data.bonds.groups[1].bonds = [];
	const proposal = await requestRebalance(request({ stocks: null, noSell: true }), data);
	const [usd, cny] = proposal.trades;
	assert.deepEqual(usd.nativePrice, { currency: 'USD', price: 20.4 });
	assert.deepEqual(cny.nativePrice, { currency: 'CNY', price: 76 });
	assert.equal(usd.lotSize, 2);
	assert.equal(usd.unitPrice, 20.4 * 90);
	assert.equal(cny.unitPrice, 76 * 12);
	assert.equal(getRebalanceEstimate(proposal).total, 2 * (usd.unitPrice + cny.unitPrice));
	assert.equal(validateRebalanceProposal(proposal, data), null);
	const result = await applyRebalance(proposal, data);
	assert.equal(result.bonds.groups[0].bonds[0].lotCount, 6);
	assert.equal(result.bonds.groups[0].bonds[1].lotCount, 6);
});

test('missing exchange rates prevent mock estimates', async () => {
	const data = snapshots();
	data.bonds.groups[0].bonds[0].cleanPrice.currency = 'XYZ';
	data.bonds.groups[0].bonds[0].accruedInterest.currency = 'XYZ';
	assert.match(validateRebalanceRequest(request({ stocks: null }), data), /курсе валюты/);
	await assert.rejects(requestRebalance(request({ stocks: null }), data), /курсе валюты/);
});

test('mixed-currency dirty prices fall back to RUB and full bond lot sales remove holdings', async () => {
	const data = snapshots();
	const bond = data.bonds.groups[0].bonds[0];
	bond.cleanPrice = { currency: 'USD', price: 10 };
	bond.accruedInterest = { currency: 'CNY', price: 2 };
	bond.lotSize = 2;
	const proposal = await requestRebalance(request({ stocks: null, bonds: -1 }), data);
	const trade = proposal.trades[0];
	assert.deepEqual(trade.nativePrice, { currency: 'RUB', price: 1848 });
	trade.quantity = trade.availableQuantity;
	const result = await applyRebalance(proposal, data);
	assert.ok(!result.bonds.groups[0].bonds.some((item) => item.code === bond.code));
	assert.equal(data.bonds.groups[0].bonds[0].lotCount, bond.lotCount);
});

test('rejects stale snapshots, mismatched profiles, invalid trades and oversells', async () => {
	const data = snapshots();
	const original = await requestRebalance(request(), data);
	const stale = structuredClone(data);
	stale.stocks.groups[0].stocks[0].price++;
	assert.match(validateRebalanceProposal(original, stale), /Портфель изменился/);
	const wrongProfile = structuredClone(original);
	wrongProfile.profileId = 'other';
	assert.match(validateRebalanceProposal(wrongProfile, data), /профиль/);
	for (const quantity of [undefined, -1, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
		const proposal = structuredClone(original);
		proposal.trades[0].quantity = quantity;
		assert.match(validateRebalanceProposal(proposal, data), /Количество/);
	}
	const sell = await requestRebalance(request({ stocks: -1, bonds: null }), data);
	sell.trades[0].quantity = sell.trades[0].availableQuantity + 1;
	assert.match(validateRebalanceProposal(sell, data), /Нельзя продать/);
	const tampered = structuredClone(original);
	tampered.trades[0].unitPrice++;
	assert.match(validateRebalanceProposal(tampered, data), /Данные операции/);
	const duplicate = structuredClone(original);
	duplicate.trades.push({ ...duplicate.trades[0] });
	assert.match(validateRebalanceProposal(duplicate, data), /повторяющаяся/);
});

test('bond trades cannot produce an unsafe individual-bond quantity', async () => {
	const data = snapshots();
	const proposal = await requestRebalance(request({ stocks: null }), data);
	const trade = proposal.trades.find((item) => item.lotSize > 1);
	if (!trade) {
		const foreignGroup = data.bonds.groups.find((group) => group.bonds.some((bond) => bond.lotSize > 1));
		data.bonds.groups.unshift(data.bonds.groups.splice(data.bonds.groups.indexOf(foreignGroup), 1)[0]);
		const foreignProposal = await requestRebalance(request({ stocks: null }), data);
		const foreignTrade = foreignProposal.trades.find((item) => item.lotSize > 1 && item.side === 'buy');
		assert.ok(foreignTrade);
		foreignTrade.quantity = Number.MAX_SAFE_INTEGER - foreignTrade.availableQuantity;
		assert.match(validateRebalanceProposal(foreignProposal, data), /количество облигаций слишком велико/);
	} else {
		trade.quantity = Number.MAX_SAFE_INTEGER - trade.availableQuantity;
		assert.match(validateRebalanceProposal(proposal, data), /количество облигаций слишком велико/);
	}
});

test('zero trades are skipped, full sells remove holdings, and sender is awaited with isolated changes', async () => {
	const data = snapshots();
	const before = structuredClone(data);
	const proposal = await requestRebalance(request({ stocks: -1, bonds: -1 }), data);
	proposal.trades[0].quantity = proposal.trades[0].availableQuantity;
	proposal.trades[1].quantity = 0;
	let release;
	let payload;
	const wait = new Promise((resolve) => { release = resolve; });
	let finished = false;
	const applying = applyRebalance(proposal, data, async (value) => { payload = value; await wait; }).then((value) => { finished = true; return value; });
	await new Promise((resolve) => setTimeout(resolve, 0));
	assert.equal(finished, false);
	assert.equal(payload.profileId, 'main');
	assert.equal(payload.proposalId, proposal.id);
	assert.ok(payload.changes.every((trade) => trade.quantity > 0));
	payload.changes[0].quantity = 999;
	assert.notEqual(proposal.trades[0].quantity, 999);
	release();
	const result = await applying;
	assert.ok(!result.stocks.groups[0].stocks.some((stock) => stock.code === proposal.trades[0].code));
	assert.deepEqual(data, before);
	await sendRebalanceChanges({ profileId: 'main', proposalId: proposal.id, changes: [] });
});

test('failures leave inputs unchanged and unselected snapshots are cloned', async () => {
	const data = snapshots();
	const proposal = await requestRebalance(request({ bonds: null }), data);
	const before = structuredClone(data);
	const proposalBefore = structuredClone(proposal);
	await assert.rejects(applyRebalance(proposal, data, async () => { throw new Error('send failed'); }), /send failed/);
	assert.deepEqual(data, before);
	assert.deepEqual(proposal, proposalBefore);
	const result = await applyRebalance(proposal, data);
	result.bonds.groups[0].bonds[0].cleanPrice.price = -1;
	assert.deepEqual(data, before);
	assert.equal(getRebalanceErrorMessage(new Error('oops'), 'fallback'), 'oops');
	assert.equal(getRebalanceErrorMessage(null, 'fallback'), 'fallback');
});
