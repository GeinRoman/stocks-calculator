export type Profile = {
	id: string;
	name: string;
};

export type StockHolding = {
	code: string;
	name: string;
	price: number;
	lotCount: number;
	lotSize: number;
	totalValue: number;
	targetPercent?: number | null;
	actualPercent?: number | null;
};

export type StockGroup = {
	id: string;
	name: string;
	totalValue: number;
	targetPercent?: number | null;
	actualPercent?: number | null;
	stocks: StockHolding[];
};

// Display-ready valuations and allocations, to be supplied by the backend.
export type PortfolioSnapshot = {
	totalStocksValue: number;
	groups: StockGroup[];
	assetValues: Record<'stocks' | 'bonds' | 'deposits', number>;
};
