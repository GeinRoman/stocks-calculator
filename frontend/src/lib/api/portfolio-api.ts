import type { ApiClient } from './client.js';
import type * as DTO from './contracts.js';

// Endpoint wrappers perform no state changes; callers commit each successful operation themselves.
export function portfolioApi(client: ApiClient) {
	const get = <T>(path: string, signal?: AbortSignal) => client.request<T>(path, { signal });
	const write = (path: string, method: 'POST' | 'PATCH' | 'DELETE', body: unknown) => client.request<void>(path, { method, body });
	const cursorQuery = (cursor?: DTO.TransactionCursor) => cursor ? `?${new URLSearchParams({ before: cursor.before, ...(cursor.before_id === undefined ? {} : { before_id: String(cursor.before_id) }) })}` : '';
	const rebalanceQuery = (noSell: boolean, valueDiff: number) => `?${new URLSearchParams({ nosell: String(noSell), valdiff: String(valueDiff) })}`;
	return {
		getAll: (signal?: AbortSignal) => get<DTO.AllInfo>('/getall', signal),
		currencyRate: (from: string, to: string, signal?: AbortSignal) => get<DTO.CurrencyRate>(`/currencyrate?${new URLSearchParams({ from, to })}`, signal),
		profiles: {
			list: () => get<DTO.Profile[]>('/getprofiles'),
			create: (profile: DTO.Profile) => write('/createprofile', 'POST', profile),
			remove: (profile: DTO.Profile) => write('/removeprofile', 'DELETE', profile),
			setDefault: (profile: DTO.Profile) => write('/setdefaultprofile', 'PATCH', profile)
		},
		stocks: {
			groups: () => get<DTO.Group[]>('/getgroups'),
			addGroups: (groups: DTO.Group[]) => write('/addgroups', 'POST', groups),
			removeGroups: (groups: DTO.Group[]) => write('/removegroups', 'DELETE', groups),
			renameGroup: (group: DTO.RenameGroup) => write('/renamegroup', 'PATCH', group),
			updateWeights: (groups: DTO.Group[]) => write('/updateweights', 'PATCH', groups),
			add: (stocks: DTO.Stock[]) => write('/addstocks', 'POST', stocks),
			remove: (stocks: DTO.Stock[]) => write('/removestocks', 'DELETE', stocks),
			find: (names: string[]) => client.request<DTO.SearchResult<DTO.Stock>[]>('/findstocks', { method: 'POST', body: names }),
			rebalance: (noSell: boolean, valueDiff: number) => get<DTO.StockRebalance>(`/rebalance${rebalanceQuery(noSell, valueDiff)}`),
			transactions: (cursor?: DTO.TransactionCursor) => get<DTO.StockTransaction[]>(`/getstocktransactions${cursorQuery(cursor)}`)
		},
		bonds: {
			groups: () => get<DTO.Group[]>('/getbondgroups'),
			addGroups: (groups: DTO.Group[]) => write('/addbondgroups', 'POST', groups),
			removeGroups: (groups: DTO.Group[]) => write('/removebondgroups', 'DELETE', groups),
			renameGroup: (group: DTO.RenameGroup) => write('/renamebondgroup', 'PATCH', group),
			updateWeights: (groups: DTO.Group[]) => write('/updatebondweights', 'PATCH', groups),
			add: (bonds: DTO.BondOperation[]) => write('/addbonds', 'POST', bonds),
			remove: (bonds: DTO.BondOperation[]) => write('/removebonds', 'DELETE', bonds),
			find: (names: string[]) => client.request<DTO.SearchResult<DTO.Bond>[]>('/findbonds', { method: 'POST', body: names }),
			coupons: (code: string) => get<DTO.BondCoupons>(`/bondcoupons?${new URLSearchParams({ code })}`),
			rebalance: (noSell: boolean, valueDiff: number) => get<DTO.BondRebalance>(`/rebalancebonds${rebalanceQuery(noSell, valueDiff)}`),
			transactions: (cursor?: DTO.TransactionCursor) => get<DTO.StockTransaction[]>(`/getbondtransactions${cursorQuery(cursor)}`)
		},
		crypto: {
			groups: () => get<DTO.Group[]>('/getcryptogroups'),
			addGroups: (groups: DTO.Group[]) => write('/addcryptogroups', 'POST', groups),
			removeGroups: (groups: DTO.Group[]) => write('/removecryptogroups', 'DELETE', groups),
			renameGroup: (group: DTO.RenameGroup) => write('/renamecryptogroup', 'PATCH', group),
			updateWeights: (groups: DTO.Group[]) => write('/updatecryptoweights', 'PATCH', groups),
			updateAsset: (asset: DTO.CryptoAssetUpdate) => write('/updatecryptoasset', 'PATCH', asset),
			buy: (assets: DTO.CryptoOperation[]) => write('/buycrypto', 'POST', assets),
			sell: (assets: DTO.CryptoOperation[]) => write('/sellcrypto', 'DELETE', assets),
			transactions: (cursor?: DTO.TransactionCursor) => get<DTO.CryptoTransaction[]>(`/getcryptotransactions${cursorQuery(cursor)}`)
		},
		personal: {
			addDeposit: (value: DTO.Deposit) => write('/adddeposit', 'POST', value),
			updateDeposit: (value: DTO.DepositUpdate) => write('/updatedeposit', 'PATCH', value),
			removeDeposit: (id: number) => write('/removedeposit', 'DELETE', { id }),
			closeDeposit: (id: number, closing_amount_rub: number) => write('/closedeposit', 'POST', { id, closing_amount_rub }),
			addInsurance: (value: DTO.Insurance) => write('/addinsurance', 'POST', value),
			updateInsurance: (value: DTO.InsuranceUpdate) => write('/updateinsurance', 'PATCH', value),
			removeInsurance: (id: number) => write('/removeinsurance', 'DELETE', { id }),
			closeInsurance: (id: number, closing_amount_rub: number) => write('/closeinsurance', 'POST', { id, closing_amount_rub }),
			addCreditCard: (value: DTO.CreditCard) => write('/addcreditcard', 'POST', value),
			updateCreditCard: (value: DTO.CreditCardUpdate) => write('/updatecreditcard', 'PATCH', value),
			removeCreditCard: (id: number) => write('/removecreditcard', 'DELETE', { id }),
			depositTransactions: (cursor?: DTO.TransactionCursor) => get<DTO.DepositTransaction[]>(`/getdeposittransactions${cursorQuery(cursor)}`),
			insuranceTransactions: (cursor?: DTO.TransactionCursor) => get<DTO.InsuranceTransaction[]>(`/getinsurancetransactions${cursorQuery(cursor)}`)
		}
	};
}
