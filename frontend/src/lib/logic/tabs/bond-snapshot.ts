import type { BondPortfolioSnapshot, DraftBondGroup } from '#lib/logic/tabs/bonds.js';
import { validateBondDraft } from '#lib/logic/tabs/bond-editor.js';

// Mock backend emulation for fixtures/services only. Future Go responses are authoritative;
// components must not calculate financial values or treat demo prices as live quotes.
export function computeBondSnapshot(groups: DraftBondGroup[], asOf = new Date()): BondPortfolioSnapshot {
	if (!Number.isFinite(asOf.getTime())) throw new Error('Некорректная дата оценки облигаций.');
	const error = validateBondDraft(groups);
	if (error) throw new Error(error);
	const updated = groups.map((group) => ({
		id: group.id, name: group.name.trim(), targetPercent: group.targetPercent,
		totalValue: 0, actualPercent: null as number | null,
		bonds: group.bonds.map((bond) => {
			const schedule = bond.couponSchedule?.map((payment) => ({ ...payment })) ?? null;
			const future = schedule?.filter(({ date }) => Date.parse(date) > asOf.getTime()) ?? null;
			const remainingCouponValue = future === null || future.some(({ amount }) => amount === null)
				? null : future.reduce((sum, { amount }) => sum + amount!, 0);
			return {
				...bond, code: bond.code.trim().toUpperCase(), name: bond.name.trim(),
				couponSchedule: schedule, quantity: bond.quantity!, totalValue: bond.price * bond.quantity!,
				remainingCoupons: future?.length ?? null, totalCoupons: schedule?.length ?? null,
				remainingCouponValue,
				holdingRemainingCouponValue: remainingCouponValue === null ? null : remainingCouponValue * bond.quantity!,
				targetPercent: group.targetPercent === null ? null : group.targetPercent / group.bonds.length,
				actualPercent: null as number | null
			};
		})
	}));
	let totalBondsValue = 0;
	for (const group of updated) {
		group.totalValue = group.bonds.reduce((sum, bond) => sum + bond.totalValue, 0);
		totalBondsValue += group.totalValue;
	}
	for (const group of updated) {
		group.actualPercent = totalBondsValue > 0 ? group.totalValue / totalBondsValue * 100 : null;
		for (const bond of group.bonds) bond.actualPercent = totalBondsValue > 0 ? bond.totalValue / totalBondsValue * 100 : null;
	}
	return { totalBondsValue, groups: updated, asOf: asOf.toISOString() };
}
