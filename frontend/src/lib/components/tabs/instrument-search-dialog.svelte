<script lang="ts" generics="T extends { code: string; name: string }">
	import { onDestroy, type Snippet } from 'svelte';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';

	let { groupName, instrument, queryLabel, placeholder, quantityLabel, search, owner, alreadyHeld, onadd, onclose, details, displayPrice }: {
		groupName: string; instrument: string; queryLabel: string; placeholder: string; quantityLabel: string;
		search: (query: string) => Promise<T[]>;
		owner: (code: string) => string | undefined;
		alreadyHeld: (code: string) => boolean;
		onadd: (result: T, quantity: number) => void;
		onclose: () => void;
		details?: Snippet<[T]>;
		displayPrice: (result: T) => string;
	} = $props();
	const id = $props.id();
	let open = $state(true);
	let query = $state('');
	let results = $state<T[]>([]);
	let selectedCode = $state('');
	let quantity = $state<number | undefined>(1);
	let loading = $state(false);
	let searched = $state(false);
	let error = $state('');
	let request = 0;
	onDestroy(() => request++);
	const selected = $derived(results.find(({ code }) => code === selectedCode));
	const validQuantity = $derived(quantity !== undefined && Number.isSafeInteger(quantity) && quantity >= 1);
	function clearSearch() {
		request++; results = []; selectedCode = ''; searched = false; loading = false; error = '';
	}
	async function find(event: SubmitEvent) {
		event.preventDefault();
		if (!query.trim()) return;
		const token = ++request;
		loading = true; error = ''; selectedCode = ''; results = []; searched = false;
		try {
			const found = await search(query);
			if (token === request) { results = found; searched = true; }
		} catch { if (token === request) error = 'Не удалось выполнить поиск. Попробуйте ещё раз.'; }
		finally { if (token === request) loading = false; }
	}
	function add() {
		if (!selected || !validQuantity || quantity === undefined) return;
		try { onadd(selected, quantity); onclose(); }
		catch (cause) { error = cause instanceof Error ? cause.message : `Не удалось добавить ${instrument}.`; }
	}
</script>

<Dialog.Root bind:open onOpenChange={(value) => { if (!value) onclose(); }}>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto border border-border sm:max-w-2xl">
		<!-- TEMP -->
		<Dialog.Header><Dialog.Title>Добавить {instrument}</Dialog.Title><Dialog.Description>Группа «{groupName}». Найдите {instrument} по названию или коду. Цены демонстрационные, за одну ценную бумагу.</Dialog.Description></Dialog.Header>
		<form onsubmit={find} class="space-y-3">
			<Label for={`instrument-query-${id}`}>{queryLabel}</Label>
			<div class="flex flex-wrap gap-2"><Input id={`instrument-query-${id}`} class="min-w-0 flex-1 basis-48" bind:value={query} oninput={clearSearch} {placeholder} /><Button type="submit" disabled={loading || !query.trim()}>{loading ? 'Поиск…' : 'Найти'}</Button></div>
		</form>
		<div aria-live="polite" aria-busy={loading}>
			{#if loading}<p class="text-sm text-muted-foreground">Поиск…</p>
			{:else if searched && !results.length}<p class="text-sm text-muted-foreground">Ничего не найдено. Измените поисковый запрос.</p>
			{:else if results.length}<p class="mb-3 text-xs text-muted-foreground">Найдено: {results.length}. Выберите одну ценную бумагу.</p>{/if}
			<ul class="max-h-72 space-y-2 overflow-y-auto">
				{#each results as result (result.code)}
					{@const other = owner(result.code)}
					<li><Button variant="outline" aria-pressed={selectedCode === result.code} disabled={!!other} class="h-auto w-full min-w-0 flex-col items-start gap-2 whitespace-normal rounded-lg p-3 text-left aria-pressed:border-primary aria-pressed:bg-primary/10" onclick={() => { selectedCode = result.code; error = ''; }}>
						<span class="break-words font-medium">{result.name}</span>
						<span class="flex w-full flex-wrap justify-between gap-2 text-xs"><span class="break-all">{result.code}</span><span>{displayPrice(result)}</span></span>
						{#if other}<span class="text-xs">Уже в группе «{other}»</span>{/if}
					</Button></li>
				{/each}
			</ul>
		</div>
		{#if selected}
			<div class="space-y-2 rounded-lg border border-border bg-muted p-4">
				<p class="break-words text-sm font-medium">{selected.name} · {selected.code}</p>
				{@render details?.(selected)}
				<Label for={`instrument-quantity-${id}`}>{quantityLabel}</Label>
				<Input id={`instrument-quantity-${id}`} type="number" min={1} step={1} max={Number.MAX_SAFE_INTEGER} bind:value={quantity} aria-invalid={!validQuantity} class="max-w-40 [&::-webkit-inner-spin-button]:opacity-100" />
				{#if !validQuantity}<p class="text-xs text-destructive">Введите целое число от 1.</p>{/if}
				{#if alreadyHeld(selected.code)}<p class="text-xs text-muted-foreground">Количество будет добавлено к существующей позиции.</p>{/if}
			</div>
		{/if}
		{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
		<Dialog.Footer><Button variant="outline" onclick={onclose}>Отмена</Button><Button disabled={!selected || !validQuantity || !!owner(selected.code)} onclick={add}>Добавить в группу</Button></Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
