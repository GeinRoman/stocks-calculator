<script lang="ts">
	import { PieChart } from 'layerchart';
	import * as Chart from '#lib/components/ui/chart/index.js';
	import { Checkbox } from '#lib/components/ui/checkbox/index.js';
	import type { PortfolioSnapshot } from '#lib/logic/tabs/portfolio.js';

	let { values }: { values: PortfolioSnapshot['assetValues'] } = $props();

	const chartConfig = {
		stocks: { label: 'Акции', color: 'var(--chart-1)' },
		bonds: { label: 'Облигации', color: 'var(--chart-2)' },
		deposits: { label: 'Вклады', color: 'var(--chart-3)' }
	} satisfies Chart.ChartConfig;

	let excluded = $state<string[]>([]);
	const assets = $derived((Object.keys(chartConfig) as (keyof typeof chartConfig)[]).map((key) => ({
		key, ...chartConfig[key], value: values[key], selected: !excluded.includes(key)
	})));
	const selectedAssets = $derived(assets.filter((asset) => asset.selected));
	const selectedTotal = $derived(selectedAssets.reduce((total, asset) => total + asset.value, 0));
	const currency = new Intl.NumberFormat('ru-RU', {
		style: 'currency',
		currency: 'RUB',
		maximumFractionDigits: 0
	});
	const percent = new Intl.NumberFormat('ru-RU', {
		style: 'percent',
		maximumFractionDigits: 1
	});
</script>

<section aria-labelledby="general-heading" class="mx-auto max-w-6xl">
	<header class="mb-8 border-b border-border pb-6">
		<h1 id="general-heading" class="text-xl font-semibold tracking-tight sm:text-2xl">Общая информация</h1>
		<p class="mt-2 text-sm text-muted-foreground">Состав вашего портфеля</p>
	</header>

	<div class="rounded-xl border border-border bg-card p-4 text-card-foreground sm:p-8">
		<div class="mb-8 flex flex-wrap items-center justify-between gap-3">
			<h2 class="text-base font-medium">Распределение активов</h2>
			<!-- TEMP -->
			<span class="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">Демонстрационные данные</span>
		</div>

		<div class="grid gap-8 lg:grid-cols-2 lg:items-stretch lg:gap-10">
			<div class="min-w-0">
				<figure aria-label="Распределение выбранных активов">
					{#if selectedTotal > 0}
						<Chart.Container config={chartConfig} class="mx-auto aspect-square w-full max-w-80" aria-hidden="true">
							<PieChart
								data={selectedAssets}
								key="key"
								value="value"
								c="color"
								innerRadius={0}
								props={{ arc: { stroke: 'var(--card)', strokeWidth: 2 } }}
							>
								{#snippet tooltip({ context })}
									<Chart.Tooltip hideLabel role="tooltip">
										{#snippet formatter({ value })}
											<span>{context.tooltip.data.label}</span>
											<span class="ml-auto font-medium tabular-nums">{percent.format(Number(value) / selectedTotal)}</span>
										{/snippet}
									</Chart.Tooltip>
								{/snippet}
							</PieChart>
						</Chart.Container>
					{:else}
						<div class="mx-auto flex aspect-square w-full max-w-80 items-center justify-center rounded-full border border-dashed border-border bg-muted/30 p-4 text-center">
							<p class="text-sm text-muted-foreground">Выберите активы для отображения</p>
						</div>
					{/if}
					<figcaption class="sr-only">
						{#each selectedAssets as asset (asset.key)}
							{asset.label}: {percent.format(asset.value / selectedTotal)}.
						{/each}
					</figcaption>
				</figure>

				<div aria-live="polite" aria-atomic="true" class="mt-8 text-center">
					<p class="text-sm text-muted-foreground">Общая стоимость выбранных активов</p>
					<p class="mt-2 text-2xl font-semibold tracking-tight tabular-nums sm:text-3xl">{currency.format(selectedTotal)}</p>
				</div>
			</div>

			<fieldset class="min-w-0 border-t border-border pt-6 lg:border-t-0 lg:border-l lg:pt-0 lg:pl-8">
				<legend class="sr-only">Активы для отображения на диаграмме</legend>
				<h3 class="mb-2 text-base font-medium">Активы</h3>
				<p class="mb-6 text-xs leading-relaxed text-muted-foreground">Выберите, что включить в диаграмму</p>
				<div class="space-y-2">
					{#each assets as asset (asset.key)}
						<div class="flex flex-wrap items-center gap-x-3 gap-y-2 rounded-md py-3">
							<Checkbox id={`asset-${asset.key}`} checked={asset.selected} onCheckedChange={(checked) => {
								excluded = checked ? excluded.filter((key) => key !== asset.key) : [...excluded, asset.key];
							}} />
							<label for={`asset-${asset.key}`} class="flex min-w-0 flex-1 cursor-pointer items-center gap-2 text-sm">
								<span class="size-2.5 shrink-0 rounded-sm" style:background-color={asset.color} aria-hidden="true"></span>
								<span class:opacity-50={!asset.selected}>{asset.label}</span>
							</label>
							<span class="ml-auto text-sm tabular-nums" class:text-muted-foreground={!asset.selected}>
								{asset.selected && selectedTotal > 0 ? percent.format(asset.value / selectedTotal) : '—'}
							</span>
						</div>
					{/each}
				</div>
			</fieldset>
		</div>
	</div>
</section>
