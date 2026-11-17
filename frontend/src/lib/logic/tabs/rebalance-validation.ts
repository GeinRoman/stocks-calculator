import type { RebalanceProposal, RebalanceRequest, RebalanceSnapshots } from '#lib/logic/tabs/rebalance.js';
import { createDraftGroups, validateDraft, validateWeights } from '#lib/logic/tabs/stock-editor.js';
import { createBondDraftGroups, validateBondDraft } from '#lib/logic/tabs/bond-editor.js';
import { convertMoney, multiplyMoney, quoteBond } from '#lib/logic/tabs/currency-service.js';

export function validateRebalanceRequest(request: RebalanceRequest, snapshots: RebalanceSnapshots): string | null {
	if (typeof request.profileId !== 'string' || !request.profileId.trim()) return 'Выберите профиль.';
	if (typeof request.noSell !== 'boolean') return 'Некорректный запрет продажи.';
	if (request.stocks === null && request.bonds === null) return 'Выберите хотя бы один вид активов.';
	for (const asset of ['stocks', 'bonds'] as const) {
		const amount = request[asset];
		if (amount === null) continue;
		if (!Number.isFinite(amount) || Number(amount.toFixed(2)) !== amount) {
			return 'Введите сумму в рублях с точностью не более двух знаков после запятой.';
		}
		if (request.noSell && amount <= 0) return 'Без продажи необходимо указать положительное пополнение для каждого выбранного вида активов.';
		const groups = snapshots[asset].groups;
		const error = validateWeights(groups.map(({ id, targetPercent }) => ({ id, value: targetPercent ?? null })), true);
		if (error) return error;
		if (asset === 'stocks') {
			if (!snapshots.stocks.groups.some(({ stocks }) => stocks.length)) return 'В профиле нет акций для ребалансировки.';
			const error = validateDraft(createDraftGroups(snapshots.stocks));
			if (error) return error;
		} else {
			if (!snapshots.bonds.groups.some(({ bonds }) => bonds.length)) return 'В профиле нет облигаций для ребалансировки.';
			const error = validateBondDraft(createBondDraftGroups(snapshots.bonds));
			if (error) return error;
			// TEMP
			if (snapshots.bonds.groups.some(({ bonds }) => bonds.some((bond) =>
				!convertMoney(multiplyMoney(quoteBond(bond, true, false), bond.lotSize), 'RUB')))) {
				return 'Нет данных о цене или курсе валюты облигации для оценки в рублях.';
			}
		}
	}
	return null;
}

export function validateRebalanceProposal(proposal: RebalanceProposal, snapshots: RebalanceSnapshots): string | null {
	const error = validateRebalanceRequest(proposal.request, snapshots);
	if (error) return error;
	if (typeof proposal.id !== 'string' || !proposal.id.trim() || proposal.profileId !== proposal.request.profileId) {
		return 'Некорректное предложение или профиль.';
	}
	for (const asset of ['stocks', 'bonds'] as const) {
		const expected = proposal.request[asset] === null ? null : JSON.stringify(snapshots[asset]);
		if (proposal.baseline[asset] !== expected) return 'Портфель изменился. Запросите новое предложение.';
	}
	const ids = new Set<string>();
	const holdings = new Set<string>();
	for (const trade of proposal.trades) {
		if ((trade.asset !== 'stocks' && trade.asset !== 'bonds') || proposal.request[trade.asset] === null ||
			(trade.side !== 'buy' && trade.side !== 'sell')) return 'Некорректный вид актива или операции.';
		const expectedId = JSON.stringify([trade.asset, trade.groupId, trade.code, trade.side]);
		const holdingId = JSON.stringify([trade.asset, trade.code]);
		if (trade.id !== expectedId || ids.has(trade.id) || holdings.has(holdingId)) return 'Неизвестная или повторяющаяся операция.';
		ids.add(trade.id);
		holdings.add(holdingId);
		const stock = trade.asset === 'stocks'
			? snapshots.stocks.groups.find(({ id }) => id === trade.groupId)?.stocks.find(({ code }) => code === trade.code)
			: undefined;
		const bond = trade.asset === 'bonds'
			? snapshots.bonds.groups.find(({ id }) => id === trade.groupId)?.bonds.find(({ code }) => code === trade.code)
			: undefined;
		const current = stock ?? bond;
		const nativePrice = bond ? multiplyMoney(quoteBond(bond, true, false), bond.lotSize) : null;
		const price = stock ? stock.price * stock.lotSize : convertMoney(nativePrice, 'RUB')?.price;
		const available = current?.lotCount;
		if (!current || trade.name !== current.name || trade.unitPrice !== price || trade.availableQuantity !== available ||
			trade.lotSize !== current.lotSize || (bond && (trade.nativePrice?.currency !== nativePrice?.currency ||
			trade.nativePrice?.price !== nativePrice?.price)) || !Number.isFinite(trade.unitPrice)) return 'Данные операции больше не соответствуют портфелю.';
		const quantity = trade.quantity;
		if (quantity === undefined || !Number.isSafeInteger(quantity) || quantity < 0) return 'Количество должно быть целым неотрицательным числом.';
		if (trade.side === 'sell' && quantity > trade.availableQuantity) return 'Нельзя продать больше, чем есть в портфеле.';
		if (trade.side === 'buy' && !Number.isSafeInteger(trade.availableQuantity + quantity)) return 'Итоговое количество слишком велико.';
		if (bond && !Number.isSafeInteger((trade.availableQuantity + (trade.side === 'buy' ? quantity : -quantity)) * bond.lotSize)) return 'Итоговое количество облигаций слишком велико.';
		if (proposal.request.noSell && trade.side === 'sell' && quantity > 0) return 'Продажа запрещена.';
	}
	if (!getRebalanceEstimate(proposal)) return 'Сумма операций слишком велика.';
	return null;
}

export function getRebalanceEstimate(proposal: RebalanceProposal): { stocks: number | null; bonds: number | null; total: number } | null {
	const estimate = {
		stocks: proposal.request.stocks === null ? null : 0,
		bonds: proposal.request.bonds === null ? null : 0,
		total: 0
	};
	for (const trade of proposal.trades) {
		if (trade.quantity === undefined || !Number.isSafeInteger(trade.quantity) || trade.quantity < 0 ||
			!Number.isFinite(trade.unitPrice) || trade.unitPrice < 0 ||
			(trade.asset !== 'stocks' && trade.asset !== 'bonds') || estimate[trade.asset] === null ||
			(trade.side !== 'buy' && trade.side !== 'sell')) return null;
		const value = trade.unitPrice * trade.quantity * (trade.side === 'buy' ? 1 : -1);
		estimate[trade.asset]! += value;
		estimate.total += value;
		if (!Number.isFinite(estimate[trade.asset]) || !Number.isFinite(estimate.total)) return null;
	}
	return estimate;
}

export function getRebalanceErrorMessage(error: unknown, fallback: string): string {
	const message = error instanceof Error ? error.message : error;
	return typeof message === 'string' && message.trim() ? message : fallback;
}
