<script lang="ts">
	import { onDestroy } from 'svelte';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { searchStocks } from '#lib/logic/tabs/portfolio-service.js';
	import type { DraftGroup, StockSearchResult } from '#lib/logic/tabs/portfolio.js';

	let { groupId, groups, onadd, onclose }: {
		groupId: string;
		groups: DraftGroup[];
		onadd: (stock: StockSearchResult, lots: number) => void;
		onclose: () => void;
	} = $props();
	let open = $state(true);
	let query = $state('');
	let results = $state<StockSearchResult[]>([]);
	let selectedCode = $state('');
	let lots = $state<number | undefined>(1);
	let loading = $state(false);
	let searched = $state(false);
	let error = $state('');
	let request = 0;
	onDestroy(() => request++);
	const group = $derived(groups.find(({ id }) => id === groupId));
	const selected = $derived(results.find(({ code }) => code === selectedCode));
	const validLots = $derived(lots !== undefined && Number.isSafeInteger(lots) && lots >= 1);
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
	function owner(code: string) {
		return groups.find((item) => item.id !== groupId && item.stocks.some((stock) => stock.code === code));
	}
	function clearSearch() {
		request++;
		results = []; selectedCode = ''; searched = false; loading = false; error = '';
	}
	async function search(event: SubmitEvent) {
		event.preventDefault();
		if (!query.trim()) return;
		const token = ++request;
		loading = true; error = ''; selectedCode = ''; results = []; searched = false;
		try {
			const found = await searchStocks(query);
			if (token === request) { results = found; searched = true; }
		} catch {
			if (token === request) error = 'Не удалось выполнить поиск. Попробуйте ещё раз.';
		} finally { if (token === request) loading = false; }
	}
	function add() {
		if (!selected || !validLots || lots === undefined) return;
		try { onadd(selected, lots); onclose(); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось добавить акцию.'; }
	}
</script>

<Dialog.Root bind:open onOpenChange={(value) => { if (!value) onclose(); }}>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto border border-border sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>Добавить акцию</Dialog.Title>
			<Dialog.Description>Группа «{group?.name}». Найдите акцию по названию или коду. Цены демонстрационные, за одну акцию.</Dialog.Description>
		</Dialog.Header>
		<form onsubmit={search} class="space-y-3">
			<Label for="stock-query">Название или код акции</Label>
			<div class="flex flex-wrap gap-2">
				<Input id="stock-query" class="min-w-0 flex-1 basis-48" bind:value={query} oninput={clearSearch} placeholder="Например, Россети" />
				<Button type="submit" disabled={loading || !query.trim()}>{loading ? 'Поиск…' : 'Найти'}</Button>
			</div>
		</form>
		<div aria-live="polite" aria-busy={loading}>
			{#if loading}<p class="text-sm text-muted-foreground">Поиск акций…</p>
			{:else if searched && !results.length}<p class="text-sm text-muted-foreground">Ничего не найдено. Измените поисковый запрос.</p>
			{:else if results.length}<p class="mb-3 text-xs text-muted-foreground">Найдено: {results.length}. Выберите одну акцию.</p>{/if}
			<ul class="max-h-72 space-y-2 overflow-y-auto">
				{#each results as stock (stock.code)}
					{@const other = owner(stock.code)}
					<li>
						<Button variant="outline" aria-pressed={selectedCode === stock.code} disabled={!!other} class="h-auto w-full min-w-0 flex-col items-start gap-2 whitespace-normal rounded-lg p-3 text-left aria-pressed:border-primary aria-pressed:bg-primary/10" onclick={() => { selectedCode = stock.code; error = ''; }}>
							<span class="break-words font-medium">{stock.name}</span>
							<span class="flex w-full flex-wrap justify-between gap-2 text-xs"><span>{stock.code}</span><span>{currency.format(stock.price)}</span></span>
							{#if other}<span class="text-xs">Уже в группе «{other.name}»</span>{/if}
						</Button>
					</li>
				{/each}
			</ul>
		</div>
		{#if selected}
			<div class="space-y-2 rounded-lg border border-border bg-muted p-4">
				<p class="break-words text-sm font-medium">{selected.name} · {selected.code}</p>
				<p class="text-xs text-muted-foreground">Размер лота: {selected.lotSize} акций.</p>
				<Label for="add-stock-lots">Количество лотов</Label>
				<Input id="add-stock-lots" type="number" min={1} step={1} max={Number.MAX_SAFE_INTEGER} bind:value={lots} aria-invalid={!validLots} class="max-w-40 [&::-webkit-inner-spin-button]:opacity-100" />
				{#if !validLots}<p class="text-xs text-destructive">Введите целое число от 1.</p>{/if}
				{#if group?.stocks.some(({ code }) => code === selected.code)}<p class="text-xs text-muted-foreground">Лоты будут добавлены к существующей позиции.</p>{/if}
			</div>
		{/if}
		{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
		<Dialog.Footer>
			<Button variant="outline" onclick={onclose}>Отмена</Button>
			<Button disabled={!selected || !validLots || !!owner(selected.code)} onclick={add}>Добавить в группу</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
