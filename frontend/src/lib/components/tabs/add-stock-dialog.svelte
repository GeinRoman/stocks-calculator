<script lang="ts">
	import InstrumentSearchDialog from './instrument-search-dialog.svelte';
	import { searchStocks } from '#lib/logic/tabs/portfolio-service.js';
	import type { DraftGroup, StockSearchResult } from '#lib/logic/tabs/portfolio.js';
	let { groupId, groups, onadd, onclose }: {
		groupId: string; groups: DraftGroup[];
		onadd: (stock: StockSearchResult, lots: number) => void; onclose: () => void;
	} = $props();
	const group = $derived(groups.find(({ id }) => id === groupId));
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
</script>

<InstrumentSearchDialog groupName={group?.name ?? ''} instrument="акцию" queryLabel="Название или код акции" placeholder="Например, Россети" quantityLabel="Количество лотов" search={searchStocks} displayPrice={(stock) => currency.format(stock.price)} owner={(code) => groups.find((item) => item.id !== groupId && item.stocks.some((stock) => stock.code === code))?.name} alreadyHeld={(code) => !!group?.stocks.some((stock) => stock.code === code)} {onadd} {onclose}>
	{#snippet details(stock)}<p class="text-xs text-muted-foreground">Размер лота: {stock.lotSize} акций.</p>{/snippet}
</InstrumentSearchDialog>
