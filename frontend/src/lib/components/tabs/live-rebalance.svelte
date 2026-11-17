<script lang="ts">
	import { onDestroy, untrack } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Checkbox } from '#lib/components/ui/checkbox/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { rebalanceLotValue, rebalanceTotals, type LiveTrade } from '#lib/logic/live-rebalance.js';
	import { usePortfolioBack } from '#lib/logic/portfolio-back.js';

	let { profileName, rates, request, record, onbusychange }: {
		profileName: string;
		rates: Readonly<Record<string, number>>;
		request: (options: { stocks: number | null; bonds: number | null; noSell: boolean }) => Promise<LiveTrade[]>;
		record: (trades: LiveTrade[]) => Promise<void>;
		onbusychange: (busy: boolean) => void;
	} = $props();
	let stocks = $state(false), bonds = $state(false), noSell = $state(false);
	let stockAmount = $state<number | undefined>(0), bondAmount = $state<number | undefined>(0);
	let trades = $state<LiveTrade[] | null>(null), loading = $state(false), blocked = $state(false), confirm = $state(false);
	type Row = { index: number; side: 'buy' | 'sell'; quantity: number | undefined };
	let rows = $state<Row[]>([]);
	let error = $state(''), notice = $state('');
	const back = usePortfolioBack();
	$effect(() => { onbusychange(loading || trades !== null); });
	$effect(() => { if (loading || trades !== null) return untrack(() => back?.enter('rebalance', () => { if (!loading && !blocked) { trades = null; error = ''; } })); });
	onDestroy(() => onbusychange(false));
	const invalid = $derived(rows.some(row => {
		const holding = trades?.[row.index]?.holding;
		return !holding || row.quantity === undefined || !Number.isSafeInteger(row.quantity) || row.quantity < 0 || (row.side === 'sell' && row.quantity > holding.lot_amount) || (row.side === 'buy' && !Number.isSafeInteger(holding.lot_amount + row.quantity));
	}));
	const changed = $derived(rows.some(row => (row.quantity ?? 0) > 0));
	const proposed = $derived(trades?.map((trade, index) => {
		const row = rows.find(item => item.index === index);
		const quantity = row?.quantity ?? 0;
		return { ...trade, target: trade.holding.lot_amount + (row?.side === 'sell' ? -quantity : quantity) };
	}) ?? []);
	const totals = $derived(rebalanceTotals(proposed, rates));
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
	const signedCurrency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2, signDisplay: 'exceptZero' });
	const money = (value: number | null) => value === null ? 'Нет данных' : currency.format(value);
	const signedMoney = (value: number | null) => value === null ? 'Нет данных' : signedCurrency.format(value);
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
		try {
			const result = await request({ stocks: stocks ? stockAmount! : null, bonds: bonds ? bondAmount! : null, noSell });
			trades = result;
			rows = result.flatMap((trade, index) => trade.target === trade.holding.lot_amount ? [] : [{ index, side: trade.target > trade.holding.lot_amount ? 'buy' as const : 'sell' as const, quantity: Math.abs(trade.target - trade.holding.lot_amount) }]);
		}
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось получить предложение.'; }
		finally { loading = false; }
	}
	async function apply() {
		if (!trades || invalid || loading || blocked) return;
		confirm = false; loading = true; error = '';
		try { await record(proposed); trades = null; notice = 'Операции записаны в портфель.'; }
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
			<p class="text-sm text-muted-foreground">Количество операций указано в лотах. Суммы приблизительны, рассчитаны по загруженным ценам и курсам; цены облигаций включают НКД.</p>
			<dl class="grid gap-3 rounded-lg bg-muted p-4 text-sm sm:grid-cols-3">
				<div><dt class="text-muted-foreground">Изменение стоимости портфеля</dt><dd class="mt-1 font-medium tabular-nums">{signedMoney(invalid ? null : totals.total)}</dd></div>
				<div><dt class="text-muted-foreground">Акции</dt><dd class="mt-1 tabular-nums">{stocks ? signedMoney(invalid ? null : totals.stocks) : '—'}</dd></div>
				<div><dt class="text-muted-foreground">Облигации</dt><dd class="mt-1 tabular-nums">{bonds ? signedMoney(invalid ? null : totals.bonds) : '—'}</dd></div>
			</dl>
			<div class="grid items-start gap-6 lg:grid-cols-2">
				{#each [{ key: 'buy', label: 'Купить' }, { key: 'sell', label: 'Продать' }] as side (side.key)}
					<section aria-label={side.label} class="min-w-0 rounded-lg border border-border">
						<h3 class="border-b border-border bg-primary/10 px-4 py-3 font-medium">{side.label}</h3>
						{#each [{ key: 'stock', label: 'Акции' }, { key: 'bond', label: 'Облигации' }] as asset (asset.key)}
							<div class="border-b border-border p-4 last:border-b-0"><h4 class="mb-3 text-sm font-medium">{asset.label}</h4>
								{#each rows.filter(row => row.side === side.key && trades?.[row.index]?.kind === asset.key) as row (row.index)}
									{@const trade = trades[row.index]}
									<div class="mb-3 flex flex-wrap items-center gap-3 rounded-lg bg-muted p-3 last:mb-0">
										<div class="min-w-0 flex-1 basis-32"><p class="break-words text-sm font-medium">{trade.holding.name}</p><p class="mt-1 break-all text-xs text-muted-foreground">{trade.holding.code} · {money(rebalanceLotValue(trade, rates))} / лот</p>{#if trade.holding.lot_size > 1}<p class="mt-1 text-xs text-muted-foreground">В лоте: {trade.holding.lot_size}</p>{/if}{#if row.side === 'sell'}<p class="mt-1 text-xs text-muted-foreground">Доступно: {trade.holding.lot_amount} лотов</p>{/if}</div>
										<div class="w-28"><Label for={`live-trade-${row.index}`} class="text-xs text-muted-foreground"><span class="sr-only">{side.label}: {trade.holding.code}, </span>Количество, лотов</Label><Input id={`live-trade-${row.index}`} type="number" min={0} max={row.side === 'sell' ? trade.holding.lot_amount : Number.MAX_SAFE_INTEGER - trade.holding.lot_amount} step={1} bind:value={row.quantity} disabled={loading || blocked} /></div>
									</div>
								{:else}<p class="text-xs text-muted-foreground">Нет операций</p>{/each}
							</div>
						{/each}
					</section>
				{/each}
			</div>
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
