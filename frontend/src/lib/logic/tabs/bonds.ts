export type CouponPayment = { date: string; amount: number | null };

export type BondSearchResult = {
	code: string;
	name: string;
	price: number;
	redemptionPrice: number | null;
	maturityDate: string | null;
	couponAmount: number | null;
	couponSchedule: CouponPayment[] | null;
};

export type DraftBond = BondSearchResult & { quantity: number | undefined };

export type DraftBondGroup = {
	id: string;
	name: string;
	targetPercent: number | null;
	bonds: DraftBond[];
};

export type BondHolding = BondSearchResult & {
	quantity: number;
	totalValue: number;
	remainingCoupons: number | null;
	totalCoupons: number | null;
	remainingCouponValue: number | null;
	holdingRemainingCouponValue: number | null;
	targetPercent: number | null;
	actualPercent: number | null;
};

export type BondGroup = {
	id: string;
	name: string;
	totalValue: number;
	targetPercent: number | null;
	actualPercent: number | null;
	bonds: BondHolding[];
};

export type BondPortfolioSnapshot = {
	totalBondsValue: number;
	groups: BondGroup[];
	asOf: string;
};
