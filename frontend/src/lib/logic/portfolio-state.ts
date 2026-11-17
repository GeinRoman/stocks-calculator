import type { AllInfo } from '#lib/api/contracts.js';
import type { portfolioApi } from '#lib/api/portfolio-api.js';
import { loadBondRates } from '#lib/api/bond-rates.js';
import { bondView } from '#lib/api/bond-view.js';
import { stockView } from '#lib/api/stock-view.js';
import { personalAssetTotals } from '#lib/api/personal-assets-view.js';

type PortfolioApi = Pick<ReturnType<typeof portfolioApi>, 'getAll' | 'currencyRate'>;

export async function loadPortfolioState(api: PortfolioApi) {
	const data: AllInfo = await api.getAll();
	const asOf = new Date().toISOString();
	const { rates, unavailable } = await loadBondRates(data.bonds, (from, to) => api.currencyRate(from, to));
	return {
		// Preserve the backend snapshot, including profile, group, weight and asset fields.
		data,
		stocks: stockView(data),
		bonds: bondView(data, rates, asOf),
		personalTotals: personalAssetTotals(data.personal_assets),
		rates,
		unavailableRates: unavailable,
		loadedAt: asOf
	};
}

export type PortfolioState = Awaited<ReturnType<typeof loadPortfolioState>>;
