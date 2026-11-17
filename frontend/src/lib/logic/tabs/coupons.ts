import type { Money } from './currency.js';
export type Coupon = { paymentDate: string; amount: Money | null };
export type CouponSchedule = { code: string; coupons: Coupon[] };
