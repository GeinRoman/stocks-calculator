import type { PersonalAssets } from './contracts.js';

export function personalAssetTotals(data: Pick<PersonalAssets, 'deposits' | 'insurances'>): { deposits: number | null; insurances: number | null } {
	function sum(assets: { amount_rub: number }[]): number | null {
		let total = 0;
		for (const asset of assets) {
			if (!Number.isFinite(asset.amount_rub) || asset.amount_rub < 0) return null;
			total += asset.amount_rub;
			if (!Number.isFinite(total)) return null;
		}
		return total;
	}
	return { deposits: sum(data.deposits), insurances: sum(data.insurances) };
}
