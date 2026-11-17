import type { PersonalAsset, PersonalAssetsSnapshot } from './personal-assets.js';
import { localCalendarDate } from './personal-assets.js';

// TEMP
export function createDemoPersonalAssets(today = new Date()): Record<string, PersonalAssetsSnapshot> {
	const day = new Date(0);
	day.setUTCHours(0, 0, 0, 0);
	const [year, month, date] = localCalendarDate(today).split('-').map(Number);
	day.setUTCFullYear(year, month - 1, date);
	const offset = (days: number): string => {
		const date = new Date(day.getTime() + days * 86400000);
		return `${String(date.getUTCFullYear()).padStart(4, '0')}-${String(date.getUTCMonth() + 1).padStart(2, '0')}-${String(date.getUTCDate()).padStart(2, '0')}`;
	};
	const asset = (id: number, type: PersonalAsset['type'], name: string, provider: string, amountRub: number, annualInterest: number, end: number): PersonalAsset => ({
		id, type, name, provider, amountRub, annualInterest, startDate: offset(-365), endDate: offset(end)
	});
	return { main: {
		deposits: [asset(1, 'deposit', 'Срочный вклад', 'Пример банка', 250000, 5, -1), asset(2, 'deposit', 'Накопительный вклад', 'Пример банка', 150000, 8, 7)],
		insurances: [asset(1, 'insurance', 'Страхование жизни', 'Пример компании', 300000, 4, 14), asset(2, 'insurance', 'Накопительное страхование', 'Пример компании', 200000, 6, 15)]
	} };
}
