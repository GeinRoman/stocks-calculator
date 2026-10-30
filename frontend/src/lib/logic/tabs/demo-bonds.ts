import type { BondPortfolioSnapshot, BondSearchResult } from '#lib/logic/tabs/bonds.js';
import { computeBondSnapshot } from '#lib/logic/tabs/bond-snapshot.js';

// TEMP
// Invented demo instruments and prices, NOT live MOEX quotes. Schedules cover each
// illustrative instrument's entire lifetime, rather than only its future payments.
export const sampleBonds: BondSearchResult[] = [
	{
		code: 'OFZ-DEMO-01', name: 'ОФЗ демонстрационная 01', price: 750, redemptionPrice: 1000,
		maturityDate: '2028-06-01', couponAmount: 40,
		couponSchedule: [
			{ date: '2025-06-01T12:00:00+03:00', amount: 40 },
			{ date: '2026-06-01T12:00:00+03:00', amount: 40 },
			{ date: '2027-06-01T12:00:00+03:00', amount: 40 },
			{ date: '2028-06-01T12:00:00+03:00', amount: 40 }
		]
	},
	{
		code: 'OFZ-DEMO-02', name: 'ОФЗ переменный купон', price: 1000, redemptionPrice: 1000,
		maturityDate: '2028-09-01', couponAmount: 45,
		couponSchedule: [
			{ date: '2025-09-01T12:00:00+03:00', amount: 35 },
			{ date: '2026-09-01T12:00:00+03:00', amount: 45 },
			{ date: '2027-09-01T12:00:00+03:00', amount: 50 },
			{ date: '2028-09-01T12:00:00+03:00', amount: 55 }
		]
	},
	{
		code: 'CORP-DEMO-01', name: 'Корпоративная демонстрационная', price: 1000, redemptionPrice: 1000,
		maturityDate: '2028-03-01', couponAmount: null,
		couponSchedule: [
			{ date: '2025-03-01T12:00:00+03:00', amount: 60 },
			{ date: '2026-03-01T12:00:00+03:00', amount: 65 },
			{ date: '2027-03-01T12:00:00+03:00', amount: 70 },
			{ date: '2028-03-01T12:00:00+03:00', amount: null }
		]
	},
	{
		code: 'CORP-DEMO-02', name: 'Корпоративная неизвестное расписание', price: 800,
		redemptionPrice: null, maturityDate: null, couponAmount: null, couponSchedule: null
	},
	{
		code: 'ZERO-DEMO', name: 'Погашенная бескупонная демонстрационная', price: 0,
		redemptionPrice: 1000, maturityDate: '2025-01-01', couponAmount: 0, couponSchedule: []
	},
	{
		code: 'NEW-DEMO', name: 'Новая корпоративная демонстрационная', price: 900,
		redemptionPrice: 1000, maturityDate: '2030-06-01', couponAmount: 30,
		couponSchedule: [
			{ date: '2025-06-01T12:00:00+03:00', amount: 30 },
			{ date: '2026-06-01T12:00:00+03:00', amount: 30 },
			{ date: '2027-06-01T12:00:00+03:00', amount: 30 },
			{ date: '2028-06-01T12:00:00+03:00', amount: 30 },
			{ date: '2029-06-01T12:00:00+03:00', amount: 30 },
			{ date: '2030-06-01T12:00:00+03:00', amount: 30 }
		]
	}
];

// TEMP
export function createEmptyBondPortfolio(asOf = new Date()): BondPortfolioSnapshot {
	return computeBondSnapshot([], asOf);
}

// TEMP
export function createDemoBondPortfolios(asOf = new Date()): Record<string, BondPortfolioSnapshot> {
	return {
		main: computeBondSnapshot([
			{ id: 'government', name: 'Государственные', targetPercent: 60, bonds: [
				{ ...sampleBonds[0], quantity: 100 }, { ...sampleBonds[1], quantity: 75 }
			] },
			{ id: 'corporate', name: 'Корпоративные', targetPercent: 40, bonds: [
				{ ...sampleBonds[2], quantity: 60 }, { ...sampleBonds[3], quantity: 50 },
				{ ...sampleBonds[4], quantity: 10 }
			] }
		], asOf),
		reserve: createEmptyBondPortfolio(asOf)
	};
}
