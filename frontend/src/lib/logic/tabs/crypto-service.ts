import { transactAmount, validateMetadata, type CryptoAsset, type CryptoMetadata, type CryptoSnapshot } from './crypto.js';
import type { GroupWeight } from './portfolio.js';

// TEMP: These operations only update a copied, in-memory snapshot; no orders are placed.
export function updateCrypto(snapshot: CryptoSnapshot, action: (draft: CryptoSnapshot) => void): CryptoSnapshot {
	const draft = { ...snapshot, groups: snapshot.groups.map((group) => ({ ...group, assets: group.assets.map((asset) => ({ ...asset })) })) };
	action(draft); return draft;
}
export function groupName(snapshot: CryptoSnapshot, name: string, except?: string): string {
	const value = name.trim();
	if (!value || new TextEncoder().encode(value).length > 64) throw new Error('Название группы должно содержать от 1 до 64 байт UTF-8.');
	if (snapshot.groups.some((group) => group.name === value && group.id !== except)) throw new Error('Группа с таким названием уже существует.');
	return value;
}
export function checkMetadata(snapshot: CryptoSnapshot, input: CryptoMetadata, except?: string): void {
	const error = validateMetadata(input);
	if (error) throw new Error(error);
	if (snapshot.groups.some((group) => group.assets.some((asset) => asset.id !== except && asset.name === input.name))) throw new Error('Актив с таким названием уже существует в профиле.');
}
export function trade(asset: CryptoAsset, amount: number, price: number, currency: string, sell: boolean): void {
	if (!Number.isFinite(price) || price <= 0 || !currency.trim() || currency.trim().length > 16) throw new Error('Укажите положительную цену и валюту (до 16 символов).');
	asset.amount = transactAmount(asset.amount, amount, sell);
}

export function saveCryptoWeights(snapshot: CryptoSnapshot, weights: GroupWeight[]): CryptoSnapshot {
	return updateCrypto(snapshot, (draft) => {
		if (weights.length !== draft.groups.length || new Set(weights.map(({ id }) => id)).size !== weights.length) throw new Error('Укажите доли всех групп.');
		let sum = 0;
		for (const group of draft.groups) {
			const value = weights.find((weight) => weight.id === group.id)?.value;
			if (value == null || !Number.isInteger(value) || value < 0 || value > 100) throw new Error('Укажите целые доли от 0 до 100.');
			group.weight = value; sum += value;
		}
		if (sum !== 100) throw new Error('Задайте доли всех групп с общей суммой 100%.');
	});
}
