import type { Money } from './currency.js';

export type BondSearchResult = {
	code: string; name: string; isin: string;
	cleanPrice: Money; accruedInterest: Money;
	maturityDate: string | null; maturityPayment: Money | null; lotSize: number;
};
export type DraftBond = BondSearchResult & { lotCount: number | undefined };
export type DraftBondGroup = { id: string; name: string; targetPercent: number | null; bonds: DraftBond[] };
export type BondHolding = BondSearchResult & {
	lotCount: number; totalValueRub: number | null; targetPercent: number | null; actualPercent: number | null;
};
export type BondGroup = {
	id: string; name: string; totalValueRub: number | null; targetPercent: number | null;
	actualPercent: number | null; bonds: BondHolding[];
};
export type BondPortfolioSnapshot = { totalBondsValueRub: number | null; groups: BondGroup[]; asOf: string };
