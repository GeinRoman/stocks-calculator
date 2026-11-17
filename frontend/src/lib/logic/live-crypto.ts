import type { portfolioApi } from '#lib/api/portfolio-api.js';
import type { CryptoCommand } from '#lib/logic/tabs/crypto.js';

type CryptoApi = ReturnType<typeof portfolioApi>['crypto'];

/** Each command is a separate transaction. Never replay a partially submitted draft. */
export async function saveLiveCrypto(api: CryptoApi, commands: CryptoCommand[]): Promise<void> {
	const ids = new Map<string, number>();
	const removed = new Set<string>();
	function id(key: string): number {
		return ids.get(key) ?? (Number.isSafeInteger(Number(key)) && Number(key) > 0 ? Number(key) : 0);
	}
	let completed = 0;
	for (const command of commands) {
		try {
			if (command.kind === 'add-group') await api.addGroups([{ name: command.name, weight: 0 }]);
			else if (command.kind === 'delete-group') await api.removeGroups([{ name: command.name, weight: 0 }]);
			else if (command.kind === 'rename-group') await api.renameGroup({ name_old: command.oldName, name_new: command.name });
			else if (command.kind === 'edit-asset') {
				const asset = command.asset;
				const assetId = id(asset.id);
				if (!assetId || removed.has(asset.id)) throw new Error('Актив больше не существует.');
				await api.updateAsset({ id: assetId, name: asset.name, platform: asset.platform, url: asset.url, annual_interest: asset.annualInterest });
			} else {
				const asset = command.asset;
				const assetId = removed.has(asset.id) ? 0 : id(asset.id);
				if (command.kind === 'sell' && !assetId) throw new Error('Актив больше не существует.');
				const operation = { id: assetId, name: asset.name, group_name: command.groupName, platform: asset.platform, url: asset.url, annual_interest: asset.annualInterest, amount: command.amount, price: { price: command.price, currency: command.currency } };
				if (command.kind === 'buy') {
					await api.buy([operation]);
					if (!assetId) {
						const assets = await api.assets();
						const created = assets.find(item => item.name === asset.name && item.group_name === command.groupName);
						if (!created) throw new Error('Не удалось получить созданный актив.');
						ids.set(asset.id, created.id);
						removed.delete(asset.id);
					}
				} else {
					await api.sell([operation]);
					if (command.amount === asset.amount) removed.add(asset.id);
				}
			}
			completed++;
		} catch {
			throw new Error(`Не удалось подтвердить все изменения (подтверждено операций: ${completed}). Результат последнего запроса может быть неизвестен. Отмените редактирование и обновите страницу перед повторной попыткой.`);
		}
	}
}
