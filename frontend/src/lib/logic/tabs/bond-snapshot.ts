import type { BondPortfolioSnapshot, DraftBondGroup } from './bonds.js';
import { validateBondDraft } from './bond-editor.js';
import { addDirtyPrice, convertMoney, multiplyMoney } from './currency-service.js';

// TEMP
export function computeBondSnapshot(groups: DraftBondGroup[], asOf = new Date()): BondPortfolioSnapshot {
	if (!Number.isFinite(asOf.getTime())) throw new Error('Некорректная дата оценки облигаций.');
	const error = validateBondDraft(groups);
	if (error) throw new Error(error);
	const updated = groups.map((group) => {
		const bonds = group.bonds.map((bond) => {
			const value = multiplyMoney(convertMoney(addDirtyPrice(bond.cleanPrice, bond.accruedInterest), 'RUB'), bond.lotSize * bond.lotCount!);
			return { code: bond.code.trim().toUpperCase(), name: bond.name.trim(), isin: bond.isin.trim(), cleanPrice: { ...bond.cleanPrice }, accruedInterest: { ...bond.accruedInterest }, maturityDate: bond.maturityDate, maturityPayment: bond.maturityPayment && { ...bond.maturityPayment }, lotSize: bond.lotSize, lotCount: bond.lotCount!, totalValueRub: value?.price ?? null, targetPercent: group.targetPercent === null ? null : group.targetPercent / group.bonds.length, actualPercent: null as number | null };
		});
		const sum = bonds.reduce<number | null>((total, bond) => total === null || bond.totalValueRub === null ? null : total + bond.totalValueRub, 0);
		return { id: group.id, name: group.name.trim(), targetPercent: group.targetPercent, totalValueRub: sum !== null && Number.isFinite(sum) ? sum : null, actualPercent: null as number | null, bonds };
	});
	const sum = updated.reduce<number | null>((total, group) => total === null || group.totalValueRub === null ? null : total + group.totalValueRub, 0);
	const totalBondsValueRub = sum !== null && Number.isFinite(sum) ? sum : null;
	for (const group of updated) {
		group.actualPercent = totalBondsValueRub !== null && totalBondsValueRub > 0 && group.totalValueRub !== null ? group.totalValueRub / totalBondsValueRub * 100 : null;
		for (const bond of group.bonds) bond.actualPercent = totalBondsValueRub !== null && totalBondsValueRub > 0 && bond.totalValueRub !== null ? bond.totalValueRub / totalBondsValueRub * 100 : null;
	}
	return { totalBondsValueRub, groups: updated, asOf: asOf.toISOString() };
}
