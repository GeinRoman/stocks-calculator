<script lang="ts">
	import { onDestroy, untrack } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Checkbox } from '#lib/components/ui/checkbox/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import type { LiveTrade } from '#lib/logic/live-rebalance.js';
	import { usePortfolioBack } from '#lib/logic/portfolio-back.js';

	let { profileName, request, record, onbusychange }: {
		profileName: string;
		request: (options: { stocks: number | null; bonds: number | null; noSell: boolean }) => Promise<LiveTrade[]>;
		record: (trades: LiveTrade[]) => Promise<void>;
		onbusychange: (busy: boolean) => void;
	} = $props();
	let stocks = $state(false), bonds = $state(false), noSell = $state(false);
	let stockAmount = $state<number | undefined>(0), bondAmount = $state<number | undefined>(0);
	let trades = $state<LiveTrade[] | null>(null), loading = $state(false), blocked = $state(false), confirm = $state(false);
	let error = $state(''), notice = $state('');
	const back = usePortfolioBack();
	$effect(() => { onbusychange(loading || trades !== null); });
	$effect(() => { if (loading || trades !== null) return untrack(() => back?.enter('rebalance', () => { if (!loading && !blocked) { trades = null; error = ''; } })); });
	onDestroy(() => onbusychange(false));
	const invalid = $derived(trades?.some(trade => !Number.isSafeInteger(trade.target) || trade.target < 0 || (noSell && trade.target < trade.holding.lot_amount)) ?? false);
	const changed = $derived(trades?.some(trade => trade.target !== trade.holding.lot_amount) ?? false);
	async function calculate(event: SubmitEvent) {
		event.preventDefault();
		if (loading || trades) return;
		error = ''; notice = '';
		if (!stocks && !bonds) { error = 'Выберите хотя бы один вид активов.'; return; }
		for (const amount of [stocks ? stockAmount : null, bonds ? bondAmount : null]) {
			if (amount !== null && (amount === undefined || !Number.isFinite(amount) || (noSell && amount <= 0))) {
				error = noSell ? 'Без продажи укажите положительное пополнение для каждого выбранного вида активов.' : 'Укажите корректное изменение суммы.';
				return;
			}
		}
		loading = true;
		try { trades = await request({ stocks: stocks ? stockAmount! : null, bonds: bonds ? bondAmount! : null, noSell }); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось получить предложение.'; }
		finally { loading = false; }
	}
	async function apply() {
		if (!trades || invalid || loading || blocked) return;
		confirm = false; loading = true; error = '';
		try { await record(trades); trades = null; notice = 'Операции записаны в портфель.'; }
		catch (cause) { blocked = true; error = cause instanceof Error ? cause.message : 'Не удалось записать операции. Обновите страницу.'; }
		finally { loading = false; }
	}
</script>

<section aria-labelledby="live-rebalance-heading" class="mx-auto max-w-6xl">
	<header class="sticky top-0 z-20 mb-6 border-b border-border bg-background py-3">
		<h1 id="live-rebalance-heading" class="text-xl font-semibold sm:text-2xl">Ребалансировка</h1>
		<p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
	</header>
	<p class="mb-6 text-sm text-muted-foreground">Сервер рассчитает примерное количество лотов. Предложение не отправляет поручения брокеру и не меняет портфель.</p>
	<form onsubmit={calculate} class="space-y-4">
		<fieldset disabled={loading || trades !== null || blocked} class="space-y-4">
			<legend class="sr-only">Параметры ребалансировки</legend>
			<div class="space-y-3 rounded-lg border border-border bg-card p-4">
				<div class="flex items-center gap-3"><Checkbox id="live-rebalance-stocks" bind:checked={stocks} /><Label for="live-rebalance-stocks">Акции</Label></div>
				<div class="flex flex-wrap items-center gap-3"><Label for="live-rebalance-stock-amount">Изменение суммы, ₽</Label><Input id="live-rebalance-stock-amount" type="number" step="0.01" bind:value={stockAmount} disabled={!stocks} class="w-40" /></div>
			</div>
			<div class="space-y-3 rounded-lg border border-border bg-card p-4">
				<div class="flex items-center gap-3"><Checkbox id="live-rebalance-bonds" bind:checked={bonds} /><Label for="live-rebalance-bonds">Облигации</Label></div>
				<div class="flex flex-wrap items-center gap-3"><Label for="live-rebalance-bond-amount">Изменение суммы, ₽</Label><Input id="live-rebalance-bond-amount" type="number" step="0.01" bind:value={bondAmount} disabled={!bonds} class="w-40" /></div>
			</div>
			<p class="text-xs text-muted-foreground">Положительная сумма — пополнение, отрицательная — вывод, ноль — без изменения суммы.</p>
			<div class="flex items-center gap-3"><Checkbox id="live-rebalance-no-sell" bind:checked={noSell} /><Label for="live-rebalance-no-sell">Не продавать</Label></div>
		</fieldset>
		{#if !trades}<Button type="submit" disabled={loading || blocked}>{loading ? 'Расчёт…' : 'Рассчитать предложение'}</Button>{/if}
	</form>
	{#if error}<p role="alert" class="mt-4 text-sm text-destructive">{error}</p>{/if}
	{#if notice}<p role="status" class="mt-4 text-sm text-primary">{notice}</p>{/if}
	{#if blocked}<div class="mt-4"><Button variant="outline" onclick={() => window.location.reload()}>Обновить страницу</Button></div>{/if}
	{#if trades}
		<section aria-labelledby="live-rebalance-proposal" class="mt-6 space-y-4">
			<h2 id="live-rebalance-proposal" class="text-base font-semibold">Предложение</h2>
			<p class="text-sm text-muted-foreground">Количество указано в лотах. Изменения приблизительны; вы можете скорректировать итоговое количество.</p>
			{#each trades as trade (`${trade.kind}-${trade.holding.code}`)}
				<div class="flex flex-wrap items-center gap-4 rounded-lg border border-border bg-card p-4">
					<div class="min-w-0 flex-1"><p class="break-words font-medium">{trade.holding.name} · {trade.holding.code}</p><p class="text-xs text-muted-foreground">{trade.kind === 'stock' ? 'Акции' : 'Облигации'} · {trade.holding.group_name} · Сейчас: {trade.holding.lot_amount} лотов · {trade.target > trade.holding.lot_amount ? `Купить: ${trade.target - trade.holding.lot_amount}` : trade.target < trade.holding.lot_amount ? `Продать: ${trade.holding.lot_amount - trade.target}` : 'Без изменений'}</p></div>
					<div class="w-32"><Label for={`live-target-${trade.kind}-${trade.holding.code}`}>Итого лотов</Label><Input id={`live-target-${trade.kind}-${trade.holding.code}`} type="number" min={noSell ? trade.holding.lot_amount : 0} step={1} bind:value={trade.target} disabled={loading || blocked} /></div>
				</div>
			{/each}
			{#if invalid}<p role="alert" class="text-sm text-destructive">Укажите целое неотрицательное количество лотов; при запрете продажи уменьшать количество нельзя.</p>{/if}
			<div class="flex flex-wrap gap-3"><Button disabled={loading || blocked || invalid || !changed} onclick={() => confirm = true}>Записать операции</Button><Button variant="outline" disabled={loading || blocked} onclick={() => { trades = null; error = ''; }}>Закрыть предложение</Button></div>
		</section>
	{/if}
</section>

<AlertDialog.Root bind:open={confirm}>
	<AlertDialog.Content>
		<AlertDialog.Header><AlertDialog.Title>Записать операции в портфель?</AlertDialog.Title><AlertDialog.Description>Продолжайте только если сделки уже выполнены. Это изменит количества и историю операций в приложении, но не отправит поручения брокеру. В истории будут использованы цены из загруженного портфеля; операции записываются по отдельности и частично сохранённые изменения нельзя отменить автоматически.</AlertDialog.Description></AlertDialog.Header>
		<AlertDialog.Footer><AlertDialog.Cancel disabled={loading}>Отмена</AlertDialog.Cancel><Button disabled={loading} onclick={apply}>Записать операции</Button></AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
