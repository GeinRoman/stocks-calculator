import type { portfolioApi } from './portfolio-api.js';
import type { HistoryCursor, HistoryKind, HistoryTransaction } from '../logic/tabs/transaction-history.js';

type PortfolioApi = ReturnType<typeof portfolioApi>;

/** History is loaded on demand for the backend's current default profile. */
export function historyLoader(api: PortfolioApi) {
	return async (kind: HistoryKind, _profileId: string, cursor?: HistoryCursor): Promise<{ items: HistoryTransaction[] }> => {
		const query = cursor && { before: cursor.before, before_id: cursor.beforeId };
		switch (kind) {
			case 'stock': return { items: (await api.stocks.transactions(query)).map(item => ({ ...item, kind })) };
			case 'bond': return { items: (await api.bonds.transactions(query)).map(item => ({ ...item, kind })) };
			case 'crypto': return { items: (await api.crypto.transactions(query)).map(item => ({ ...item, kind })) };
			case 'deposit': return { items: (await api.personal.depositTransactions(query)).map(item => ({ ...item, kind })) };
			case 'insurance': return { items: (await api.personal.insuranceTransactions(query)).map(item => ({ ...item, kind })) };
		}
	};
}
