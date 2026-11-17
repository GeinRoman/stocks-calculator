import type { Bond, BondOperation, Stock } from '#lib/api/contracts.js';
import type { portfolioApi } from '#lib/api/portfolio-api.js';

type Api = Pick<ReturnType<typeof portfolioApi>, 'stocks' | 'bonds'>;
export type LiveTrade = (
	| { kind: 'stock'; holding: Stock; target: number }
	| { kind: 'bond'; holding: Bond; target: number }) & { uncertain?: boolean };

/** Approximate RUB value of a lot using the loaded portfolio quotes and rates. */
export function rebalanceLotValue(trade: LiveTrade, rates: Readonly<Record<string, number>>): number | null {
	const holding = trade.holding;
	if (!Number.isSafeInteger(holding.lot_size) || holding.lot_size <= 0) return null;
	if (trade.kind === 'stock') {
		const value = trade.holding.price * holding.lot_size;
		return Number.isFinite(value) && value >= 0 ? value : null;
	}
	const rub = (price: { currency: string; price: number }) => {
		const rate = price.currency === 'RUB' ? 1 : rates[price.currency];
		return Number.isFinite(price.price) && price.price >= 0 && Number.isFinite(rate) && rate > 0 ? price.price * rate : NaN;
	};
	const value = (rub(trade.holding.clean_price) + rub(trade.holding.accrued_interest)) * holding.lot_size;
	return Number.isFinite(value) && value >= 0 ? value : null;
}

export function rebalanceTotals(trades: LiveTrade[], rates: Readonly<Record<string, number>>): { stocks: number | null; bonds: number | null; total: number | null } {
	let stocks: number | null = 0, bonds: number | null = 0;
	for (const trade of trades) {
		const delta = trade.target - trade.holding.lot_amount;
		if (!Number.isSafeInteger(trade.target) || trade.target < 0 || !Number.isSafeInteger(delta)) return { stocks: null, bonds: null, total: null };
		if (!delta) continue;
		const price = rebalanceLotValue(trade, rates);
		if (trade.kind === 'stock') stocks = stocks === null || price === null ? null : stocks + price * delta;
		else bonds = bonds === null || price === null ? null : bonds + price * delta;
	}
	if (stocks !== null && !Number.isFinite(stocks)) stocks = null;
	if (bonds !== null && !Number.isFinite(bonds)) bonds = null;
	const sum = stocks === null || bonds === null ? null : stocks + bonds;
	return { stocks, bonds, total: sum !== null && Number.isFinite(sum) ? sum : null };
}

/** Server proposals contain final amounts, not transaction deltas. */
export async function requestLiveRebalance(api: Api, holdings: { stocks: Stock[]; bonds: Bond[] }, request: { stocks: number | null; bonds: number | null; noSell: boolean }): Promise<LiveTrade[]> {
	if (request.stocks === null && request.bonds === null) throw new Error('Выберите хотя бы один вид активов.');
	for (const amount of [request.stocks, request.bonds]) {
		if (amount !== null && (!Number.isFinite(amount) || (request.noSell && amount <= 0))) throw new Error('Укажите корректную сумму.');
	}
	const [stocks, bonds] = await Promise.all([
		request.stocks === null ? Promise.resolve(null) : api.stocks.rebalance(request.noSell, request.stocks),
		request.bonds === null ? Promise.resolve(null) : api.bonds.rebalance(request.noSell, request.bonds)
	]);
	function map<T extends Stock | Bond>(current: T[], proposed: T[]): { holding: T; target: number }[] {
		if (!Array.isArray(proposed) || proposed.length !== current.length) throw new Error('Предложение не соответствует портфелю. Обновите данные.');
		const seen = new Set<string>();
		return proposed.map(item => {
			const existing = current.find(value => value.code === item.code && value.group_name === item.group_name);
			if (!existing || seen.has(item.code) || existing.lot_size !== item.lot_size || !Number.isSafeInteger(item.lot_amount) || item.lot_amount < 0 || (request.noSell && item.lot_amount < existing.lot_amount)) throw new Error('Предложение не соответствует портфелю. Обновите данные.');
			seen.add(item.code);
			return { holding: existing, target: item.lot_amount };
		});
	}
	return [
		...(stocks ? map(holdings.stocks, stocks.stocks).map(value => ({ kind: 'stock' as const, ...value })) : []),
		...(bonds ? map(holdings.bonds, bonds.bonds).map(value => ({ kind: 'bond' as const, ...value })) : [])
	];
}

/** Apply one transaction at a time. The callback must synchronously commit each confirmed result. */
export async function applyLiveRebalance(api: Api, trades: LiveTrade[], rates: Readonly<Record<string, number>>, commit: (trade: LiveTrade) => void): Promise<void> {
	for (const trade of trades) {
		if (trade.uncertain) throw new Error('Результат предыдущего запроса неизвестен. Обновите портфель и запросите новое предложение.');
		const { holding, target } = trade;
		if (!Number.isSafeInteger(target) || target < 0 || !Number.isSafeInteger(holding.lot_amount) || holding.lot_amount < 0) throw new Error('Некорректное количество лотов.');
		const delta = target - holding.lot_amount;
		if (!delta) continue;
		try {
			if (trade.kind === 'stock') {
				const operation = { ...trade.holding, lot_amount: Math.abs(delta) };
				await (delta > 0 ? api.stocks.add([operation]) : api.stocks.remove([operation]));
			} else {
				const bond = trade.holding;
				const rate = (currency: string) => currency === 'RUB' ? 1 : rates[currency];
				const cleanRate = rate(bond.clean_price.currency), accruedRate = rate(bond.accrued_interest.currency);
				const price = bond.clean_price.currency === bond.accrued_interest.currency
					? { currency: bond.clean_price.currency, price: bond.clean_price.price + bond.accrued_interest.price }
					: { currency: 'RUB', price: bond.clean_price.price * cleanRate + bond.accrued_interest.price * accruedRate };
				if (!Number.isFinite(price.price) || price.price <= 0) throw new Error('Нет корректной цены облигации.');
				const operation: BondOperation = { name: bond.name, code: bond.code, group_name: bond.group_name, lot_size: bond.lot_size, lot_amount: Math.abs(delta), price };
				await (delta > 0 ? api.bonds.add([operation]) : api.bonds.remove([operation]));
			}
		} catch (cause) {
			trade.uncertain = true;
			throw new Error(`Не удалось подтвердить операцию ${holding.code}. Предыдущие подтверждённые операции сохранены. Перед повтором проверьте состояние портфеля: результат последнего запроса может быть неизвестен.`, { cause });
		}
		// A retry must not repeat a transaction that returned HTTP 200.
		trade.holding = { ...trade.holding, lot_amount: target };
		commit(trade);
	}
}
