import type {
	DraftGroup, GroupWeight, PortfolioSnapshot, StockSearchResult
} from '#lib/logic/tabs/portfolio.js';
import { createDraftGroups, validateDraft, validateWeights } from '#lib/logic/tabs/stock-editor.js';

// Illustrative quotes only, not current market data.
const samples: StockSearchResult[] = [
	{ code: 'SBER', name: 'Сбербанк', price: 300, lotSize: 10 },
	{ code: 'VTBR', name: 'ВТБ', price: 100, lotSize: 10 },
	{ code: 'LKOH', name: 'Лукойл', price: 7000, lotSize: 1 },
	{ code: 'GAZP', name: 'Газпром', price: 140, lotSize: 10 },
	{ code: 'YDEX', name: 'Яндекс', price: 4000, lotSize: 1 },
	{ code: 'POSI', name: 'Группа Позитив', price: 1000, lotSize: 1 },
	{ code: 'FEES', name: 'Россети', price: 0.06, lotSize: 10000 },
	{ code: 'MRKC', name: 'РоссЦентр', price: 0.76, lotSize: 1000 },
	{ code: 'LSNGP', name: 'РСетиЛЭ-п', price: 365.75, lotSize: 10 },
	{ code: 'MRKV', name: 'РсетВол ао', price: 0.23, lotSize: 10000 },
	{ code: 'MRKP', name: 'РоссетиЦП ао', price: 0.60, lotSize: 1000 },
	{ code: 'MSRS', name: 'РСетиМР ао', price: 1.76, lotSize: 1000 },
	{ code: 'MRKZ', name: 'РСетиСЗ ао', price: 0.14, lotSize: 10000 },
	{ code: 'MRKU', name: 'Россети Ур', price: 0.67, lotSize: 1000 },
	{ code: 'LSNG', name: 'РСетиЛЭ', price: 14.50, lotSize: 100 },
	{ code: 'MRKS', name: 'РсетСиб ао', price: 0.49, lotSize: 1000 },
	{ code: 'MRKY', name: 'РоссЮг ао', price: 0.07, lotSize: 10000 },
	{ code: 'TORSP', name: 'РСТомск ап', price: 0.53, lotSize: 1000 },
	{ code: 'MRKK', name: 'Россети СК', price: 14.48, lotSize: 100 },
	{ code: 'TORS', name: 'РСТомск ао', price: 0.73, lotSize: 1000 }
];
const rossetiCodes = new Set(['FEES', 'MRKC', 'LSNGP', 'MRKV', 'MRKP', 'MSRS', 'MRKZ', 'MRKU', 'LSNG', 'MRKS', 'MRKY', 'TORSP', 'MRKK', 'TORS']);

export async function searchStocks(query: string): Promise<StockSearchResult[]> {
	const normalized = query.trim().toLowerCase();
	if (!normalized) return [];
	return samples.filter(({ code, name }) =>
		code.toLowerCase().includes(normalized) || name.toLowerCase().includes(normalized) ||
		(rossetiCodes.has(code) && 'россети'.includes(normalized))
	).map((stock) => ({ ...stock }));
}

export async function saveComposition(snapshot: PortfolioSnapshot, groups: DraftGroup[]): Promise<PortfolioSnapshot> {
	const error = validateDraft(groups);
	if (error) throw new Error(error);
	// Mock backend simulation only. Production valuations and allocation rules belong in Go,
	// not in components or browser-side portfolio business logic.
	const updated = groups.map((group) => ({
		id: group.id,
		name: group.name.trim(),
		targetPercent: group.targetPercent,
		totalValue: 0,
		actualPercent: null as number | null,
		stocks: group.stocks.map((stock) => ({
			...stock,
			code: stock.code.trim().toUpperCase(),
			name: stock.name.trim(),
			lotCount: stock.lotCount!,
			totalValue: stock.price * stock.lotSize * stock.lotCount!,
			targetPercent: group.targetPercent === null ? null : group.targetPercent / group.stocks.length,
			actualPercent: null as number | null
		}))
	}));
	let totalStocksValue = 0;
	for (const group of updated) {
		group.totalValue = group.stocks.reduce((total, stock) => total + stock.totalValue, 0);
		totalStocksValue += group.totalValue;
	}
	if (!Number.isFinite(totalStocksValue)) throw new Error('Общая стоимость акций слишком велика.');
	for (const group of updated) {
		group.actualPercent = totalStocksValue > 0 ? group.totalValue / totalStocksValue * 100 : null;
		for (const stock of group.stocks) {
			stock.actualPercent = totalStocksValue > 0 ? stock.totalValue / totalStocksValue * 100 : null;
		}
	}
	return {
		totalStocksValue,
		groups: updated,
		assetValues: { ...snapshot.assetValues, stocks: totalStocksValue }
	};
}

export async function saveWeights(snapshot: PortfolioSnapshot, weights: GroupWeight[]): Promise<PortfolioSnapshot> {
	const error = validateWeights(weights);
	if (error) throw new Error(error);
	const ids = new Set(snapshot.groups.map(({ id }) => id));
	if (ids.size !== snapshot.groups.length || weights.length !== ids.size || weights.some(({ id }) => !ids.has(id))) {
		throw new Error('Список долей должен точно соответствовать существующим группам.');
	}
	const values = new Map(weights.map(({ id, value }) => [id, value]));
	return saveComposition(snapshot, createDraftGroups(snapshot).map((group) => ({
		...group, targetPercent: values.get(group.id)!
	})));
}
