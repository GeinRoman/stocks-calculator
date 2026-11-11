import type { RebalanceProposal, RebalanceRequest, RebalanceSnapshots, RebalanceTrade } from '#lib/logic/tabs/rebalance.js';
import { validateRebalanceProposal, validateRebalanceRequest } from '#lib/logic/tabs/rebalance-validation.js';
import { createDraftGroups } from '#lib/logic/tabs/stock-editor.js';
import { createBondDraftGroups } from '#lib/logic/tabs/bond-editor.js';
import { saveComposition } from '#lib/logic/tabs/portfolio-service.js';
import { saveBondComposition } from '#lib/logic/tabs/bond-service.js';
import { convertMoney, multiplyMoney, quoteBond } from '#lib/logic/tabs/currency-service.js';

// TEMP
export async function requestRebalance(request: RebalanceRequest, snapshots: RebalanceSnapshots): Promise<RebalanceProposal> {
	const error = validateRebalanceRequest(request, snapshots);
	if (error) throw new Error(error);
	const trades: RebalanceTrade[] = [];
	// Demonstration deltas only. The backend will provide actual allocation proposals.
	for (const asset of ['stocks', 'bonds'] as const) {
		const amount = request[asset];
		if (amount === null) continue;
		const holdings = asset === 'stocks'
			? snapshots.stocks.groups.flatMap((group) => group.stocks.map((stock) => ({
				groupId: group.id, code: stock.code, name: stock.name,
				unitPrice: stock.price * stock.lotSize, lotSize: stock.lotSize, availableQuantity: stock.lotCount
			})))
			: snapshots.bonds.groups.flatMap((group) => group.bonds.map((bond) => {
				const nativePrice = multiplyMoney(quoteBond(bond, true, false), bond.lotSize);
				const rubPrice = convertMoney(nativePrice, 'RUB');
				if (!rubPrice) throw new Error('Нет данных о цене или курсе валюты облигации для оценки в рублях.');
				return { groupId: group.id, code: bond.code, name: bond.name, nativePrice,
					unitPrice: rubPrice.price, lotSize: bond.lotSize, availableQuantity: bond.lotCount };
			}));
		for (const [index, holding] of holdings.slice(0, 2).entries()) {
			const side = amount < 0 || (!request.noSell && index === 1) ? 'sell' : 'buy';
			const quantity = side === 'buy' ? 2 : Math.min(amount < 0 ? 2 : 1, holding.availableQuantity);
			trades.push({ ...holding, id: JSON.stringify([asset, holding.groupId, holding.code, side]), asset, side, quantity });
		}
	}
	return {
		id: globalThis.crypto.randomUUID(), profileId: request.profileId, request: { ...request },
		createdAt: new Date().toISOString(),
		baseline: {
			stocks: request.stocks === null ? null : JSON.stringify(snapshots.stocks),
			bonds: request.bonds === null ? null : JSON.stringify(snapshots.bonds)
		},
		trades
	};
}

// TEMP
export async function sendRebalanceChanges(_payload: {
	profileId: string; proposalId: string; changes: RebalanceTrade[];
}): Promise<void> {
	// Future HTTP persistence hook: send these deltas atomically to the backend.
}

// TEMP
export async function applyRebalance(
	proposal: RebalanceProposal, snapshots: RebalanceSnapshots,
	sender: typeof sendRebalanceChanges = sendRebalanceChanges
): Promise<RebalanceSnapshots> {
	const error = validateRebalanceProposal(proposal, snapshots);
	if (error) throw new Error(error);
	const changes = proposal.trades.filter(({ quantity }) => quantity! > 0).map((trade) => ({ ...trade }));
	const profileId = proposal.profileId;
	const proposalId = proposal.id;
	const saveStocks = proposal.request.stocks !== null;
	const saveBonds = proposal.request.bonds !== null;
	const next: RebalanceSnapshots = {
		stocks: {
			...snapshots.stocks, assetValues: { ...snapshots.stocks.assetValues },
			groups: snapshots.stocks.groups.map((group) => ({
				...group, stocks: group.stocks.map((stock) => ({ ...stock }))
			}))
		},
		bonds: JSON.parse(JSON.stringify(snapshots.bonds))
	};
	// Build isolated drafts before any await.
	const stocks = createDraftGroups(snapshots.stocks);
	const bonds = createBondDraftGroups(snapshots.bonds);
	for (const trade of changes) {
		const delta = trade.quantity! * (trade.side === 'buy' ? 1 : -1);
		if (trade.asset === 'stocks') {
			const holding = stocks.find(({ id }) => id === trade.groupId)!.stocks.find(({ code }) => code === trade.code)!;
			holding.lotCount! += delta;
		} else {
			const holding = bonds.find(({ id }) => id === trade.groupId)!.bonds.find(({ code }) => code === trade.code)!;
			holding.lotCount! += delta;
		}
	}
	if (saveStocks) {
		next.stocks = await saveComposition(next.stocks, stocks.map((group) => ({
			...group, stocks: group.stocks.filter(({ lotCount }) => lotCount !== 0)
		})));
	}
	if (saveBonds) {
		next.bonds = await saveBondComposition(next.bonds, bonds.map((group) => ({
			...group, bonds: group.bonds.filter(({ lotCount }) => lotCount !== 0)
		})));
	}
	await sender({ profileId, proposalId, changes });
	return next;
}
