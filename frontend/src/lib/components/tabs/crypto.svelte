<script lang="ts">
	import ExternalLink from '@lucide/svelte/icons/external-link';
	import { Button } from '#lib/components/ui/button/index.js';
	import { masonryGroups } from '#lib/logic/tabs/group-layout.js';
	import { cryptoValue, type CryptoSnapshot } from '#lib/logic/tabs/crypto.js';
	import { saveCryptoWeights } from '#lib/logic/tabs/crypto-service.js';
	import type { GroupWeight } from '#lib/logic/tabs/portfolio.js';
	import GroupWeightsDialog from './group-weights-dialog.svelte';
	import CryptoEditor from './crypto-editor.svelte';
	import CryptoPricesDrawer from './crypto-prices-drawer.svelte';

	let { snapshot, profileName, oncommit, onbusychange }: { snapshot: CryptoSnapshot; profileName: string; oncommit: (snapshot: CryptoSnapshot) => void; onbusychange: (busy: boolean) => void } = $props();
	let editing = $state(false), weightsOpen = $state(false), pricesOpen = $state(false), notice = $state('');
	const total = $derived(cryptoValue(snapshot.groups.flatMap(g => g.assets)));
	const currency = $derived(snapshot.priceCurrency ?? 'USDT');
	$effect(() => onbusychange(editing || weightsOpen || pricesOpen));
	const number = (value: number) => value !== 0 && Math.abs(value) < 1e-20 ? value.toExponential() : new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 20 }).format(value);
	const money = (value: number | null) => value === null ? 'Нет данных' : `${number(value)} ${currency}`;
	async function saveWeights(weights: GroupWeight[]) {
		oncommit(saveCryptoWeights(snapshot, weights)); weightsOpen = false; notice = 'Доли сохранены в памяти.';
	}
</script>

{#if editing}
	<CryptoEditor {snapshot} {profileName} oncancel={() => editing = false} onsave={(updated) => { oncommit(updated); editing = false; notice = 'Изменения сохранены в памяти.'; }} />
{:else}
	<section aria-labelledby="crypto-heading" class="mx-auto max-w-6xl">
		<!-- svelte-ignore a11y_no_noninteractive_tabindex: The bounded header must be keyboard-scrollable. -->
		<header tabindex="0" aria-label="Заголовок и управление криптовалютой" class="sticky top-0 z-10 mb-6 max-h-[50svh] overflow-y-auto border-b border-border bg-background py-2 focus-visible:outline-2 focus-visible:outline-ring">
			<div class="mb-2 flex flex-wrap items-start justify-between gap-x-4 gap-y-1"><div class="min-w-0"><h1 id="crypto-heading" class="text-lg font-semibold tracking-tight sm:text-xl">Криптовалюта</h1><p class="break-words text-xs text-muted-foreground">{profileName}</p></div><div class="ml-auto min-w-0 max-w-full break-words text-right"><p class="text-xs text-muted-foreground">Общая стоимость криптовалюты</p><p class="text-lg font-semibold tracking-tight tabular-nums">{money(total.value)}</p>{#if total.partial && total.value !== null}<p class="text-xs text-muted-foreground">Частичная стоимость</p>{/if}<!-- TEMP --><span class="text-xs text-muted-foreground">Демонстрационные данные</span></div></div>
			<div class="flex flex-wrap items-center gap-2"><Button class="h-auto min-h-9 max-w-full whitespace-normal py-2" onclick={() => { notice = ''; editing = true; }}>Редактировать состав</Button><Button variant="outline" disabled={!snapshot.groups.length} onclick={() => weightsOpen = true}>Настроить доли</Button><Button class="ml-auto h-auto min-h-9 max-w-full whitespace-normal py-2" variant="outline" disabled={!snapshot.groups.some(g => g.assets.length)} onclick={() => pricesOpen = true}>Указать цены</Button></div>
			<p role="status" class="text-sm text-primary">{notice}</p>
		</header>
		<h2 class="mb-6 text-base font-medium">Группы</h2>
		<div use:masonryGroups class="group-layout grid min-w-0 items-start gap-6 xl:grid-cols-2">
			{#each snapshot.groups as group (group.id)}{@const value = cryptoValue(group.assets)}<section aria-label={`Группа: ${group.name}`} class="min-w-0 overflow-hidden rounded-xl border border-border bg-card text-card-foreground">
				<header class="border-b border-border bg-primary/10 p-4 sm:p-6"><div class="flex flex-wrap items-baseline justify-between gap-2"><h3 class="break-words text-lg font-semibold">{group.name}</h3><div class="text-right"><p class="font-medium tabular-nums"><span class="sr-only">Стоимость группы: </span>{money(value.value)}</p>{#if value.partial && value.value !== null}<p class="text-xs text-muted-foreground">Частичная стоимость</p>{/if}</div></div><dl class="mt-4 grid grid-cols-2 gap-3 [&>div]:min-w-0 [&_dd]:wrap-anywhere"><div><dt class="text-xs text-muted-foreground">Целевая доля</dt><dd class="text-sm">{group.weight}%</dd></div><div><dt class="text-xs text-muted-foreground">Текущая доля</dt><dd class="text-sm text-primary">{value.value === null || total.value === null || total.value === 0 || total.partial ? 'Нет данных' : `${number(value.value / total.value * 100)}%`}</dd></div></dl></header>
				<ul class="divide-y divide-border">{#each group.assets as asset (asset.id)}<li class="p-4 odd:bg-card even:bg-muted sm:p-6"><div class="mb-4 flex items-center gap-2"><h4 class="break-words text-sm font-semibold">{asset.name}</h4>{#if asset.url && /^https?:\/\//i.test(asset.url)}<a href={asset.url} target="_blank" rel="noopener noreferrer" aria-label={`Открыть ссылку ${asset.name} в новой вкладке`} title="Открыть ссылку" class="text-primary focus-visible:outline-2 focus-visible:outline-ring"><ExternalLink class="size-4" aria-hidden="true" /></a>{/if}</div><dl class="grid grid-cols-2 gap-x-4 gap-y-4 sm:grid-cols-3 [&>div]:min-w-0 [&_dd]:wrap-anywhere"><div><dt class="text-xs text-muted-foreground">Количество</dt><dd class="mt-1 text-sm tabular-nums">{number(asset.amount)}</dd></div><div><dt class="text-xs text-muted-foreground">Платформа</dt><dd class="mt-1 text-sm">{asset.platform}</dd></div><div><dt class="text-xs text-muted-foreground">Годовая ставка</dt><dd class="mt-1 text-sm">{number(asset.annualInterest)}%</dd></div><div><dt class="text-xs text-muted-foreground">Цена</dt><dd class="mt-1 text-sm">{money(asset.price)}</dd></div><div><dt class="text-xs text-muted-foreground">Общая стоимость</dt><dd class="mt-1 text-sm font-medium">{money(cryptoValue([asset]).value)}</dd></div></dl></li>{:else}<li class="p-6 text-sm text-muted-foreground">В этой группе пока нет активов.</li>{/each}</ul>
			</section>{:else}<div data-group-empty class="rounded-xl border border-dashed border-border p-6 text-center xl:col-span-2 sm:p-12"><p class="font-medium">В профиле пока нет криптовалюты</p><p class="mt-2 text-sm text-muted-foreground">Добавьте группу и активы в режиме редактирования.</p></div>{/each}
		</div>
		<!-- TEMP --><p class="mt-6 text-xs leading-relaxed text-muted-foreground">Цены вводятся вручную и не являются текущими котировками. Данные сохраняются только в памяти и пропадают при перезагрузке; реальные операции не выполняются.</p>
	</section>
{/if}

{#if weightsOpen}<GroupWeightsDialog groups={snapshot.groups.map(group => ({ id: group.id, name: group.name, targetPercent: group.weight }))} saving={false} onsave={saveWeights} onclose={() => weightsOpen = false} />{/if}
{#if pricesOpen}<CryptoPricesDrawer {snapshot} onclose={() => pricesOpen = false} oncommit={(updated) => { oncommit(updated); notice = 'Цены сохранены в памяти.'; }} />{/if}
