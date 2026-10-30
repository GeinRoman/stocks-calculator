<script lang="ts">
	import type { PortfolioSnapshot } from '#lib/logic/tabs/portfolio.js';
	import type { Snippet } from 'svelte';
	import { masonryGroups } from '#lib/logic/tabs/group-layout.js';

	let { portfolio, profileName, controls }: { portfolio: PortfolioSnapshot; profileName: string; controls?: Snippet } = $props();
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
	const number = new Intl.NumberFormat('ru-RU');
	const percent = new Intl.NumberFormat('ru-RU', { style: 'unit', unit: 'percent', maximumFractionDigits: 2 });
</script>

<section aria-labelledby="stocks-heading" class="mx-auto max-w-6xl">
	<header class="mb-8 border-b border-border pb-6">
		<h1 id="stocks-heading" class="text-xl font-semibold tracking-tight sm:text-2xl">Акции</h1>
		<p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
	</header>
	{@render controls?.()}

	<div class="mb-8 flex flex-wrap items-start justify-between gap-4 rounded-xl border border-border bg-card p-4 text-card-foreground sm:p-6">
		<div>
			<p class="text-sm text-muted-foreground">Общая стоимость акций</p>
			<p class="mt-2 text-2xl font-semibold tracking-tight tabular-nums sm:text-3xl">{currency.format(portfolio.totalStocksValue)}</p>
		</div>
		<!-- TEMP -->
		<span class="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">Демонстрационные данные</span>
	</div>

	<div class="mb-6">
		<h2 class="text-base font-medium">Группы (отрасли)</h2>
		<p class="mt-2 text-xs leading-relaxed text-muted-foreground">Целевая доля задаётся пользователем, текущая отражает стоимость активов. Все доли указаны от общей стоимости акций в профиле.</p>
		<!-- TEMP -->
		<p class="mt-1 text-xs leading-relaxed text-muted-foreground">Внутри группы целевая доля распределяется поровну. Цены условные, не являются текущими котировками.</p>
	</div>

	<div use:masonryGroups class="group-layout grid min-w-0 items-start gap-6 xl:grid-cols-2">
		{#each portfolio.groups as group (group.id)}
			<section aria-label={`Группа: ${group.name}`} class="min-w-0 overflow-hidden rounded-xl border border-border bg-card text-card-foreground">
				<header class="border-b border-border bg-primary/10 p-4 sm:p-6">
					<div class="flex flex-wrap items-baseline justify-between gap-2">
						<h3 class="break-words text-lg font-semibold">{group.name}</h3>
						<p class="text-base font-medium tabular-nums"><span class="sr-only">Стоимость группы: </span>{currency.format(group.totalValue)}</p>
					</div>
					<dl class="mt-4 grid gap-3 sm:grid-cols-2">
						<div class="flex flex-wrap items-baseline gap-x-2">
							<dt class="text-xs text-muted-foreground">Целевая доля</dt>
							<dd class="text-sm font-medium tabular-nums">{group.targetPercent != null ? percent.format(group.targetPercent) : 'Не задана'}</dd>
						</div>
						<div class="flex flex-wrap items-baseline gap-x-2">
							<dt class="text-xs text-muted-foreground">Текущая доля</dt>
							<dd class="text-sm font-medium text-primary tabular-nums">{group.actualPercent != null ? percent.format(group.actualPercent) : 'Нет данных'}</dd>
						</div>
					</dl>
				</header>
				<ul class="divide-y divide-border">
					{#each group.stocks as stock (stock.code)}
						<li class="p-4 odd:bg-card even:bg-muted sm:p-6">
							<div class="mb-4 flex flex-wrap items-baseline justify-between gap-2">
								<h4 class="min-w-0 break-words text-sm font-semibold">{stock.name}</h4>
								<span class="text-xs text-muted-foreground">{stock.code}</span>
							</div>
							<dl class="grid gap-x-4 gap-y-4 sm:grid-cols-3 [&>div]:min-w-0 [&_dd]:wrap-anywhere">
								<div><dt class="text-xs text-muted-foreground">Цена за акцию</dt><dd class="mt-1 text-sm tabular-nums">{currency.format(stock.price)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Количество лотов</dt><dd class="mt-1 text-sm tabular-nums">{number.format(stock.lotCount)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Целевая доля</dt><dd class="mt-1 text-sm tabular-nums">{group.targetPercent != null && stock.targetPercent != null ? percent.format(stock.targetPercent) : 'Не задана'}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Общая стоимость</dt><dd class="mt-1 text-sm font-medium tabular-nums">{currency.format(stock.totalValue)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Размер лота, акций</dt><dd class="mt-1 text-sm tabular-nums">{number.format(stock.lotSize)}</dd></div>
								<div><dt class="text-xs text-muted-foreground">Текущая доля</dt><dd class="mt-1 text-sm text-primary tabular-nums">{stock.actualPercent != null ? percent.format(stock.actualPercent) : 'Нет данных'}</dd></div>
							</dl>
						</li>
					{:else}
						<li class="p-6 text-sm text-muted-foreground">В этой группе пока нет акций.</li>
					{/each}
				</ul>
			</section>
		{:else}
			<div data-group-empty class="rounded-xl border border-dashed border-border p-6 text-center xl:col-span-2 sm:p-12">
				<p class="text-base font-medium">В профиле пока нет акций</p>
				<p class="mt-2 text-sm text-muted-foreground">Здесь появятся группы и состав портфеля.</p>
			</div>
		{/each}
	</div>
</section>
