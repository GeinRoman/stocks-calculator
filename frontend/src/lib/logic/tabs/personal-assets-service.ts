import { clonePersonalAssets, emptyPersonalAssets, validatePersonalAsset } from './personal-assets.js';
import type { PersonalAsset, PersonalAssetInput, PersonalAssetType, PersonalAssetUpdate, PersonalAssetsSnapshot } from './personal-assets.js';
import { createDemoPersonalAssets } from './personal-assets-demo.js';

// TEMP
function checkedSnapshot(snapshot: PersonalAssetsSnapshot): PersonalAssetsSnapshot {
	for (const [type, assets] of [['deposit', snapshot.deposits], ['insurance', snapshot.insurances]] as const) {
		const ids = new Set<number>();
		for (const asset of assets) {
			if (asset.type !== type || !Number.isSafeInteger(asset.id) || asset.id <= 0 || ids.has(asset.id)) throw new Error('Некорректные или повторяющиеся идентификаторы активов.');
			ids.add(asset.id);
		}
	}
	return clonePersonalAssets(snapshot);
}

function assetsFor(snapshot: PersonalAssetsSnapshot, type: PersonalAssetType): PersonalAsset[] {
	if (type === 'deposit') return snapshot.deposits;
	if (type === 'insurance') return snapshot.insurances;
	throw new Error('Неизвестный тип актива.');
}

export async function getPersonalAssets(profileId: string): Promise<PersonalAssetsSnapshot> {
	return profileId === 'main' ? clonePersonalAssets(createDemoPersonalAssets().main) : emptyPersonalAssets();
}

export async function createPersonalAsset(snapshot: PersonalAssetsSnapshot, type: PersonalAssetType, input: PersonalAssetInput): Promise<PersonalAssetsSnapshot> {
	const result = checkedSnapshot(snapshot), assets = assetsFor(result, type);
	const error = validatePersonalAsset(input);
	if (error) throw new Error(error);
	const id = assets.reduce((max, asset) => Math.max(max, asset.id), 0) + 1;
	if (!Number.isSafeInteger(id)) throw new Error('Невозможно создать идентификатор актива.');
	assets.push({ id, type, name: input.name.trim(), provider: input.provider.trim(), amountRub: input.amountRub, annualInterest: input.annualInterest, startDate: input.startDate, endDate: input.endDate });
	return result;
}

export async function updatePersonalAsset(snapshot: PersonalAssetsSnapshot, type: PersonalAssetType, id: number, update: PersonalAssetUpdate): Promise<PersonalAssetsSnapshot> {
	const result = checkedSnapshot(snapshot), assets = assetsFor(result, type);
	const index = assets.findIndex((asset) => asset.id === id);
	if (index < 0) throw new Error('Актив не найден.');
	const asset = assets[index];
	// TEMP
	const input = { name: update.name, provider: update.provider, amountRub: asset.amountRub, annualInterest: update.annualInterest, startDate: asset.startDate, endDate: update.endDate };
	const error = validatePersonalAsset(input);
	if (error) throw new Error(error);
	assets[index] = { ...asset, ...input, name: input.name.trim(), provider: input.provider.trim() };
	return result;
}

export async function deletePersonalAsset(snapshot: PersonalAssetsSnapshot, type: PersonalAssetType, id: number): Promise<PersonalAssetsSnapshot> {
	const result = checkedSnapshot(snapshot), assets = assetsFor(result, type);
	const index = assets.findIndex((asset) => asset.id === id);
	if (index < 0) throw new Error('Актив не найден.');
	assets.splice(index, 1);
	return result;
}

// TEMP
export async function closePersonalAsset(snapshot: PersonalAssetsSnapshot, type: PersonalAssetType, id: number, closingAmountRub: number): Promise<PersonalAssetsSnapshot> {
	if (!Number.isFinite(closingAmountRub) || closingAmountRub <= 0) throw new Error('Сумма закрытия должна быть положительным числом.');
	return deletePersonalAsset(snapshot, type, id);
}
