import type { BondPortfolioSnapshot, BondSearchResult, DraftBondGroup } from './bonds.js';
import type { Money } from './currency.js';
import { validateWeights } from './stock-editor.js';

function copyBond(bond: BondSearchResult): BondSearchResult {
	return { code: bond.code, name: bond.name, isin: bond.isin, cleanPrice: { ...bond.cleanPrice }, accruedInterest: { ...bond.accruedInterest }, maturityDate: bond.maturityDate, maturityPayment: bond.maturityPayment === null ? null : { ...bond.maturityPayment }, lotSize: bond.lotSize };
}
export function createBondDraftGroups(snapshot: BondPortfolioSnapshot): DraftBondGroup[] {
	return snapshot.groups.map((group) => ({ id: group.id, name: group.name, targetPercent: group.targetPercent, bonds: group.bonds.map((bond) => ({ ...copyBond(bond), lotCount: bond.lotCount })) }));
}
function validMoney(value: Money): boolean {
	return !!value && typeof value.currency === 'string' && /^[A-Z]{3}$/.test(value.currency) && Number.isFinite(value.price) && value.price >= 0;
}
function validDate(value: string): boolean {
	if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
	const date = new Date(`${value}T00:00:00Z`);
	return Number.isFinite(date.getTime()) && date.toISOString().slice(0, 10) === value;
}
export function validateBondDraft(groups: DraftBondGroup[]): string | null {
	const weightError = validateWeights(groups.map(({ id, targetPercent }) => ({ id, value: targetPercent })));
	if (weightError) return weightError;
	const names = new Set<string>(), codes = new Set<string>();
	for (const group of groups) {
		const name = group.name.trim();
		if (!name || new TextEncoder().encode(name).length > 64) return 'Название группы должно содержать от 1 до 64 байт UTF-8.';
		if (names.has(name.toLowerCase())) return 'Названия групп облигаций должны быть уникальными.';
		names.add(name.toLowerCase());
		for (const bond of group.bonds) {
			const code = bond.code.trim().toUpperCase();
			if (!code || !bond.name.trim() || typeof bond.isin !== 'string') return 'Укажите код и название облигации.';
			if (codes.has(code)) return 'Облигация может находиться только в одной группе и не должна повторяться.';
			codes.add(code);
			if (!Number.isSafeInteger(bond.lotCount) || bond.lotCount! < 1) return 'Количество облигаций в лотах должно быть положительным безопасным целым числом.';
			if (!Number.isSafeInteger(bond.lotSize) || bond.lotSize < 1) return 'Размер лота должен быть положительным безопасным целым числом.';
			if (!Number.isSafeInteger(bond.lotCount! * bond.lotSize)) return 'Количество облигаций слишком велико.';
			if (!validMoney(bond.cleanPrice) || !validMoney(bond.accruedInterest) || (bond.maturityPayment !== null && !validMoney(bond.maturityPayment))) return 'Некорректная денежная сумма облигации.';
			if (bond.maturityDate !== null && (typeof bond.maturityDate !== 'string' || !validDate(bond.maturityDate))) return 'Укажите корректную дату погашения или оставьте её неизвестной.';
		}
	}
	return null;
}
export function addBondToGroup(groups: DraftBondGroup[], groupId: string, bond: BondSearchResult, lotCount: number): DraftBondGroup[] {
	if (!groups.some(({ id }) => id === groupId)) throw new Error('Группа не найдена.');
	const error = validateBondDraft([{ id: groupId, name: 'Проверка', targetPercent: null, bonds: [{ ...bond, lotCount }] }]) ?? validateBondDraft(groups);
	if (error) throw new Error(error);
	const code = bond.code.trim().toUpperCase();
	if (groups.some((group) => group.id !== groupId && group.bonds.some((item) => item.code.trim().toUpperCase() === code))) throw new Error('Облигация уже находится в другой группе.');
	const updated = groups.map((group) => {
		const bonds = group.bonds.map((item) => ({ ...copyBond(item), lotCount: item.lotCount }));
		if (group.id === groupId) {
			const existing = bonds.find((item) => item.code.trim().toUpperCase() === code);
			if (existing) existing.lotCount = existing.lotCount! + lotCount;
			else bonds.push({ ...copyBond(bond), code, name: bond.name.trim(), lotCount });
		}
		return { ...group, bonds };
	});
	const updatedError = validateBondDraft(updated);
	if (updatedError) throw new Error(updatedError);
	return updated;
}
