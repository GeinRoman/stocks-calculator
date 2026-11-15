import type { AllInfo } from './contracts.js';
import type { StockGroup } from '../logic/tabs/portfolio.js';

/** Derive display values from the prices and integer lots supplied by /getall. */
export function stockView(data: Pick<AllInfo, 'stocks' | 'stock_groups'>): { total: number; groups: StockGroup[] } {
	const names = new Set(data.stock_groups.map(group => group.name));
	if (names.size !== data.stock_groups.length || data.stocks.some(stock => !names.has(stock.group_name))) throw new Error('Stock groups do not match holdings');
	const total = data.stocks.reduce((sum, stock) => sum + stock.price * stock.lot_amount * stock.lot_size, 0);
	if (!Number.isFinite(total)) throw new Error('Stock portfolio value is not finite');
	const groups = data.stock_groups.map(group => {
		const holdings = data.stocks.filter(stock => stock.group_name === group.name);
		const groupTotal = holdings.reduce((sum, stock) => sum + stock.price * stock.lot_amount * stock.lot_size, 0);
		return {
			id: group.name,
			name: group.name,
			totalValue: groupTotal,
			targetPercent: group.weight,
			actualPercent: total > 0 ? groupTotal / total * 100 : null,
			stocks: holdings.map(stock => {
				const value = stock.price * stock.lot_amount * stock.lot_size;
				return {
					code: stock.code, name: stock.name, price: stock.price,
					lotCount: stock.lot_amount, lotSize: stock.lot_size, totalValue: value,
					targetPercent: holdings.length ? group.weight / holdings.length : null,
					actualPercent: total > 0 ? value / total * 100 : null
				};
			})
		};
	});
	return { total, groups };
}
