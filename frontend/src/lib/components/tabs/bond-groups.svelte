<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { BondPortfolioSnapshot } from '#lib/logic/tabs/bonds.js';
	import { masonryGroups } from '#lib/logic/tabs/group-layout.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import BondCouponsDialog from './bond-coupons-dialog.svelte';
	import { convertMoney, formatMoney, multiplyMoney, quoteBond, sampleRates } from '#lib/logic/tabs/currency-service.js';
	import type { BondHolding } from '#lib/logic/tabs/bonds.js';
	let { portfolio, profileName, controls, includeAccrued = $bindable(true), inRub = $bindable(false) }: { portfolio: BondPortfolioSnapshot; profileName: string; controls?: Snippet; includeAccrued?: boolean; inRub?: boolean } = $props();
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
	const number = new Intl.NumberFormat('ru-RU');
	const percent = new Intl.NumberFormat('ru-RU', { style: 'unit', unit: 'percent', maximumFractionDigits: 2 });
	const date = new Intl.DateTimeFormat('ru-RU', { timeZone: 'UTC' });
	let selectedBond = $state<BondHolding | null>(null);
	function rubles(value: number | null) { return value === null ? 'Нет данных' : currency.format(value); }
</script>

<section aria-labelledby="bonds-heading" class="mx-auto max-w-6xl">
	<header class="mb-8 flex flex-wrap items-start justify-between gap-4 border-b border-border pb-6">
		<div><h1 id="bonds-heading" class="text-xl font-semibold tracking-tight sm:text-2xl">Облигации</h1><p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p></div>
		<div class="space-y-3">
			<div role="group" aria-label="Цена облигации" class="flex flex-wrap gap-2"><Button size="sm" variant={includeAccrued ? 'outline' : 'secondary'} aria-pressed={!includeAccrued} onclick={() => includeAccrued = false}>Чистая цена</Button><Button size="sm" variant={includeAccrued ? 'secondary' : 'outline'} aria-pressed={includeAccrued} onclick={() => includeAccrued = true}>С НКД</Button></div>
			<div role="group" aria-label="Валюта цены облигации" class="flex flex-wrap gap-2"><Button size="sm" variant={inRub ? 'outline' : 'secondary'} aria-pressed={!inRub} onclick={() => inRub = false}>Исходная валюта</Button><Button size="sm" variant={inRub ? 'secondary' : 'outline'} aria-pressed={inRub} onclick={() => inRub = true}>В рублях</Button></div>
		</div>
	</header>
	{@render controls?.()}
	<div class="mb-8 flex flex-wrap items-start justify-between gap-4 rounded-xl border border-border bg-card p-4 text-card-foreground sm:p-6">
		<div><p class="text-sm text-muted-foreground">Общая стоимость облигаций, с НКД</p><p class="mt-2 text-2xl font-semibold tracking-tight tabular-nums sm:text-3xl">{rubles(portfolio.totalBondsValueRub)}</p></div>
		<!-- TEMP -->
		<span class="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">Демонстрационные данные</span>
	</div>
	<div class="mb-6 space-y-2">
		<h2 class="text-base font-medium">Группы облигаций</h2>
		<p class="text-xs leading-relaxed text-muted-foreground">Группы и доли независимы от акций. Все доли указаны от общей стоимости облигаций в профиле, целевая доля внутри группы распределяется поровну.</p>
		<!-- TEMP -->
		<p class="text-xs leading-relaxed text-muted-foreground">Итоги групп и доли рассчитаны в рублях с НКД и не зависят от настроек отображения. Цены и курсы демонстрационные, не являются текущими котировками.</p>
		<!-- TEMP -->
		<p class="text-xs leading-relaxed text-muted-foreground">Демонстрационные курсы: 1 USD = {sampleRates.USD} RUB · 1 EUR = {sampleRates.EUR} RUB · 1 CNY = {sampleRates.CNY} RUB.</p>
	</div>
	<div use:masonryGroups class="group-layout grid min-w-0 items-start gap-6 xl:grid-cols-2">
		{#each portfolio.groups as group (group.id)}
			<section aria-label={`Группа облигаций: ${group.name}`} class="min-w-0 overflow-hidden rounded-xl border border-border bg-card text-card-foreground">
				<header class="border-b border-border bg-primary/10 p-4 sm:p-6">
					<div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="break-words text-lg font-semibold">{group.name}</h3><p class="text-base font-medium tabular-nums"><span class="sr-only">Стоимость группы: </span>{rubles(group.totalValueRub)}</p></div>
					<dl class="mt-4 grid gap-3 sm:grid-cols-2">
						<div class="flex flex-wrap items-baseline gap-x-2"><dt class="text-xs text-muted-foreground">Целевая доля</dt><dd class="text-sm font-medium tabular-nums">{group.targetPercent === null ? 'Не задана' : percent.format(group.targetPercent)}</dd></div>
						<div class="flex flex-wrap items-baseline gap-x-2"><dt class="text-xs text-muted-foreground">Текущая доля</dt><dd class="text-sm font-medium text-primary tabular-nums">{group.actualPercent === null ? 'Нет данных' : percent.format(group.actualPercent)}</dd></div>
					</dl>
				</header>
				<ul class="divide-y divide-border">
					{#each group.bonds as bond (bond.code)}
						<li class="p-4 odd:bg-card even:bg-muted sm:p-6">
							<div class="mb-4 flex flex-wrap items-baseline justify-between gap-2"><h4 class="min-w-0 break-words text-sm font-semibold">{bond.name}</h4><span class="break-all text-xs text-muted-foreground">{bond.code}{bond.isin ? ` · ${bond.isin}` : ''}</span></div>
							<dl class="grid gap-x-4 gap-y-4 sm:grid-cols-3 [&>div]:min-w-0 [&_dd]:wrap-anywhere">
								<div><dt class="text-xs text-muted-foreground">{includeAccrued ? 'Цена с НКД' : 'Чистая цена'}</dt><dd class="mt-1 text-sm tabular-nums">{formatMoney(quoteBond(bond, includeAccrued, inRub))}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Количество облигаций</dt><dd class="mt-1 text-sm tabular-nums">{number.format(bond.lotCount * bond.lotSize)}{#if bond.lotSize !== 1}<span class="mt-1 block text-xs text-muted-foreground">Размер лота: {number.format(bond.lotSize)}</span>{/if}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Целевая доля</dt><dd class="mt-1 text-sm tabular-nums">{group.targetPercent === null || bond.targetPercent === null ? 'Не задана' : percent.format(bond.targetPercent)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Общая стоимость</dt><dd class="mt-1 text-sm font-medium tabular-nums">{formatMoney(multiplyMoney(quoteBond(bond, includeAccrued, inRub), bond.lotCount * bond.lotSize))}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Дата погашения</dt><dd class="mt-1 text-sm tabular-nums">{bond.maturityDate === null ? 'Нет данных' : date.format(new Date(bond.maturityDate))}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Текущая доля</dt><dd class="mt-1 text-sm text-primary tabular-nums">{bond.actualPercent === null ? 'Нет данных' : percent.format(bond.actualPercent)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Выплата при погашении</dt><dd class="mt-1 text-sm tabular-nums">{formatMoney(inRub ? convertMoney(bond.maturityPayment, 'RUB') : bond.maturityPayment)}</dd></div>
							</dl>
							<Button variant="outline" size="sm" class="mt-4" aria-label={`Купоны ${bond.code}`} onclick={() => selectedBond = bond}>Купоны</Button>
						</li>
					{:else}<li class="p-6 text-sm text-muted-foreground">В этой группе пока нет облигаций.</li>{/each}
				</ul>
			</section>
		{:else}<div data-group-empty class="rounded-xl border border-dashed border-border p-6 text-center sm:p-12 xl:col-span-2"><p class="text-base font-medium">В профиле пока нет облигаций</p><p class="mt-2 text-sm text-muted-foreground">Добавьте группы и облигации через редактирование состава.</p></div>{/each}
	</div>
</section>

{#if selectedBond}
	<BondCouponsDialog bondCode={selectedBond.code} bondName={selectedBond.name} quantity={selectedBond.lotCount * selectedBond.lotSize} onclose={() => selectedBond = null} />
{/if}
