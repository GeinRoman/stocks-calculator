import type { portfolioApi } from '#lib/api/portfolio-api.js';
import type { Stock } from '#lib/api/contracts.js';
import { stockView } from '#lib/api/stock-view.js';
import { validateDraft } from '#lib/logic/tabs/stock-editor.js';
import type { DraftGroup, StockSearchResult } from '#lib/logic/tabs/portfolio.js';
import type { PortfolioState } from './portfolio-state.js';

type StocksApi = ReturnType<typeof portfolioApi>['stocks'];

export async function searchLiveStocks(api: StocksApi, query: string): Promise<StockSearchResult[]> {
	const result = await api.find([query.trim()]);
	if (!Array.isArray(result) || result.length !== 1) throw new Error('Некорректный ответ поиска.');
	if (result[0].err_msg) throw new Error(result[0].err_msg);
	return (result[0].search_results ?? []).map(stock => ({ code: stock.code, name: stock.name, price: stock.price, lotSize: stock.lot_size }));
}

export class StockSaveError extends Error {
	constructor(message: string, public uncertain: boolean) { super(message); }
}

/** Diff against the latest confirmed snapshot on every attempt. */
export async function saveLiveStocks(api: StocksApi, initial: PortfolioState, draft: DraftGroup[], commit: (state: PortfolioState) => void): Promise<void> {
	const invalid = validateDraft(draft);
	if (invalid) throw new Error(invalid);
	let state = initial;
	const desired = new Map(draft.map(group => [group.name, group]));
	function update(groups: typeof state.data.stock_groups, stocks: Stock[]) {
		const data = { ...state.data, stock_groups: groups, stocks };
		state = { ...state, data, stocks: stockView(data) };
		commit(state);
	}
	async function request(label: string, action: () => Promise<void>) {
		try { await action(); }
		catch (error) {
			const status = typeof error === 'object' && error !== null && 'status' in error ? Number(error.status) : 0;
			const uncertain = !status || status >= 500;
			throw new StockSaveError(`Не удалось выполнить: ${label}. Подтверждённые ранее изменения сохранены. ${uncertain ? 'Результат запроса неизвестен: отмените редактирование и обновите страницу перед повтором.' : 'Остальные изменения остались в черновике.'}`, uncertain);
		}
	}
	// Sales precede group deletion and purchases, preserving the transaction history.
	for (const stock of [...state.data.stocks]) {
		const target = desired.get(stock.group_name)?.stocks.find(item => item.code === stock.code)?.lotCount ?? 0;
		if (target >= stock.lot_amount) continue;
		await request(`продажа ${stock.code}`, () => api.remove([{ ...stock, lot_amount: stock.lot_amount - target }]));
		update(state.data.stock_groups, state.data.stocks.flatMap(item => item.code !== stock.code ? [item] : target ? [{ ...item, lot_amount: target }] : []));
	}
	for (const group of [...state.data.stock_groups]) {
		if (desired.has(group.name)) continue;
		await request(`удаление группы ${group.name}`, () => api.removeGroups([group]));
		update(state.data.stock_groups.filter(item => item.name !== group.name), state.data.stocks);
	}
	for (const group of draft) {
		if (state.data.stock_groups.some(item => item.name === group.name)) continue;
		const added = { name: group.name, weight: 0 };
		await request(`создание группы ${group.name}`, () => api.addGroups([added]));
		update([...state.data.stock_groups, added], state.data.stocks);
	}
	for (const group of draft) for (const item of group.stocks) {
		const existing = state.data.stocks.find(stock => stock.code === item.code);
		const target = item.lotCount!;
		const delta = target - (existing?.lot_amount ?? 0);
		if (delta <= 0) continue;
		const stock: Stock = existing ?? { code: item.code, name: item.name, group_name: group.name, price: item.price, lot_size: item.lotSize, lot_amount: 0 };
		await request(`покупка ${item.code}`, () => api.add([{ ...stock, lot_amount: delta }]));
		update(state.data.stock_groups, existing ? state.data.stocks.map(value => value.code === stock.code ? { ...value, lot_amount: target } : value) : [...state.data.stocks, { ...stock, lot_amount: target }]);
	}
}
