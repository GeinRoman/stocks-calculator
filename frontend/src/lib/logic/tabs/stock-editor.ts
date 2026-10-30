import type {
	DraftGroup, GroupWeight, PortfolioSnapshot, StockSearchResult
} from '#lib/logic/tabs/portfolio.js';

export function createDraftGroups(snapshot: PortfolioSnapshot): DraftGroup[] {
	return snapshot.groups.map((group) => ({
		id: group.id,
		name: group.name,
		targetPercent: group.targetPercent ?? null,
		stocks: group.stocks.map(({ code, name, price, lotSize, lotCount }) => ({
			code, name, price, lotSize, lotCount
		}))
	}));
}

export function validateWeights(weights: GroupWeight[]): string | null {
	const ids = new Set<string>();
	let total = 0;
	for (const { id, value } of weights) {
		if (!id.trim() || ids.has(id)) return 'Идентификаторы групп должны быть непустыми и уникальными.';
		ids.add(id);
		if (value === null) continue;
		if (!Number.isInteger(value) || value < 0 || value > 100) {
			return 'Доля группы должна быть целым числом от 0 до 100 или не задана.';
		}
		total += value;
	}
	return total > 100 ? 'Сумма долей групп не должна превышать 100%.' : null;
}

export function validateDraft(groups: DraftGroup[]): string | null {
	const weightError = validateWeights(groups.map(({ id, targetPercent }) => ({ id, value: targetPercent })));
	if (weightError) return weightError;
	const names = new Set<string>();
	const codes = new Set<string>();
	for (const group of groups) {
		const name = group.name.trim();
		if (!name || new TextEncoder().encode(name).length > 64) {
			return 'Название группы должно содержать от 1 до 64 байт UTF-8.';
		}
		const normalizedName = name.toLowerCase();
		if (names.has(normalizedName)) return 'Названия групп должны быть уникальными.';
		names.add(normalizedName);
		for (const stock of group.stocks) {
			const code = stock.code.trim().toUpperCase();
			if (!code || !stock.name.trim()) return 'Укажите код и название акции.';
			if (codes.has(code)) return 'Акция может находиться только в одной группе и не должна повторяться.';
			codes.add(code);
			if (!Number.isSafeInteger(stock.lotCount) || stock.lotCount === undefined || stock.lotCount <= 0) {
				return 'Количество лотов должно быть положительным безопасным целым числом.';
			}
			if (!Number.isSafeInteger(stock.lotSize) || stock.lotSize <= 0) {
				return 'Размер лота должен быть положительным безопасным целым числом.';
			}
			if (!Number.isFinite(stock.price) || stock.price <= 0 ||
				!Number.isFinite(stock.price * stock.lotSize * stock.lotCount)) {
				return 'Цена и стоимость акции должны быть положительными конечными числами.';
			}
		}
	}
	return null;
}

export function addStockToGroup(
	groups: DraftGroup[], groupId: string, stock: StockSearchResult, lotCount: number
): DraftGroup[] {
	if (!groups.some(({ id }) => id === groupId)) throw new Error('Группа не найдена.');
	if (!Number.isSafeInteger(lotCount) || lotCount <= 0) {
		throw new Error('Количество лотов должно быть положительным безопасным целым числом.');
	}
	const code = stock.code.trim().toUpperCase();
	if (groups.some((group) => group.id !== groupId &&
		group.stocks.some((item) => item.code.trim().toUpperCase() === code))) {
		throw new Error('Акция уже находится в другой группе.');
	}
	// Validate incoming quote metadata even when merging with an existing holding.
	const quoteError = validateDraft([{ id: groupId, name: 'Проверка', targetPercent: null, stocks: [{ ...stock, lotCount }] }]);
	if (quoteError) throw new Error(quoteError);
	const updated = groups.map((group) => {
		const stocks = group.stocks.map((item) => ({ ...item }));
		if (group.id === groupId) {
			const existing = stocks.find((item) => item.code.trim().toUpperCase() === code);
			if (existing) {
				if (!Number.isSafeInteger(existing.lotCount) || existing.lotCount === undefined || existing.lotCount <= 0) {
					throw new Error('Количество лотов должно быть положительным безопасным целым числом.');
				}
				existing.lotCount += lotCount;
			} else stocks.push({ ...stock, code, name: stock.name.trim(), lotCount });
		}
		return { ...group, stocks };
	});
	const error = validateDraft(updated);
	if (error) throw new Error(error);
	return updated;
}
