import type { CouponSchedule } from './coupons.js';

// TEMP
const schedules: Record<string, CouponSchedule['coupons']> = {
	'OFZ-DEMO-01': [{ paymentDate: '2025-06-01', amount: { currency: 'RUB', price: 40 } }, { paymentDate: '2026-06-01', amount: { currency: 'RUB', price: 40 } }, { paymentDate: '2027-06-01', amount: { currency: 'RUB', price: 40 } }, { paymentDate: '2028-06-01', amount: { currency: 'RUB', price: 40 } }],
	'OFZ-DEMO-02': [{ paymentDate: '2025-09-01', amount: { currency: 'RUB', price: 35 } }, { paymentDate: '2026-09-01', amount: { currency: 'RUB', price: 45 } }, { paymentDate: '2027-09-01', amount: { currency: 'RUB', price: 50 } }, { paymentDate: '2028-09-01', amount: { currency: 'RUB', price: 55 } }],
	'CORP-DEMO-01': [{ paymentDate: '2025-03-01', amount: { currency: 'USD', price: 2 } }, { paymentDate: '2028-03-01', amount: null }],
	'CORP-DEMO-02': [{ paymentDate: '2029-03-01', amount: { currency: 'EUR', price: 3 } }],
	'ZERO-DEMO': [],
	'NEW-DEMO': [{ paymentDate: '2030-06-01', amount: { currency: 'CNY', price: 30 } }]
};
export async function getBondCoupons(code: string): Promise<CouponSchedule> {
	const normalized = code.trim().toUpperCase();
	const coupons = schedules[normalized];
	if (!coupons) throw new Error('Расписание купонов не найдено.');
	return { code: normalized, coupons: coupons.map(({ paymentDate, amount }) => ({ paymentDate, amount: amount && { ...amount } })) };
}
