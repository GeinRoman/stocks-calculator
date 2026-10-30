<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { BondPortfolioSnapshot } from '#lib/logic/tabs/bonds.js';
	let { portfolio, profileName, controls }: { portfolio: BondPortfolioSnapshot; profileName: string; controls?: Snippet } = $props();
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
	const number = new Intl.NumberFormat('ru-RU');
	const percent = new Intl.NumberFormat('ru-RU', { style: 'unit', unit: 'percent', maximumFractionDigits: 2 });
	const date = new Intl.DateTimeFormat('ru-RU', { timeZone: 'UTC' });
	const timestamp = new Intl.DateTimeFormat('ru-RU', { dateStyle: 'short', timeStyle: 'short', timeZone: 'UTC' });
	function money(value: number | null) { return value === null ? 'Нет данных' : currency.format(value); }
	function count(value: number | null) { return value === null ? 'Нет данных' : number.format(value); }
</script>

<section aria-labelledby="bonds-heading" class="mx-auto max-w-6xl">
	<header class="mb-8 border-b border-border pb-6">
		<h1 id="bonds-heading" class="text-xl font-semibold tracking-tight sm:text-2xl">Облигации</h1>
		<p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
	</header>
	{@render controls?.()}
	<div class="mb-8 flex flex-wrap items-start justify-between gap-4 rounded-xl border border-border bg-card p-4 text-card-foreground sm:p-6">
		<div><p class="text-sm text-muted-foreground">Общая стоимость облигаций</p><p class="mt-2 text-2xl font-semibold tracking-tight tabular-nums sm:text-3xl">{currency.format(portfolio.totalBondsValue)}</p></div>
		<span class="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">Демонстрационные данные</span>
	</div>
	<div class="mb-6 space-y-2">
		<h2 class="text-base font-medium">Группы облигаций</h2>
		<p class="text-xs leading-relaxed text-muted-foreground">Группы и доли независимы от акций. Все доли указаны от общей стоимости облигаций в профиле, целевая доля внутри группы распределяется поровну.</p>
		<p class="text-xs leading-relaxed text-muted-foreground">Стоимость — текущая цена × количество облигаций. Будущие купоны и цена погашения в неё не включены. Цены условные, не являются текущими котировками.</p>
		<p class="text-xs text-muted-foreground">Оставшиеся купоны рассчитаны на {timestamp.format(new Date(portfolio.asOf))} UTC. Неизвестные будущие выплаты не подменяются нулём.</p>
	</div>
	<div class="grid min-w-0 gap-6 xl:grid-cols-2">
		{#each portfolio.groups as group (group.id)}
			<section aria-label={`Группа облигаций: ${group.name}`} class="min-w-0 overflow-hidden rounded-xl border border-border bg-card text-card-foreground">
				<header class="border-b border-border bg-primary/10 p-4 sm:p-6">
					<div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="break-words text-lg font-semibold">{group.name}</h3><p class="text-base font-medium tabular-nums"><span class="sr-only">Стоимость группы: </span>{currency.format(group.totalValue)}</p></div>
					<dl class="mt-4 grid gap-3 sm:grid-cols-2">
						<div class="flex flex-wrap items-baseline gap-x-2"><dt class="text-xs text-muted-foreground">Целевая доля</dt><dd class="text-sm font-medium tabular-nums">{group.targetPercent === null ? 'Не задана' : percent.format(group.targetPercent)}</dd></div>
						<div class="flex flex-wrap items-baseline gap-x-2"><dt class="text-xs text-muted-foreground">Текущая доля</dt><dd class="text-sm font-medium text-primary tabular-nums">{group.actualPercent === null ? 'Нет данных' : percent.format(group.actualPercent)}</dd></div>
					</dl>
				</header>
				<ul class="divide-y divide-border">
					{#each group.bonds as bond (bond.code)}
						<li class="p-4 odd:bg-card even:bg-muted sm:p-6">
							<div class="mb-4 flex flex-wrap items-baseline justify-between gap-2"><h4 class="min-w-0 break-words text-sm font-semibold">{bond.name}</h4><span class="break-all text-xs text-muted-foreground">{bond.code}</span></div>
							<dl class="grid gap-x-4 gap-y-4 sm:grid-cols-3 [&>div]:min-w-0 [&_dd]:wrap-anywhere">
								<div><dt class="text-xs text-muted-foreground">Цена / погашение</dt><dd class="mt-1 text-sm tabular-nums">{currency.format(bond.price)} / {money(bond.redemptionPrice)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Количество облигаций</dt><dd class="mt-1 text-sm tabular-nums">{number.format(bond.quantity)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Целевая доля</dt><dd class="mt-1 text-sm tabular-nums">{group.targetPercent === null || bond.targetPercent === null ? 'Не задана' : percent.format(bond.targetPercent)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Общая стоимость</dt><dd class="mt-1 text-sm font-medium tabular-nums">{currency.format(bond.totalValue)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Дата погашения</dt><dd class="mt-1 text-sm tabular-nums">{bond.maturityDate === null ? 'Нет данных' : date.format(new Date(bond.maturityDate))}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Текущая доля</dt><dd class="mt-1 text-sm text-primary tabular-nums">{bond.actualPercent === null ? 'Нет данных' : percent.format(bond.actualPercent)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Купоны, осталось / всего</dt><dd class="mt-1 text-sm tabular-nums">{count(bond.remainingCoupons)} / {count(bond.totalCoupons)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Купон на 1 облигацию</dt><dd class="mt-1 text-sm tabular-nums">{money(bond.couponAmount)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Оставшиеся купоны на 1 облигацию</dt><dd class="mt-1 text-sm tabular-nums">{money(bond.remainingCouponValue)}</dd></div>
								<div class="sm:col-span-3"><dt class="text-xs text-muted-foreground">Оставшиеся купоны на все облигации</dt><dd class="mt-1 text-sm font-medium tabular-nums">{money(bond.holdingRemainingCouponValue)}</dd></div>
							</dl>
						</li>
					{:else}<li class="p-6 text-sm text-muted-foreground">В этой группе пока нет облигаций.</li>{/each}
				</ul>
			</section>
		{:else}<div class="rounded-xl border border-dashed border-border p-6 text-center sm:p-12 xl:col-span-2"><p class="text-base font-medium">В профиле пока нет облигаций</p><p class="mt-2 text-sm text-muted-foreground">Добавьте группы и облигации через редактирование состава.</p></div>{/each}
	</div>
</section>
