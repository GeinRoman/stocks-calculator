import type { PortfolioSnapshot } from '#lib/logic/tabs/portfolio.js';
import type { BondPortfolioSnapshot } from '#lib/logic/tabs/bonds.js';
import type { Money } from '#lib/logic/tabs/currency.js';

export type RebalanceRequest = {
	profileId: string;
	stocks: number | null;
	bonds: number | null;
	noSell: boolean;
};

export type RebalanceTrade = {
	id: string;
	asset: 'stocks' | 'bonds';
	side: 'buy' | 'sell';
	groupId: string;
	code: string;
	name: string;
	unitPrice: number;
	nativePrice?: Money | null;
	lotSize: number;
	availableQuantity: number;
	quantity: number | undefined;
};

export type RebalanceProposal = {
	id: string;
	profileId: string;
	request: RebalanceRequest;
	createdAt: string;
	baseline: { stocks: string | null; bonds: string | null };
	trades: RebalanceTrade[];
};

export type RebalanceSnapshots = {
	stocks: PortfolioSnapshot;
	bonds: BondPortfolioSnapshot;
};
