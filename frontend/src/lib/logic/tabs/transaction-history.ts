export type HistoryKind = 'stock' | 'bond' | 'crypto' | 'deposit' | 'insurance';
type Base = { id: number; buying: boolean };
export type SecurityTransaction = Base & { instrument_code: string; lot_amount: number; lot_size: number; price: number; currency: string; time: string } & ({ kind: 'stock' } | { kind: 'bond' });
export type CryptoTransaction = Base & { kind: 'crypto'; asset_id: number; asset_name: string; platform: string; amount: number; price: number; currency: string; created_at: string };
export type DepositTransaction = Base & { kind: 'deposit'; name: string; bank_name: string; amount_rub: number; created_at: string };
export type InsuranceTransaction = Base & { kind: 'insurance'; name: string; company_name: string; amount_rub: number; created_at: string };
export type HistoryTransaction = SecurityTransaction | CryptoTransaction | DepositTransaction | InsuranceTransaction;
export type HistoryCursor = { before: string; beforeId: number };

// TEMP: Fixed illustrative records only; they are unrelated to edits or actual transactions.
const samples: Record<HistoryKind, HistoryTransaction[]> = {
	stock: [
		{ kind: 'stock', id: 8, instrument_code: 'SBER', lot_amount: 2, lot_size: 10, buying: true, price: 300, currency: 'RUB', time: '2026-10-22T12:00:00Z' },
		{ kind: 'stock', id: 7, instrument_code: 'GAZP', lot_amount: 1, lot_size: 10, buying: false, price: 140, currency: 'RUB', time: '2026-10-22T12:00:00Z' },
		{ kind: 'stock', id: 6, instrument_code: 'YDEX', lot_amount: 1, lot_size: 1, buying: true, price: 4000, currency: 'RUB', time: '2026-09-12T09:30:00Z' },
		{ kind: 'stock', id: 5, instrument_code: 'VTBR', lot_amount: 3, lot_size: 10, buying: true, price: 100, currency: 'RUB', time: '2026-08-10T08:00:00Z' }
	],
	bond: [
		{ kind: 'bond', id: 14, instrument_code: 'OFZ-DEMO-01', lot_amount: 10, lot_size: 1, buying: true, price: 760, currency: 'RUB', time: '2026-10-17T10:00:00Z' },
		{ kind: 'bond', id: 13, instrument_code: 'CORP-DEMO-01', lot_amount: 2, lot_size: 2, buying: false, price: 10.2, currency: 'USD', time: '2026-09-15T11:00:00Z' },
		{ kind: 'bond', id: 12, instrument_code: 'OFZ-DEMO-02', lot_amount: 5, lot_size: 1, buying: true, price: 1015, currency: 'RUB', time: '2026-08-21T11:00:00Z' },
		{ kind: 'bond', id: 11, instrument_code: 'CORP-DEMO-02', lot_amount: 3, lot_size: 1, buying: true, price: 8, currency: 'EUR', time: '2026-07-01T11:00:00Z' }
	],
	crypto: [
		{ kind: 'crypto', id: 24, asset_id: 1, asset_name: 'Bitcoin', platform: 'Биржа', amount: 0.025, buying: true, price: 62000, currency: 'USDT', created_at: '2026-10-19T13:00:00Z' },
		{ kind: 'crypto', id: 23, asset_id: 2, asset_name: 'Ethereum', platform: 'Кошелёк', amount: 0.5, buying: false, price: 3000, currency: 'USDT', created_at: '2026-09-10T08:00:00Z' },
		{ kind: 'crypto', id: 22, asset_id: 3, asset_name: 'Solana', platform: 'Биржа', amount: 1.75, buying: true, price: 120, currency: 'USDT', created_at: '2026-08-02T15:00:00Z' },
		{ kind: 'crypto', id: 21, asset_id: 1, asset_name: 'Bitcoin', platform: 'Биржа', amount: 0.1, buying: true, price: 60000, currency: 'USDT', created_at: '2026-07-01T10:00:00Z' }
	],
	deposit: [
		{ kind: 'deposit', id: 34, name: 'Срочный вклад', bank_name: 'Пример банка', amount_rub: 250000, buying: true, created_at: '2026-10-20T10:00:00Z' },
		{ kind: 'deposit', id: 33, name: 'Накопительный вклад', bank_name: 'Пример банка', amount_rub: 100000, buying: false, created_at: '2026-09-20T10:00:00Z' },
		{ kind: 'deposit', id: 32, name: 'Резервный вклад', bank_name: 'Пример банка', amount_rub: 50000, buying: true, created_at: '2026-08-20T10:00:00Z' },
		{ kind: 'deposit', id: 31, name: 'Накопительный вклад', bank_name: 'Пример банка', amount_rub: 150000, buying: true, created_at: '2026-07-20T10:00:00Z' }
	],
	insurance: [
		{ kind: 'insurance', id: 44, name: 'Страхование жизни', company_name: 'Пример компании', amount_rub: 300000, buying: true, created_at: '2026-10-15T10:00:00Z' },
		{ kind: 'insurance', id: 43, name: 'Накопительное страхование', company_name: 'Пример компании', amount_rub: 50000, buying: false, created_at: '2026-09-15T10:00:00Z' },
		{ kind: 'insurance', id: 42, name: 'Страхование здоровья', company_name: 'Пример компании', amount_rub: 30000, buying: true, created_at: '2026-08-15T10:00:00Z' },
		{ kind: 'insurance', id: 41, name: 'Накопительное страхование', company_name: 'Пример компании', amount_rub: 200000, buying: true, created_at: '2026-07-15T10:00:00Z' }
	]
};

export function historyTime(item: HistoryTransaction): string { return 'time' in item ? item.time : item.created_at; }

// TEMP: Small pages make pagination visible with the short static fixtures. The real endpoint returns up to 100.
export function demoHistoryPage(kind: HistoryKind, profileId: string, cursor?: HistoryCursor): { items: HistoryTransaction[] } {
	const sorted = (profileId === 'main' ? samples[kind] : []).filter(item => {
		const time = historyTime(item);
		return !cursor || time < cursor.before || (time === cursor.before && item.id < cursor.beforeId);
	}).sort((a, b) => historyTime(b).localeCompare(historyTime(a)) || b.id - a.id);
	return { items: sorted.slice(0, 3).map(item => ({ ...item })) };
}
