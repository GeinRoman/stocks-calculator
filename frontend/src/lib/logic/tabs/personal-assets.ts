export type PersonalAssetType = 'deposit' | 'insurance';
export interface PersonalAsset {
	id: number;
	type: PersonalAssetType;
	name: string;
	provider: string;
	amountRub: number;
	annualInterest: number;
	startDate: string;
	endDate: string;
}
export interface PersonalAssetsSnapshot { deposits: PersonalAsset[]; insurances: PersonalAsset[] }
export type PersonalAssetInput = Omit<PersonalAsset, 'id' | 'type'>;
export type PersonalAssetUpdate = Omit<PersonalAssetInput, 'amountRub' | 'startDate'>;

export function emptyPersonalAssets(): PersonalAssetsSnapshot {
	return { deposits: [], insurances: [] };
}

export function clonePersonalAssets(snapshot: PersonalAssetsSnapshot): PersonalAssetsSnapshot {
	return { deposits: snapshot.deposits.map((asset) => ({ ...asset })), insurances: snapshot.insurances.map((asset) => ({ ...asset })) };
}

export function localCalendarDate(date = new Date()): string {
	return `${String(date.getFullYear()).padStart(4, '0')}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

// TEMP
function calendarDay(value: string): number | null {
	if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null;
	const [year, month, day] = value.split('-').map(Number);
	if (year < 1 || month < 1 || month > 12 || day < 1) return null;
	const date = new Date(0);
	date.setUTCHours(0, 0, 0, 0);
	date.setUTCFullYear(year, month - 1, day);
	return date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day
		? date.getTime() / 86400000 : null;
}

export function validatePersonalAsset(input: PersonalAssetInput): string | null {
	if (typeof input.name !== 'string' || !input.name.trim()) return 'Укажите название.';
	if (typeof input.provider !== 'string' || !input.provider.trim()) return 'Укажите банк или компанию.';
	if (!Number.isFinite(input.amountRub) || input.amountRub <= 0) return 'Сумма должна быть положительным числом.';
	if (!Number.isFinite(input.annualInterest) || input.annualInterest < 0) return 'Ставка должна быть неотрицательным числом.';
	const start = calendarDay(input.startDate), end = calendarDay(input.endDate);
	if (start === null || end === null) return 'Укажите корректные даты в формате ГГГГ-ММ-ДД.';
	if (end < start) return 'Дата окончания не может быть раньше даты начала.';
	return null;
}

export function getPersonalAssetStatus(endDate: string, today: string): 'expired' | 'ending' | 'normal' {
	const end = calendarDay(endDate), now = calendarDay(today);
	if (end === null || now === null) throw new Error('Укажите корректные даты.');
	return end < now ? 'expired' : end - now <= 14 ? 'ending' : 'normal';
}

// TEMP
export function estimatedMaturityAmount(asset: PersonalAsset): number | null {
	const start = calendarDay(asset.startDate), end = calendarDay(asset.endDate);
	if (start === null || end === null || end < start || !Number.isFinite(asset.amountRub) || asset.amountRub <= 0 || !Number.isFinite(asset.annualInterest) || asset.annualInterest < 0) return null;
	const result = asset.amountRub * (1 + asset.annualInterest / 100 * ((end - start) / 365));
	return Number.isFinite(result) ? result : null;
}

// TEMP
export function summarizePersonalAssets(snapshot: PersonalAssetsSnapshot): { deposits: number | null; insurances: number | null } {
	const sum = (assets: PersonalAsset[]): number | null => {
		let total = 0;
		for (const asset of assets) {
			if (!Number.isFinite(asset.amountRub)) return null;
			total += asset.amountRub;
			if (!Number.isFinite(total)) return null;
		}
		return total;
	};
	return { deposits: sum(snapshot.deposits), insurances: sum(snapshot.insurances) };
}
