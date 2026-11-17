import type { Money } from './currency.js';

// TEMP
export const sampleRates: Readonly<Record<string, number>> = { RUB: 1, USD: 90, EUR: 100, CNY: 12 };
const valid = (money: Money | null): money is Money => money !== null && typeof money.currency === 'string' && /^[A-Z]{3}$/.test(money.currency) && Number.isFinite(money.price) && money.price >= 0;

export function formatMoney(money: Money | null): string {
	return valid(money) ? `${new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 2 }).format(money.price)} ${money.currency}` : 'Нет данных';
}
export function convertMoney(money: Money | null, to: string, rates: Readonly<Record<string, number>> = sampleRates): Money | null {
	if (!valid(money) || typeof to !== 'string' || !/^[A-Z]{3}$/.test(to)) return null;
	if (money.currency === to) return { ...money };
	const fromRate = money.currency === 'RUB' ? 1 : rates[money.currency], toRate = to === 'RUB' ? 1 : rates[to];
	if (!Number.isFinite(fromRate) || !Number.isFinite(toRate) || fromRate <= 0 || toRate <= 0) return null;
	const price = money.price * fromRate / toRate;
	return Number.isFinite(price) ? { currency: to, price } : null;
}
export function addDirtyPrice(clean: Money | null, accrued: Money | null, rates: Readonly<Record<string, number>> = sampleRates): Money | null {
	if (!valid(clean) || !valid(accrued)) return null;
	const a = clean.currency === accrued.currency ? clean : convertMoney(clean, 'RUB', rates);
	const b = clean.currency === accrued.currency ? accrued : convertMoney(accrued, 'RUB', rates);
	if (!a || !b || !Number.isFinite(a.price + b.price)) return null;
	return { currency: a.currency, price: a.price + b.price };
}
export function quoteBond(bond: { cleanPrice: Money; accruedInterest: Money }, includeAccrued: boolean, inRub: boolean, rates: Readonly<Record<string, number>> = sampleRates): Money | null {
	const quote = includeAccrued ? addDirtyPrice(bond.cleanPrice, bond.accruedInterest, rates) : valid(bond.cleanPrice) ? { ...bond.cleanPrice } : null;
	return inRub ? convertMoney(quote, 'RUB', rates) : quote;
}
export function multiplyMoney(money: Money | null, quantity: number): Money | null {
	if (!valid(money) || !Number.isFinite(quantity) || quantity < 0 || !Number.isFinite(money.price * quantity)) return null;
	return { currency: money.currency, price: money.price * quantity };
}
