<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Checkbox } from '#lib/components/ui/checkbox/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as RadioGroup from '#lib/components/ui/radio-group/index.js';
	import { requestRebalance, applyRebalance } from '#lib/logic/tabs/rebalance-service.js';
	import { getRebalanceEstimate, getRebalanceErrorMessage, validateRebalanceProposal, validateRebalanceRequest } from '#lib/logic/tabs/rebalance-validation.js';
	import type { RebalanceProposal, RebalanceRequest, RebalanceSnapshots } from '#lib/logic/tabs/rebalance.js';

	let { profileId, profileName, snapshots, oncommit, onbusychange }: {
		profileId: string; profileName: string; snapshots: RebalanceSnapshots;
		oncommit: (snapshots: RebalanceSnapshots) => void; onbusychange: (busy: boolean) => void;
	} = $props();
	type Mode = 'deposit' | 'withdraw' | 'unchanged';
	type Asset = 'stocks' | 'bonds';
	let options = $state<Record<Asset, { selected: boolean; mode: Mode; amount: number | undefined }>>({
		stocks: { selected: false, mode: 'unchanged', amount: 0 },
		bonds: { selected: false, mode: 'unchanged', amount: 0 }
	});
	let noSell = $state(false);
	let proposal = $state<RebalanceProposal | null>(null);
	let loading = $state(false);
	let error = $state('');
	let notice = $state('');
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
	const estimate = $derived(proposal ? getRebalanceEstimate(proposal) : null);
	const proposalError = $derived(proposal ? validateRebalanceProposal(proposal, snapshots) : null);
	$effect(() => onbusychange(loading || proposal !== null));

	function selectedAmount(asset: Asset): number | null {
		const option = options[asset];
		if (!option.selected) return null;
		if (option.mode === 'unchanged') return 0;
		if (option.amount === undefined || option.amount < 0 || !Number.isFinite(option.amount)) return NaN;
		return option.mode === 'withdraw' ? -option.amount : option.amount;
	}
	async function calculate(event: SubmitEvent) {
		event.preventDefault();
		if (loading || proposal) return;
		error = ''; notice = '';
		const request: RebalanceRequest = { profileId, stocks: selectedAmount('stocks'), bonds: selectedAmount('bonds'), noSell };
		const invalid = validateRebalanceRequest(request, snapshots);
		if (invalid) { error = invalid; return; }
		loading = true;
		try { proposal = await requestRebalance(request, snapshots); }
		catch (cause) { error = getRebalanceErrorMessage(cause, 'Не удалось получить предложение. Попробуйте ещё раз.'); }
		finally { loading = false; }
	}
	async function apply() {
		if (!proposal || loading) return;
		error = '';
		const invalid = validateRebalanceProposal(proposal, snapshots);
		if (invalid) { error = invalid; return; }
		loading = true;
		try {
			const updated = await applyRebalance(proposal, snapshots);
			// TEMP
			oncommit(updated); proposal = null; notice = 'Демонстрационные изменения применены.';
		} catch (cause) { error = getRebalanceErrorMessage(cause, 'Не удалось применить изменения. Попробуйте ещё раз.'); }
		finally { loading = false; }
	}
</script>

<section aria-labelledby="rebalance-heading" class="mx-auto max-w-6xl">
	<header class="mb-8 border-b border-border pb-6">
		<h1 id="rebalance-heading" class="text-xl font-semibold tracking-tight sm:text-2xl">Ребалансировка</h1>
		<p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
	</header>
	<div class="text-card-foreground">
		<!-- TEMP -->
		<div class="mb-6 flex flex-wrap items-center justify-between gap-3"><h2 class="text-base font-medium">Параметры</h2><span class="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">Демонстрационные данные</span></div>
		<p class="mb-6 text-sm leading-relaxed text-muted-foreground">Выберите активы и укажите сумму пополнения или вывода для каждого вида отдельно.</p>
		<form onsubmit={calculate} class="space-y-6">
			<fieldset disabled={loading || !!proposal} class="min-w-0 space-y-4">
				<legend class="sr-only">Параметры ребалансировки</legend>
				{#each [{ key: 'stocks', label: 'Акции' }, { key: 'bonds', label: 'Облигации' }] as asset (asset.key)}
					{@const key = asset.key as Asset}
					<div class="rounded-lg border border-border bg-muted/30 p-4">
						<div class="flex items-center gap-3"><Checkbox id={`rebalance-${key}`} bind:checked={options[key].selected} /><Label for={`rebalance-${key}`}>{asset.label}</Label></div>
						<fieldset disabled={!options[key].selected || loading || !!proposal} class="mt-4 min-w-0 space-y-3 disabled:opacity-50">
							<legend class="sr-only">Параметры: {asset.label}</legend>
							<RadioGroup.Root aria-label={`Направление: ${asset.label}`} bind:value={options[key].mode} class="flex flex-wrap gap-x-5 gap-y-3">
								{#each [{ value: 'deposit', label: 'Внести' }, { value: 'withdraw', label: 'Вывести' }, { value: 'unchanged', label: 'Без изменения суммы' }] as mode (mode.value)}
									<div class="flex min-w-0 items-center gap-2"><RadioGroup.Item id={`${key}-${mode.value}`} value={mode.value} disabled={!options[key].selected || loading || !!proposal} /><Label class="min-w-0 leading-normal" for={`${key}-${mode.value}`}>{mode.label}</Label></div>
								{/each}
							</RadioGroup.Root>
							<div class="flex flex-wrap items-center gap-3"><Label for={`amount-${key}`}>Сумма, ₽</Label><Input id={`amount-${key}`} type="number" min={0} step="0.01" bind:value={options[key].amount} disabled={!options[key].selected || options[key].mode === 'unchanged' || loading || !!proposal} class="w-40 max-w-full" /></div>
						</fieldset>
					</div>
				{/each}
				<div class="flex items-center gap-3"><Checkbox id="rebalance-no-sell" bind:checked={noSell} disabled={loading || !!proposal} /><Label for="rebalance-no-sell">Не продавать</Label></div>
			</fieldset>
			{#if error}<p role="alert" class="break-words text-sm text-destructive">{error}</p>{/if}
			{#if !proposal}<Button type="submit" disabled={loading} class="h-auto min-h-9 max-w-full whitespace-normal py-2">{loading ? 'Расчёт…' : 'Провести ребалансировку'}</Button>{/if}
		</form>
		<p role="status" class="mt-4 text-sm text-primary">{notice}</p>
		{#if proposal}
			<div class="mt-6 flex flex-wrap gap-3"><Button class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={loading || !!proposalError} onclick={apply}>{loading ? 'Применение…' : 'Применить результаты'}</Button><Button variant="outline" disabled={loading} onclick={() => { proposal = null; error = ''; }}>Отменить</Button></div>
			{#if proposalError}<p role="alert" class="mt-3 text-sm text-destructive">{proposalError}</p>{/if}
			<div class="mt-6">
				<h2 class="text-base font-semibold">Предложение</h2>
				<!-- TEMP -->
				<p class="mt-2 text-xs leading-relaxed text-muted-foreground">Пример предложения, не расчёт сервера. Суммы приблизительны и могут отличаться от указанного пополнения или вывода. Брокерские операции не выполняются.</p>
				<dl class="mt-4 grid gap-3 rounded-lg bg-muted p-4 text-sm sm:grid-cols-3">
					<div><dt class="text-muted-foreground">Изменение стоимости портфеля</dt><dd class="mt-1 font-medium tabular-nums">{estimate ? currency.format(estimate.total) : 'Нет данных'}</dd></div>
					<div><dt class="text-muted-foreground">Акции</dt><dd class="mt-1 tabular-nums">{estimate?.stocks != null ? currency.format(estimate.stocks) : '—'}</dd></div>
					<div><dt class="text-muted-foreground">Облигации</dt><dd class="mt-1 tabular-nums">{estimate?.bonds != null ? currency.format(estimate.bonds) : '—'}</dd></div>
				</dl>
				<div class="mt-6 grid items-start gap-6 lg:grid-cols-2">
					{#each [{ key: 'buy', label: 'Купить' }, { key: 'sell', label: 'Продать' }] as side (side.key)}
						<section aria-label={side.label} class="min-w-0 rounded-lg border border-border">
							<h3 class="border-b border-border bg-primary/10 px-4 py-3 font-medium">{side.label}</h3>
							{#each [{ key: 'stocks', label: 'Акции', unit: 'лотов', priceUnit: 'лот' }, { key: 'bonds', label: 'Облигации', unit: 'облигаций', priceUnit: 'облигация' }] as asset (asset.key)}
								<div class="border-b border-border p-4 last:border-b-0">
									<h4 class="mb-3 text-sm font-medium">{asset.label}</h4>
									{#each proposal.trades.filter((trade) => trade.side === side.key && trade.asset === asset.key) as trade (trade.id)}
										<div class="mb-3 flex flex-wrap items-center gap-3 rounded-lg bg-muted p-3 last:mb-0">
											<div class="min-w-0 flex-1 basis-32"><p class="break-words text-sm font-medium">{trade.name}</p><p class="mt-1 break-all text-xs text-muted-foreground">{trade.code} · {currency.format(trade.unitPrice)} / {asset.priceUnit}</p>{#if trade.side === 'sell'}<p class="mt-1 text-xs text-muted-foreground">Доступно: {trade.availableQuantity}</p>{/if}</div>
											<div class="w-28"><Label for={`trade-${trade.asset}-${trade.side}-${trade.code}`} class="sr-only">{side.label}: {trade.code}, {asset.unit}</Label><Input id={`trade-${trade.asset}-${trade.side}-${trade.code}`} type="number" min={0} max={trade.side === 'sell' ? trade.availableQuantity : Number.MAX_SAFE_INTEGER - trade.availableQuantity} step={1} bind:value={trade.quantity} aria-invalid={trade.quantity === undefined || !Number.isSafeInteger(trade.quantity) || trade.quantity < 0 || (trade.side === 'sell' && trade.quantity > trade.availableQuantity)} disabled={loading} class="text-center [&::-webkit-inner-spin-button]:opacity-100" /></div>
										</div>
									{:else}<p class="text-xs text-muted-foreground">Нет операций</p>{/each}
								</div>
							{/each}
						</section>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</section>
