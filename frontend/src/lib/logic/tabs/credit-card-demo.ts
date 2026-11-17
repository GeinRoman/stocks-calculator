import { localCalendarDate } from './personal-assets.js';
import type { CreditCard, CreditCardsSnapshot } from './credit-cards.js';

// TEMP
export function createDemoCreditCards(today = new Date()): Record<string, CreditCardsSnapshot> {
	const [year, month, date] = localCalendarDate(today).split('-').map(Number);
	const day = new Date(0);
	day.setUTCHours(0, 0, 0, 0);
	day.setUTCFullYear(year, month - 1, date);
	const offset = (days: number): string => {
		const result = new Date(day.getTime() + days * 86400000);
		const year = result.getUTCFullYear();
		if (!Number.isFinite(year) || year < 1 || year > 9999) throw new Error('Дата примера вне допустимого диапазона.');
		return `${String(year).padStart(4, '0')}-${String(result.getUTCMonth() + 1).padStart(2, '0')}-${String(result.getUTCDate()).padStart(2, '0')}`;
	};
	const card = (id: number, days: number, paymentAmountRub: number, overallLimitRub: number): CreditCard => ({
		id, name: `Пример карты ${id}`, bankName: 'Пример банка', paymentDate: offset(days), paymentAmountRub, overallLimitRub
	});
	return { main: { cards: [card(1, 14, 5000, 100000), card(2, -3, 0, 50000), card(3, 15, 2000, 75000), card(4, 0, 1500, 30000), card(5, 7, 0, 0)] }, reserve: { cards: [] } };
}
