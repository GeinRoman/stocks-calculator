import type { Bond, CurrencyRate } from './contracts.js';

/** Load each currency needed for bond valuation once; failed lookups remain unavailable. */
export async function loadBondRates(
	bonds: readonly Bond[],
	getRate: (from: string, to: string) => Promise<CurrencyRate>
): Promise<{ rates: Record<string, number>; unavailable: string[] }> {
	const currencies = [...new Set(bonds.flatMap(bond => [bond.clean_price.currency, bond.accrued_interest.currency]).filter(currency => currency !== 'RUB'))];
	const results = await Promise.allSettled(currencies.map(currency => getRate(currency, 'RUB')));
	const rates: Record<string, number> = Object.create(null);
	const unavailable: string[] = [];
	for (const [index, result] of results.entries()) {
		const currency = currencies[index];
		if (result.status === 'fulfilled' && result.value.from === currency && result.value.to === 'RUB' && Number.isFinite(result.value.rate) && result.value.rate > 0) rates[currency] = result.value.rate;
		else unavailable.push(currency);
	}
	return { rates, unavailable };
}
