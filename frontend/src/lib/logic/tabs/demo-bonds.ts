import type { BondPortfolioSnapshot, BondSearchResult } from './bonds.js';
import { computeBondSnapshot } from './bond-snapshot.js';

// TEMP
export const sampleBonds: BondSearchResult[] = [
	{ code: 'OFZ-DEMO-01', name: 'ОФЗ демонстрационная 01', isin: 'RUDEMO000001', cleanPrice: { currency: 'RUB', price: 750 }, accruedInterest: { currency: 'RUB', price: 10 }, maturityDate: '2028-06-01', maturityPayment: { currency: 'RUB', price: 1000 }, lotSize: 1 },
	{ code: 'OFZ-DEMO-02', name: 'ОФЗ переменный купон', isin: 'RUDEMO000002', cleanPrice: { currency: 'RUB', price: 1000 }, accruedInterest: { currency: 'RUB', price: 15 }, maturityDate: '2028-09-01', maturityPayment: { currency: 'RUB', price: 1000 }, lotSize: 1 },
	{ code: 'CORP-DEMO-01', name: 'Корпоративная демонстрационная', isin: 'RUDEMO000003', cleanPrice: { currency: 'USD', price: 10 }, accruedInterest: { currency: 'USD', price: 0.2 }, maturityDate: '2028-03-01', maturityPayment: { currency: 'USD', price: 10 }, lotSize: 2 },
	{ code: 'CORP-DEMO-02', name: 'Корпоративная неизвестное расписание', isin: 'RUDEMO000004', cleanPrice: { currency: 'EUR', price: 8 }, accruedInterest: { currency: 'EUR', price: 0.1 }, maturityDate: null, maturityPayment: null, lotSize: 1 },
	{ code: 'ZERO-DEMO', name: 'Погашенная бескупонная демонстрационная', isin: 'RUDEMO000005', cleanPrice: { currency: 'RUB', price: 0 }, accruedInterest: { currency: 'RUB', price: 0 }, maturityDate: '2025-01-01', maturityPayment: { currency: 'RUB', price: 1000 }, lotSize: 1 },
	{ code: 'NEW-DEMO', name: 'Новая корпоративная демонстрационная', isin: 'RUDEMO000006', cleanPrice: { currency: 'CNY', price: 75 }, accruedInterest: { currency: 'CNY', price: 1 }, maturityDate: '2030-06-01', maturityPayment: { currency: 'CNY', price: 80 }, lotSize: 1 }
];
// TEMP
export function createEmptyBondPortfolio(asOf = new Date()): BondPortfolioSnapshot { return computeBondSnapshot([], asOf); }
export function createDemoBondPortfolios(asOf = new Date()): Record<string, BondPortfolioSnapshot> {
	return { main: computeBondSnapshot([
		{ id: 'government', name: 'Государственные', targetPercent: 60, bonds: [{ ...sampleBonds[0], lotCount: 100 }, { ...sampleBonds[1], lotCount: 75 }] },
		{ id: 'corporate', name: 'Корпоративные', targetPercent: 40, bonds: [{ ...sampleBonds[2], lotCount: 60 }, { ...sampleBonds[3], lotCount: 50 }, { ...sampleBonds[4], lotCount: 10 }] }
	], asOf), reserve: createEmptyBondPortfolio(asOf) };
}
