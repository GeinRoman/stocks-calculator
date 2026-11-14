<script lang="ts">
	import HistoryIcon from '@lucide/svelte/icons/history';
	import LandmarkIcon from '@lucide/svelte/icons/landmark';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import XIcon from '@lucide/svelte/icons/x';
	import * as Drawer from '#lib/components/ui/drawer/index.js';
	import { Button, buttonVariants } from '#lib/components/ui/button/index.js';
	import { demoHistoryPage, historyTime, type HistoryKind, type HistoryTransaction } from '#lib/logic/tabs/transaction-history.js';
	let { kind, profileId, stacked = false }: { kind: HistoryKind; profileId: string; stacked?: boolean } = $props();
	let open = $state(false), items = $state<HistoryTransaction[]>([]), hasMore = $state(false), endReached = $state(false), error = $state('');
	const labels: Record<HistoryKind, string> = { stock: 'История операций с акциями', bond: 'История операций с облигациями', crypto: 'История операций с криптовалютой', deposit: 'История операций по вкладам', insurance: 'История операций по страхованию' };
	const title = $derived(labels[kind]);
	const date = new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'UTC' });
	const number = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 20 });
	const money = (amount: number, currency: string) => `${number.format(amount)} ${currency}`;
	function load(first = false) {
		if (first) { items = []; hasMore = false; endReached = false; }
		error = '';
		try {
			const last = items.at(-1);
			const page = demoHistoryPage(kind, profileId, !first && last ? { before: historyTime(last), beforeId: last.id } : undefined);
			items = first ? page.items : [...items, ...page.items];
			hasMore = page.items.length > 0;
			endReached = !first && page.items.length === 0;
		} catch {
			error = 'Не удалось загрузить операции. Попробуйте ещё раз.';
			hasMore = true;
		}
	}
	function changeOpen(value: boolean) { open = value; if (value) load(true); }
</script>

<Drawer.Root direction="right" {open} onOpenChange={changeOpen} shouldScaleBackground={false}>
	<Drawer.Trigger aria-label={title} title={title} class={`${buttonVariants({ size: 'icon-lg' })} fixed right-4 ${stacked ? 'bottom-20' : 'bottom-4'} z-30 shadow-lg`}>
		{#if kind === 'deposit'}<LandmarkIcon aria-hidden="true" />{:else if kind === 'insurance'}<ShieldIcon aria-hidden="true" />{:else}<HistoryIcon aria-hidden="true" />{/if}
	</Drawer.Trigger>
	<Drawer.Content class="h-[100dvh] gap-0 overflow-hidden bg-popover p-0 before:inset-0 before:rounded-none data-[vaul-drawer-direction=right]:w-[min(100vw,28rem)] sm:data-[vaul-drawer-direction=right]:max-w-none">
		<Drawer.Header class="shrink-0 border-b border-border pr-14">
			<Drawer.Title>{title}</Drawer.Title>
			<!-- TEMP -->
			<Drawer.Description>Демонстрационные данные, не отражающие реальные операции. Время указано в UTC.</Drawer.Description>
		</Drawer.Header>
		<Drawer.Close aria-label="Закрыть историю" title="Закрыть" class={`${buttonVariants({ variant: 'ghost', size: 'icon' })} absolute right-3 top-3`}><XIcon aria-hidden="true" /></Drawer.Close>
		<!-- svelte-ignore a11y_no_noninteractive_tabindex: The transaction list must be keyboard-scrollable. -->
		<div role="region" tabindex="0" aria-label="Список операций" class="min-h-0 flex-1 overflow-y-auto p-4 focus-visible:outline-2 focus-visible:outline-ring">
			<ul class="space-y-3">
				{#each items as item (item.id)}
					<li class="min-w-0 rounded-lg border border-border bg-card p-4 text-card-foreground">
						<div class="flex flex-wrap items-baseline justify-between gap-2"><strong class="break-words text-sm">{item.kind === 'stock' || item.kind === 'bond' ? item.instrument_code : item.kind === 'crypto' ? item.asset_name : item.name}</strong><span class="text-xs text-muted-foreground">{item.buying ? (item.kind === 'deposit' || item.kind === 'insurance' ? 'Внесение' : 'Покупка') : (item.kind === 'deposit' || item.kind === 'insurance' ? 'Вывод' : 'Продажа')}</span></div>
						<time datetime={historyTime(item)} class="mt-2 block text-xs text-muted-foreground">{date.format(new Date(historyTime(item)))} UTC</time>
						<dl class="mt-3 grid grid-cols-2 gap-3 text-sm [&>div]:min-w-0 [&_dd]:break-words">
							{#if item.kind === 'stock' || item.kind === 'bond'}
								<div><dt class="text-xs text-muted-foreground">Лотов</dt><dd>{number.format(item.lot_amount)}</dd></div><div><dt class="text-xs text-muted-foreground">Размер лота</dt><dd>{number.format(item.lot_size)}</dd></div><div><dt class="text-xs text-muted-foreground">Цена</dt><dd>{money(item.price, item.currency)}</dd></div>
							{:else if item.kind === 'crypto'}
								<div><dt class="text-xs text-muted-foreground">Платформа</dt><dd>{item.platform}</dd></div><div><dt class="text-xs text-muted-foreground">Количество</dt><dd>{number.format(item.amount)}</dd></div><div><dt class="text-xs text-muted-foreground">Цена</dt><dd>{money(item.price, item.currency)}</dd></div>
							{:else}
								<div><dt class="text-xs text-muted-foreground">{item.kind === 'deposit' ? 'Банк' : 'Компания'}</dt><dd>{item.kind === 'deposit' ? item.bank_name : item.company_name}</dd></div><div><dt class="text-xs text-muted-foreground">Сумма</dt><dd>{money(item.amount_rub, '₽')}</dd></div>
							{/if}
						</dl>
					</li>
				{:else}{#if !error}<li class="rounded-lg border border-dashed border-border p-5 text-sm text-muted-foreground">Операций пока нет.</li>{/if}{/each}
			</ul>
			{#if error}<p role="alert" class="pt-4 text-center text-sm text-destructive">{error}</p>{/if}
			{#if hasMore}<div class="pt-4 text-center"><Button variant="outline" onclick={() => load(items.length === 0)}>{error ? 'Повторить' : 'Загрузить ещё'}</Button></div>{:else if endReached}<p role="status" class="pt-4 text-center text-xs text-muted-foreground">Это все операции.</p>{/if}
		</div>
	</Drawer.Content>
</Drawer.Root>
