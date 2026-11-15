<script lang="ts">
	import * as Drawer from '#lib/components/ui/drawer/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { decimal, priceInput, priceInputValue, type CryptoSnapshot } from '#lib/logic/tabs/crypto.js';
	import { updateCrypto } from '#lib/logic/tabs/crypto-service.js';
	import { onMount, untrack } from 'svelte';
	import { usePortfolioBack } from '#lib/logic/portfolio-back.js';
	let { snapshot, onclose, oncommit }: { snapshot: CryptoSnapshot; onclose: () => void; oncommit: (updated: CryptoSnapshot) => void } = $props();
	const initial = untrack(() => snapshot);
	let currency = $state<'USDT' | 'ETH' | 'BTC'>(initial.priceCurrency ?? 'USDT');
	let prices = $state<Record<string, string>>(Object.fromEntries(initial.groups.flatMap(g => g.assets.map(a => [a.id, a.price === null ? '' : priceInputValue(a.price)]))));
	let error = $state('');
	const back = usePortfolioBack();
	onMount(() => back?.enter('drawer', onclose));
	function save(event: SubmitEvent) {
		event.preventDefault(); error = '';
		try {
			const result = updateCrypto(snapshot, draft => {
				draft.priceCurrency = currency;
				for (const group of draft.groups) for (const asset of group.assets) {
					const text = (prices[asset.id] ?? '').trim();
					const value = text ? decimal(text) : null;
					if (text && (value === null || value < 0)) throw new Error(`Укажите неотрицательную цену для ${asset.name}.`);
					asset.price = value;
				}
			});
			oncommit(result); onclose();
		} catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось сохранить цены.'; }
	}
</script>

<Drawer.Root open={true} onOpenChange={(open) => { if (!open) onclose(); }}>
	<Drawer.Content class="mx-auto max-h-[90svh] max-w-3xl">
		<Drawer.Header><Drawer.Title>Указать цены</Drawer.Title><Drawer.Description>Укажите цены известных активов. Смена валюты меняет только обозначение, без пересчёта.</Drawer.Description></Drawer.Header>
		<form onsubmit={save} class="flex min-h-0 flex-1 flex-col">
			<div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-4">
				<div class="space-y-2"><Label for="crypto-price-currency">Валюта цен</Label><select id="crypto-price-currency" bind:value={currency} class="h-9 rounded-md border border-input bg-background px-3 text-sm"><option>USDT</option><option>ETH</option><option>BTC</option></select></div>
				<table class="w-full text-left text-sm"><thead><tr class="border-b"><th scope="col" class="p-2">Актив</th><th scope="col" class="p-2">Цена, {currency}</th></tr></thead><tbody>{#each snapshot.groups as group (group.id)}{#each group.assets as asset (asset.id)}<tr class="border-b"><th scope="row" class="break-words p-2 font-medium">{asset.name}</th><td class="p-2"><Input aria-label={`Цена ${asset.name}`} inputmode="decimal" value={prices[asset.id]} oninput={(event) => { const input = event.currentTarget; const cursor = input.selectionStart ?? input.value.length; const before = priceInput(input.value.slice(0, cursor)); const clean = priceInput(input.value); input.value = clean; prices[asset.id] = clean; input.setSelectionRange(before.length, before.length); }} placeholder="Нет данных" /></td></tr>{/each}{/each}</tbody></table>
				<Button type="button" variant="outline" onclick={() => { for (const id of Object.keys(prices)) prices[id] = ''; }}>Очистить все цены</Button>
				{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			</div>
			<Drawer.Footer class="shrink-0 border-t border-border bg-popover"><Button type="button" variant="outline" onclick={onclose}>Отмена</Button><Button type="submit">Применить</Button></Drawer.Footer>
		</form>
	</Drawer.Content>
</Drawer.Root>
