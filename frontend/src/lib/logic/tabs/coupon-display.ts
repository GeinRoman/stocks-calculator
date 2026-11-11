import type { Money } from '#lib/logic/tabs/currency.js';
import type { Coupon } from '#lib/logic/tabs/coupons.js';
import { convertMoney } from '#lib/logic/tabs/currency-service.js';

export function localCalendarDate(date: Date): string {
	return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

export function couponDate(coupon: Coupon): string {
	return coupon.paymentDate.slice(0, 10);
}

export function summarizeCoupons(coupons: Coupon[], today: string, quantity: number) {
	const sorted = [...coupons].sort((a, b) => couponDate(a).localeCompare(couponDate(b)));
	const upcoming = sorted.filter((coupon) => couponDate(coupon) >= today);
	let perBond: Money | null = null;
	const currencies = new Set(upcoming.flatMap(({ amount }) => amount ? [amount.currency] : []));
	const amounts = upcoming.map(({ amount }) => currencies.size > 1 ? convertMoney(amount, 'RUB') : amount);
	if (amounts.length && amounts.every((amount) => amount !== null && Number.isFinite(amount.price) && Math.abs(amount.price) <= Number.MAX_SAFE_INTEGER)) {
		const price = amounts.reduce((sum, amount) => sum + amount!.price, 0);
		if (Number.isFinite(price) && Math.abs(price) <= Number.MAX_SAFE_INTEGER) {
			perBond = { currency: amounts[0]!.currency, price };
		}
	}
	const holdingPrice = perBond && Number.isSafeInteger(quantity) && quantity >= 0 ? perBond.price * quantity : NaN;
	const holding: Money | null = perBond && Number.isFinite(holdingPrice) && Math.abs(holdingPrice) <= Number.MAX_SAFE_INTEGER
		? { currency: perBond.currency, price: holdingPrice } : null;
	return { sorted, remaining: upcoming.length, perBond, holding };
}
