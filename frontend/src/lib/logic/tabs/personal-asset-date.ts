const pad = (value: number, length = 2) => String(value).padStart(length, '0');

export function daysInMonth(year: number, month: number): number {
	if (month === 2) return year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0) ? 29 : 28;
	return [4, 6, 9, 11].includes(month) ? 30 : 31;
}

export function parseDateText(text: string): string | null {
	const match = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(text);
	if (!match) return null;
	const [, dd, mm, yyyy] = match;
	const day = Number(dd), month = Number(mm), year = Number(yyyy);
	if (year < 1 || month < 1 || month > 12 || day < 1 || day > daysInMonth(year, month)) return null;
	return `${yyyy}-${mm}-${dd}`;
}

export function formatDateISO(iso: string): string {
	const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso);
	if (!match) return '';
	const text = `${match[3]}/${match[2]}/${match[1]}`;
	return parseDateText(text) === iso ? text : '';
}

export function monthDays(year: number, month: number): (number | null)[] {
	const first = new Date(0);
	first.setUTCFullYear(year, month - 1, 1);
	const offset = (first.getUTCDay() + 6) % 7;
	return [...Array(offset).fill(null), ...Array.from({ length: daysInMonth(year, month) }, (_, i) => i + 1)];
}

export function dateISO(year: number, month: number, day: number): string {
	return `${pad(year, 4)}-${pad(month)}-${pad(day)}`;
}
