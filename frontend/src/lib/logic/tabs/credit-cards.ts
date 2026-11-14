import { formatDateISO } from './personal-asset-date.js';
export { getPersonalAssetStatus as getCreditCardStatus } from './personal-assets.js';

export interface CreditCard {
	id: number;
	name: string;
	bankName: string;
	paymentDate: string;
	paymentAmountRub: number;
	overallLimitRub: number;
}
export type CreditCardInput = Omit<CreditCard, 'id'>;
export type CreditCardsSnapshot = { cards: CreditCard[] };

export function emptyCreditCards(): CreditCardsSnapshot {
	return { cards: [] };
}

export function validateCreditCard(input: CreditCardInput): string | null {
	if (typeof input?.name !== 'string' || !input.name.trim()) return 'Укажите название карты.';
	if (typeof input.bankName !== 'string' || !input.bankName.trim()) return 'Укажите название банка.';
	if (typeof input.paymentDate !== 'string' || !formatDateISO(input.paymentDate)) return 'Укажите корректную дату платежа.';
	if (!Number.isFinite(input.paymentAmountRub) || input.paymentAmountRub < 0) return 'Сумма платежа должна быть неотрицательным числом.';
	if (!Number.isFinite(input.overallLimitRub) || input.overallLimitRub < 0) return 'Лимит должен быть неотрицательным числом.';
	return null;
}

export function sortCreditCards(cards: CreditCard[]): CreditCard[] {
	return [...cards].sort((a, b) => a.paymentDate < b.paymentDate ? -1 : a.paymentDate > b.paymentDate ? 1 : a.id - b.id);
}
