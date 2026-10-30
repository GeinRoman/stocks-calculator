import type { BondPortfolioSnapshot, BondSearchResult, DraftBondGroup } from '#lib/logic/tabs/bonds.js';
import type { GroupWeight } from '#lib/logic/tabs/portfolio.js';
import { createBondDraftGroups } from '#lib/logic/tabs/bond-editor.js';
import { computeBondSnapshot } from '#lib/logic/tabs/bond-snapshot.js';
import { sampleBonds } from '#lib/logic/tabs/demo-bonds.js';
import { validateWeights } from '#lib/logic/tabs/stock-editor.js';

export async function searchBonds(query: string): Promise<BondSearchResult[]> {
	const normalized = query.trim().toLowerCase();
	if (!normalized) return [];
	return sampleBonds.filter(({ code, name }) => code.toLowerCase().includes(normalized) || name.toLowerCase().includes(normalized))
		.map((bond) => ({ ...bond, couponSchedule: bond.couponSchedule?.map((payment) => ({ ...payment })) ?? null }));
}

export async function saveBondComposition(
	_snapshot: BondPortfolioSnapshot, groups: DraftBondGroup[], asOf = new Date()
): Promise<BondPortfolioSnapshot> {
	return computeBondSnapshot(groups, asOf);
}

export async function saveBondWeights(
	snapshot: BondPortfolioSnapshot, weights: GroupWeight[], asOf = new Date()
): Promise<BondPortfolioSnapshot> {
	const error = validateWeights(weights, true);
	if (error) throw new Error(error);
	const ids = new Set(snapshot.groups.map(({ id }) => id));
	if (ids.size !== snapshot.groups.length || weights.length !== ids.size || weights.some(({ id }) => !ids.has(id))) {
		throw new Error('Список долей должен точно соответствовать существующим группам.');
	}
	const values = new Map(weights.map(({ id, value }) => [id, value]));
	return saveBondComposition(snapshot, createBondDraftGroups(snapshot).map((group) => ({
		...group, targetPercent: values.get(group.id)!
	})), asOf);
}
