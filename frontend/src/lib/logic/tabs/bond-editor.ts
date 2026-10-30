import type { BondPortfolioSnapshot, BondSearchResult, DraftBondGroup } from '#lib/logic/tabs/bonds.js';
import { validateWeights } from '#lib/logic/tabs/stock-editor.js';

export function createBondDraftGroups(snapshot: BondPortfolioSnapshot): DraftBondGroup[] {
	return snapshot.groups.map((group) => ({
		id: group.id,
		name: group.name,
		targetPercent: group.targetPercent ?? null,
		bonds: group.bonds.map((bond) => ({
			code: bond.code, name: bond.name, price: bond.price, quantity: bond.quantity,
			redemptionPrice: bond.redemptionPrice ?? null,
			maturityDate: bond.maturityDate ?? null,
			couponAmount: bond.couponAmount ?? null,
			couponSchedule: bond.couponSchedule?.map((payment) => ({ ...payment })) ?? null
		}))
	}));
}

function validDate(value: string, dateOnly = false): boolean {
	const match = /^(\d{4})-(\d{2})-(\d{2})(.*)$/.exec(value);
	if (!match) return false;
	const [, year, month, day, time] = match;
	const calendar = new Date(`${year}-${month}-${day}T00:00:00Z`);
	if (!Number.isFinite(calendar.getTime()) || calendar.toISOString().slice(0, 10) !== `${year}-${month}-${day}`) return false;
	if (dateOnly && !time) return true;
	return /^T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/.test(time) && Number.isFinite(Date.parse(value));
}

export function validateBondDraft(groups: DraftBondGroup[]): string | null {
	const weightError = validateWeights(groups.map(({ id, targetPercent }) => ({ id, value: targetPercent })));
	if (weightError) return weightError;
	const names = new Set<string>();
	const codes = new Set<string>();
	let totalValue = 0;
	for (const group of groups) {
		const name = group.name.trim();
		if (!name || new TextEncoder().encode(name).length > 64) return 'Название группы должно содержать от 1 до 64 байт UTF-8.';
		if (names.has(name.toLowerCase())) return 'Названия групп облигаций должны быть уникальными.';
		names.add(name.toLowerCase());
		for (const bond of group.bonds) {
			const code = bond.code.trim().toUpperCase();
			if (!code || !bond.name.trim()) return 'Укажите код и название облигации.';
			if (codes.has(code)) return 'Облигация может находиться только в одной группе и не должна повторяться.';
			codes.add(code);
			if (bond.quantity === undefined || !Number.isSafeInteger(bond.quantity) || bond.quantity < 1) {
				return 'Количество облигаций должно быть положительным безопасным целым числом.';
			}
			if (!Number.isFinite(bond.price) || bond.price < 0) return 'Цена облигации должна быть неотрицательным конечным числом.';
			for (const amount of [bond.redemptionPrice, bond.couponAmount]) {
				if (amount !== null && (!Number.isFinite(amount) || amount < 0)) return 'Номинал и купон должны быть неотрицательными конечными числами или не заданы.';
			}
			if (bond.maturityDate !== null && !validDate(bond.maturityDate, true)) return 'Укажите корректную дату погашения или оставьте её неизвестной.';
			if (bond.couponSchedule !== null) {
				if (!Array.isArray(bond.couponSchedule) || !Number.isSafeInteger(bond.couponSchedule.length)) return 'Некорректное расписание купонов.';
				let couponValue = 0;
				for (const payment of bond.couponSchedule) {
					if (!payment || !validDate(payment.date)) return 'Дата купона должна быть ISO datetime с часовым поясом.';
					if (payment.amount !== null && (!Number.isFinite(payment.amount) || payment.amount < 0)) return 'Сумма купона должна быть неотрицательным конечным числом или не задана.';
					couponValue += payment.amount ?? 0;
					if (!Number.isFinite(couponValue) || !Number.isFinite(couponValue * bond.quantity)) return 'Общая сумма купонов слишком велика.';
				}
			}
			totalValue += bond.price * bond.quantity;
			if (!Number.isFinite(totalValue)) return 'Общая стоимость облигаций слишком велика.';
		}
	}
	return null;
}

export function addBondToGroup(
	groups: DraftBondGroup[], groupId: string, bond: BondSearchResult, quantity: number
): DraftBondGroup[] {
	if (!groups.some(({ id }) => id === groupId)) throw new Error('Группа не найдена.');
	const code = bond.code.trim().toUpperCase();
	if (groups.some((group) => group.id !== groupId && group.bonds.some((item) => item.code.trim().toUpperCase() === code))) {
		throw new Error('Облигация уже находится в другой группе.');
	}
	const quoteError = validateBondDraft([{ id: groupId, name: 'Проверка', targetPercent: null, bonds: [{ ...bond, quantity }] }]);
	if (quoteError) throw new Error(quoteError);
	const currentError = validateBondDraft(groups);
	if (currentError) throw new Error(currentError);
	const updated = groups.map((group) => {
		const bonds = group.bonds.map((item) => ({ ...item, couponSchedule: item.couponSchedule?.map((payment) => ({ ...payment })) ?? null }));
		if (group.id === groupId) {
			const existing = bonds.find((item) => item.code.trim().toUpperCase() === code);
			if (existing) existing.quantity = existing.quantity! + quantity;
			else bonds.push({ ...bond, code, name: bond.name.trim(), quantity, couponSchedule: bond.couponSchedule?.map((payment) => ({ ...payment })) ?? null });
		}
		return { ...group, bonds };
	});
	const error = validateBondDraft(updated);
	if (error) throw new Error(error);
	return updated;
}
