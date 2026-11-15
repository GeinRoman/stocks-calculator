import type { AllInfo, Money } from './contracts.js';
import type { BondPortfolioSnapshot } from '../logic/tabs/bonds.js';

/** Rates are RUB per unit, obtained from the backend, not demo exchange rates. */
export function bondView(data: Pick<AllInfo, 'bonds' | 'bond_groups'>, rates: Readonly<Record<string, number>>, asOf: string): BondPortfolioSnapshot {
	if (!Number.isFinite(Date.parse(asOf))) throw new Error('Invalid bond snapshot time');
	const names = new Set(data.bond_groups.map(group => group.name));
	if (names.size !== data.bond_groups.length || data.bonds.some(bond => !names.has(bond.group_name))) throw new Error('Bond groups do not match holdings');
	function rub(money: Money): number | null {
		const rate = money.currency === 'RUB' ? 1 : Object.hasOwn(rates, money.currency) ? rates[money.currency] : undefined;
		if (rate === undefined || !Number.isFinite(rate) || rate <= 0 || !Number.isFinite(money.price) || money.price < 0) return null;
		const value = money.price * rate;
		return Number.isFinite(value) ? value : null;
	}
	const groups = data.bond_groups.map(group => {
		const holdings = data.bonds.filter(bond => bond.group_name === group.name);
		const bonds = holdings.map(bond => {
			const clean = rub(bond.clean_price), accrued = rub(bond.accrued_interest);
			const value = clean === null || accrued === null ? null : (clean + accrued) * bond.lot_amount * bond.lot_size;
			return {
				code: bond.code, name: bond.name, isin: bond.isin,
				cleanPrice: { ...bond.clean_price }, accruedInterest: { ...bond.accrued_interest },
				maturityDate: bond.maturity_date || null, maturityPayment: bond.maturity_payment && { ...bond.maturity_payment },
				lotSize: bond.lot_size, lotCount: bond.lot_amount,
				totalValueRub: value !== null && Number.isFinite(value) ? value : null,
				targetPercent: holdings.length ? group.weight / holdings.length : null,
				actualPercent: null as number | null
			};
		});
		const sum = bonds.reduce<number | null>((total, bond) => total === null || bond.totalValueRub === null ? null : total + bond.totalValueRub, 0);
		return { id: group.name, name: group.name, targetPercent: group.weight, totalValueRub: sum !== null && Number.isFinite(sum) ? sum : null, actualPercent: null as number | null, bonds };
	});
	const sum = groups.reduce<number | null>((total, group) => total === null || group.totalValueRub === null ? null : total + group.totalValueRub, 0);
	const totalBondsValueRub = sum !== null && Number.isFinite(sum) ? sum : null;
	for (const group of groups) {
		group.actualPercent = totalBondsValueRub !== null && totalBondsValueRub > 0 && group.totalValueRub !== null ? group.totalValueRub / totalBondsValueRub * 100 : null;
		for (const bond of group.bonds) bond.actualPercent = totalBondsValueRub !== null && totalBondsValueRub > 0 && bond.totalValueRub !== null ? bond.totalValueRub / totalBondsValueRub * 100 : null;
	}
	return { totalBondsValueRub, groups, asOf };
}
