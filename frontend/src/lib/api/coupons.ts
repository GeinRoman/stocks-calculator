import type { BondCoupons } from './contracts.js';
import type { CouponSchedule } from '../logic/tabs/coupons.js';

export function couponSchedule(code: string, result: BondCoupons): CouponSchedule {
	if (!result || result.code !== code || !Array.isArray(result.coupons)) throw new Error('Некорректный график купонов.');
	return { code: result.code, coupons: result.coupons.map(coupon => ({ paymentDate: coupon.payment_date, amount: coupon.amount })) };
}
