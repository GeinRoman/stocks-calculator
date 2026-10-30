export type Profile = {
	id: string;
	name: string;
};

export type StockSearchResult = {
	code: string;
	name: string;
	price: number;
	lotSize: number;
};

export type DraftStock = StockSearchResult & { lotCount: number | undefined };

export type DraftGroup = {
	id: string;
	name: string;
	targetPercent: number | null;
	stocks: DraftStock[];
};

export type GroupWeight = { id: string; value: number | null };

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
