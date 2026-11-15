import type { portfolioApi } from '#lib/api/portfolio-api.js';
import type { Bond, BondOperation } from '#lib/api/contracts.js';
import { bondView } from '#lib/api/bond-view.js';
import { validateBondDraft } from '#lib/logic/tabs/bond-editor.js';
import { addDirtyPrice } from '#lib/logic/tabs/currency-service.js';
import type { DraftBondGroup, BondSearchResult } from '#lib/logic/tabs/bonds.js';
import type { PortfolioState } from './portfolio-state.js';

type BondsApi = ReturnType<typeof portfolioApi>['bonds'];

export async function searchLiveBonds(api: BondsApi, query: string): Promise<BondSearchResult[]> {
	const result = await api.find([query.trim()]);
	if (!Array.isArray(result) || result.length !== 1) throw new Error('Некорректный ответ поиска.');
	if (result[0].err_msg) throw new Error(result[0].err_msg);
	return (result[0].search_results ?? []).map(bond => ({ code: bond.code, name: bond.name, isin: bond.isin, cleanPrice: bond.clean_price, accruedInterest: bond.accrued_interest, maturityDate: bond.maturity_date || null, maturityPayment: bond.maturity_payment, lotSize: bond.lot_size }));
}

export class BondSaveError extends Error {
	constructor(message: string, public uncertain: boolean) { super(message); }
}

export async function saveLiveBonds(api: BondsApi, initial: PortfolioState, draft: DraftBondGroup[], commit: (state: PortfolioState) => void): Promise<void> {
	const invalid = validateBondDraft(draft);
	if (invalid) throw new Error(invalid);
	let state = initial;
	const desired = new Map(draft.map(group => [group.name, group]));
	function update(groups: typeof state.data.bond_groups, bonds: Bond[]) {
		const data = { ...state.data, bond_groups: groups, bonds };
		const unavailableRates = [...new Set(bonds.flatMap(bond => [bond.clean_price.currency, bond.accrued_interest.currency]).filter(currency => currency !== 'RUB' && !(Number.isFinite(state.rates[currency]) && state.rates[currency] > 0)))];
		state = { ...state, data, bonds: bondView(data, state.rates, state.loadedAt), unavailableRates };
		commit(state);
	}
	function operation(bond: Bond, amount: number): BondOperation {
		const price = addDirtyPrice(bond.clean_price, bond.accrued_interest, state.rates);
		if (!price || price.price <= 0) throw new Error(`Нет корректной цены с НКД или курса для ${bond.code}. Ранее подтверждённые изменения сохранены; остальные остались в черновике.`);
		return { code: bond.code, name: bond.name, group_name: bond.group_name, lot_size: bond.lot_size, lot_amount: amount, price };
	}
	async function request(label: string, action: () => Promise<void>) {
		try { await action(); }
		catch (error) {
			const status = typeof error === 'object' && error !== null && 'status' in error ? Number(error.status) : 0;
			const uncertain = !status || status >= 500;
			throw new BondSaveError(`Не удалось выполнить: ${label}. Подтверждённые ранее изменения сохранены. ${uncertain ? 'Результат запроса неизвестен: отмените редактирование и обновите страницу перед повтором.' : 'Остальные изменения остались в черновике.'}`, uncertain);
		}
	}
	for (const bond of [...state.data.bonds]) {
		const target = desired.get(bond.group_name)?.bonds.find(item => item.code === bond.code)?.lotCount ?? 0;
		if (target >= bond.lot_amount) continue;
		const sale = operation(bond, bond.lot_amount - target);
		await request(`продажа ${bond.code}`, () => api.remove([sale]));
		update(state.data.bond_groups, state.data.bonds.flatMap(item => item.code !== bond.code ? [item] : target ? [{ ...item, lot_amount: target }] : []));
	}
	for (const group of [...state.data.bond_groups]) {
		if (desired.has(group.name)) continue;
		await request(`удаление группы ${group.name}`, () => api.removeGroups([group]));
		update(state.data.bond_groups.filter(item => item.name !== group.name), state.data.bonds);
	}
	for (const group of draft) {
		if (state.data.bond_groups.some(item => item.name === group.name)) continue;
		const added = { name: group.name, weight: 0 };
		await request(`создание группы ${group.name}`, () => api.addGroups([added]));
		update([...state.data.bond_groups, added], state.data.bonds);
	}
	for (const group of draft) for (const item of group.bonds) {
		const existing = state.data.bonds.find(bond => bond.code === item.code);
		const target = item.lotCount!, delta = target - (existing?.lot_amount ?? 0);
		if (delta <= 0) continue;
		const bond: Bond = existing ?? { code: item.code, name: item.name, isin: item.isin, group_name: group.name, clean_price: item.cleanPrice, accrued_interest: item.accruedInterest, maturity_date: item.maturityDate ?? '', maturity_payment: item.maturityPayment, lot_size: item.lotSize, lot_amount: 0 };
		const purchase = operation(bond, delta);
		await request(`покупка ${item.code}`, () => api.add([purchase]));
		update(state.data.bond_groups, existing ? state.data.bonds.map(value => value.code === bond.code ? { ...value, lot_amount: target } : value) : [...state.data.bonds, { ...bond, lot_amount: target }]);
	}
}
