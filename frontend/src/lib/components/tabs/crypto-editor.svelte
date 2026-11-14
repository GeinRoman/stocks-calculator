<script lang="ts">
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { masonryGroups } from '#lib/logic/tabs/group-layout.js';
	import { decimal, type CryptoSnapshot, type CryptoAsset } from '#lib/logic/tabs/crypto.js';
	import { checkMetadata, groupName, trade, updateCrypto } from '#lib/logic/tabs/crypto-service.js';
	import { untrack } from 'svelte';

	let { snapshot, profileName, oncancel, onsave }: { snapshot: CryptoSnapshot; profileName: string; oncancel: () => void; onsave: (snapshot: CryptoSnapshot) => void } = $props();
	// TEMP: Drafts simulate composition changes and transactions only in memory.
	let draft = $state(untrack(() => updateCrypto(snapshot, () => {})));
	type Transaction = { id: string; assetId: string; kind: 'buy' | 'sell'; amount: number; price: number; currency: string };
	let transactions = $state<Transaction[]>([]);
	type Mode = 'delete-group' | 'rename-group' | 'add-asset' | 'edit-asset' | 'buy' | 'sell' | null;
	let mode = $state<Mode>(null), groupId = $state(''), assetId = $state('');
	let name = $state(''), platform = $state(''), url = $state(''), interest = $state('0'), amount = $state(''), price = $state(''), currency = $state('USDT'), newGroup = $state(''), error = $state(''), groupError = $state('');
	const selectedGroup = $derived(draft.groups.find(g => g.id === groupId));
	const selectedAsset = $derived(selectedGroup?.assets.find(a => a.id === assetId));
	function open(next: Mode, group: string, asset?: CryptoAsset) {
		mode = next; groupId = group; assetId = asset?.id ?? ''; name = next === 'rename-group' ? draft.groups.find(g => g.id === group)?.name ?? '' : asset?.name ?? ''; platform = asset?.platform ?? ''; url = asset?.url ?? ''; interest = String(asset?.annualInterest ?? 0); amount = ''; price = ''; currency = 'USDT'; error = '';
	}
	function addGroup(event: SubmitEvent) {
		event.preventDefault(); groupError = '';
		try { draft.groups.push({ id: crypto.randomUUID(), name: groupName(draft, newGroup), weight: 0, assets: [] }); newGroup = ''; }
		catch (cause) { groupError = cause instanceof Error ? cause.message : 'Не удалось добавить группу.'; }
	}
	function submit(event: SubmitEvent) {
		event.preventDefault(); error = '';
		try {
			const group = draft.groups.find(g => g.id === groupId);
			if (!group) throw new Error('Группа не найдена.');
			const asset = group.assets.find(a => a.id === assetId);
			if (mode === 'delete-group') {
				if (group.assets.length) throw new Error('Нельзя удалить группу с активами. Сначала продайте все активы и подтвердите изменения.');
				draft.groups = draft.groups.filter(g => g.id !== groupId);
			} else if (mode === 'rename-group') group.name = groupName(draft, name, group.id);
			else if (mode === 'edit-asset' || mode === 'add-asset') {
				const rate = decimal(interest);
				if (rate === null) throw new Error('Укажите корректную годовую ставку.');
				const metadata = { name: name.trim(), platform: platform.trim(), url: url.trim(), annualInterest: rate };
				checkMetadata(draft, metadata, asset?.id);
				if (mode === 'edit-asset' && asset) Object.assign(asset, metadata);
				else if (mode === 'add-asset') {
					const added: CryptoAsset = { id: crypto.randomUUID(), ...metadata, amount: 0, price: null };
					const quantity = decimal(amount) ?? NaN, unitPrice = decimal(price) ?? NaN;
					trade(added, quantity, unitPrice, currency, false);
					group.assets.push(added); transactions.push({ id: crypto.randomUUID(), assetId: added.id, kind: 'buy', amount: quantity, price: unitPrice, currency: currency.trim() });
				} else throw new Error('Актив не найден.');
			} else if ((mode === 'buy' || mode === 'sell') && asset) {
				const quantity = decimal(amount) ?? NaN, unitPrice = decimal(price) ?? NaN;
				trade(asset, quantity, unitPrice, currency, mode === 'sell');
				transactions.push({ id: crypto.randomUUID(), assetId: asset.id, kind: mode, amount: quantity, price: unitPrice, currency: currency.trim() });
			} else throw new Error('Актив не найден.');
			mode = null;
		} catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось изменить черновик.'; }
	}
	function save() {
		const result = updateCrypto(draft, copy => { for (const g of copy.groups) g.assets = g.assets.filter(a => a.amount > 0); });
		onsave(result);
	}
</script>

<section aria-labelledby="crypto-editor-title" class="mx-auto max-w-6xl">
	<!-- svelte-ignore a11y_no_noninteractive_tabindex: The bounded editor header must be keyboard-scrollable. -->
	<header tabindex="0" aria-label="Заголовок и действия редактирования состава" class="sticky top-0 z-20 mb-6 max-h-[50svh] overflow-y-auto border-b border-border bg-background py-3 focus-visible:outline-2 focus-visible:outline-ring">
		<h1 id="crypto-editor-title" class="text-xl font-semibold sm:text-2xl">Редактирование состава</h1><p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
		<p class="mt-3 text-xs text-muted-foreground">Изменения и операции применятся только после подтверждения. Данные демонстрационные, реальные операции не выполняются.</p>
		<div class="mt-4 flex flex-wrap gap-3"><Button onclick={save}>Подтвердить изменения</Button><Button variant="outline" onclick={oncancel}>Отмена</Button></div>
	</header>
	<div use:masonryGroups class="group-layout grid min-w-0 items-start gap-6 xl:grid-cols-2">
		{#each draft.groups as group (group.id)}<section aria-label={`Редактирование группы: ${group.name}`} class="min-w-0 overflow-hidden rounded-xl border border-border bg-card">
			<header class="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-primary/10 p-4"><h2 class="break-words font-semibold">{group.name}</h2><div class="flex flex-wrap gap-2"><Button size="sm" variant="outline" onclick={() => open('rename-group', group.id)}>Переименовать</Button><Button size="sm" variant="destructive" onclick={() => open('delete-group', group.id)}>Удалить группу</Button></div></header>
			<ul class="divide-y divide-border">{#each group.assets as asset (asset.id)}<li class="space-y-3 p-4"><div class="flex flex-wrap justify-between gap-2"><strong class="break-words">{asset.name}</strong><span class="text-sm">Количество: {asset.amount}</span></div><div class="flex flex-wrap gap-2"><Button size="sm" variant="outline" onclick={() => open('buy', group.id, asset)}>Покупка</Button><Button size="sm" variant="outline" disabled={asset.amount === 0} onclick={() => open('sell', group.id, asset)}>Продажа</Button><Button size="sm" variant="outline" onclick={() => open('edit-asset', group.id, asset)}>Изменить</Button></div>
				{#each transactions.filter(t => t.assetId === asset.id) as transaction (transaction.id)}<p class="rounded bg-muted px-3 py-2 text-xs">{transaction.kind === 'buy' ? 'Покупка' : 'Продажа'}: {transaction.amount} · цена {transaction.price} {transaction.currency} <span class="text-muted-foreground">(черновик)</span></p>{/each}
			</li>{:else}<li class="p-4 text-sm text-muted-foreground">В группе пока нет активов.</li>{/each}</ul>
			<div class="border-t border-border p-4"><Button size="sm" variant="outline" onclick={() => open('add-asset', group.id)}>Добавить актив</Button></div>
		</section>{:else}<p data-group-empty class="rounded-xl border border-dashed p-6 text-sm text-muted-foreground xl:col-span-2">Групп пока нет.</p>{/each}
	</div>
	<form onsubmit={addGroup} class="mt-6 space-y-3 rounded-xl border border-border bg-card p-4"><Label for="crypto-new-group">Новая группа</Label><div class="flex flex-wrap gap-3"><Input id="crypto-new-group" bind:value={newGroup} required class="min-w-0 flex-1 basis-48" /><Button type="submit">Добавить группу</Button></div>{#if groupError}<p role="alert" class="text-sm text-destructive">{groupError}</p>{/if}</form>
</section>

{#if mode}
	<Dialog.Root open={true} onOpenChange={(open) => { if (!open) mode = null; }}><Dialog.Content class="max-h-[90svh] overflow-y-auto"><Dialog.Header><Dialog.Title>{mode === 'delete-group' ? 'Удалить группу' : mode === 'rename-group' ? 'Переименовать группу' : mode === 'add-asset' ? 'Добавить криптоактив' : mode === 'edit-asset' ? 'Изменить криптоактив' : mode === 'buy' ? 'Записать покупку' : 'Записать продажу'}</Dialog.Title><Dialog.Description>{mode === 'delete-group' && selectedGroup?.assets.length ? 'Нельзя удалить группу с активами. Сначала продайте все активы и подтвердите изменения.' : mode === 'delete-group' ? `Удалить пустую группу «${selectedGroup?.name}» из черновика?` : 'Изменения сохраняются в черновике до подтверждения состава.'}</Dialog.Description></Dialog.Header>
		<form onsubmit={submit} class="space-y-4">
			{#if mode === 'rename-group'}<div class="space-y-2"><Label for="edit-crypto-group">Название группы</Label><Input id="edit-crypto-group" bind:value={name} required /></div>{/if}
			{#if mode === 'add-asset' || mode === 'edit-asset'}<div class="space-y-2"><Label for="edit-crypto-name">Название актива</Label><Input id="edit-crypto-name" bind:value={name} required /></div><div class="space-y-2"><Label for="edit-crypto-platform">Платформа</Label><Input id="edit-crypto-platform" bind:value={platform} required /></div><div class="space-y-2"><Label for="edit-crypto-url">Ссылка (необязательно)</Label><Input id="edit-crypto-url" bind:value={url} /></div><div class="space-y-2"><Label for="edit-crypto-interest">Годовая ставка, %</Label><Input id="edit-crypto-interest" inputmode="decimal" bind:value={interest} required /></div>{/if}
			{#if mode === 'add-asset' || mode === 'buy' || mode === 'sell'}{#if selectedAsset}<p class="text-sm">{selectedAsset.name} · Доступно: {selectedAsset.amount}</p>{/if}<div class="space-y-2"><Label for="edit-crypto-amount">Количество</Label><div class="flex items-center gap-2"><Input id="edit-crypto-amount" inputmode="decimal" bind:value={amount} required class="min-w-0 flex-1" />{#if mode === 'sell' && selectedAsset}<Button type="button" variant="outline" onclick={() => amount = String(selectedAsset!.amount)}>Всё</Button>{/if}</div></div><div class="space-y-2"><Label for="edit-crypto-price">Цена за единицу</Label><Input id="edit-crypto-price" inputmode="decimal" bind:value={price} required /></div><div class="space-y-2"><Label for="edit-crypto-currency">Валюта цены</Label><Input id="edit-crypto-currency" bind:value={currency} maxlength={16} required /></div>{/if}
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button type="button" variant="outline" onclick={() => mode = null}>{mode === 'delete-group' && selectedGroup?.assets.length ? 'Понятно' : 'Отмена'}</Button>{#if mode !== 'delete-group' || !selectedGroup?.assets.length}<Button type="submit">{mode === 'delete-group' ? 'Удалить' : 'Сохранить'}</Button>{/if}</Dialog.Footer>
		</form>
	</Dialog.Content></Dialog.Root>
{/if}
