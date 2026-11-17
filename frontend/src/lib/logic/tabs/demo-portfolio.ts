import type { PortfolioSnapshot, Profile } from '#lib/logic/tabs/portfolio.js';

// TEMP
export const demoProfiles: Profile[] = [
	{ id: 'main', name: 'Профиль номер 1' },
	{ id: 'reserve', name: 'Профиль номер 2' }
];

export const emptyPortfolio: PortfolioSnapshot = {
	totalStocksValue: 0,
	groups: [],
	assetValues: { stocks: 0, bonds: 0, deposits: 0 }
};

// TEMP
// Static demo valuations, not browser-side financial calculations or live quotes.
export const demoPortfolios: Record<string, PortfolioSnapshot> = {
	main: {
		totalStocksValue: 100000,
		assetValues: { stocks: 100000, bonds: 250000, deposits: 150000 },
		groups: [
			{
				id: 'banks', name: 'Банки', totalValue: 40000, targetPercent: 30, actualPercent: 40,
				stocks: [
					{ code: 'SBER', name: 'Сбербанк', price: 300, lotCount: 10, lotSize: 10, totalValue: 30000, targetPercent: 15, actualPercent: 30 },
					{ code: 'VTBR', name: 'ВТБ', price: 100, lotCount: 10, lotSize: 10, totalValue: 10000, targetPercent: 15, actualPercent: 10 }
				]
			},
			{
				id: 'energy', name: 'Энергетика', totalValue: 35000, targetPercent: 40, actualPercent: 35,
				stocks: [
					{ code: 'LKOH', name: 'Лукойл', price: 7000, lotCount: 4, lotSize: 1, totalValue: 28000, targetPercent: 20, actualPercent: 28 },
					{ code: 'GAZP', name: 'Газпром', price: 140, lotCount: 5, lotSize: 10, totalValue: 7000, targetPercent: 20, actualPercent: 7 }
				]
			},
			{
				id: 'technology', name: 'Технологии', totalValue: 25000, targetPercent: 30, actualPercent: 25,
				stocks: [
					{ code: 'YDEX', name: 'Яндекс', price: 4000, lotCount: 5, lotSize: 1, totalValue: 20000, targetPercent: 15, actualPercent: 20 },
					{ code: 'POSI', name: 'Группа Позитив', price: 1000, lotCount: 5, lotSize: 1, totalValue: 5000, targetPercent: 15, actualPercent: 5 }
				]
			}
		]
	},
	reserve: emptyPortfolio
};
