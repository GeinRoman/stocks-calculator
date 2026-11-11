<script lang="ts">
	import { onDestroy, onMount, tick } from 'svelte';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import type { Money } from '#lib/logic/tabs/currency.js';
	import type { CouponSchedule } from '#lib/logic/tabs/coupons.js';
	import { getBondCoupons } from '#lib/logic/tabs/coupon-service.js';
	import { convertMoney, formatMoney } from '#lib/logic/tabs/currency-service.js';
	import { couponDate, localCalendarDate, summarizeCoupons } from '#lib/logic/tabs/coupon-display.js';

	let { bondCode, bondName, quantity, onclose }: { bondCode: string; bondName: string; quantity: number; onclose: () => void } = $props();
	let open = $state(true);
	let schedule = $state<CouponSchedule | null>(null);
	let loading = $state(true);
	let error = $state(false);
	let rubles = $state(false);
	let today = $state('');
	let list = $state<HTMLElement>();
	let boundary = $state<HTMLElement>();
	let request = 0;
	const summary = $derived(summarizeCoupons(schedule?.coupons ?? [], today, quantity));
	const dateFormatter = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC' });

	function displayMoney(money: Money | null): string {
		const value = rubles ? convertMoney(money, 'RUB') : money;
		return value ? formatMoney(value) : 'Нет данных';
	}
	function displayDate(date: string): string {
		const parsed = new Date(`${date}T00:00:00Z`);
		return Number.isNaN(parsed.getTime()) ? date : dateFormatter.format(parsed);
	}
	function close() {
		request++;
		onclose();
	}
	async function load() {
		const token = ++request;
		loading = true;
		error = false;
		schedule = null;
		try {
			const result = await getBondCoupons(bondCode);
			if (token !== request) return;
			schedule = result;
			loading = false;
			await tick();
			if (token === request && list && boundary) {
				list.scrollTop += boundary.getBoundingClientRect().top - list.getBoundingClientRect().top - list.clientHeight / 2;
			}
		} catch {
			if (token === request) { error = true; loading = false; }
		}
	}
	onMount(() => {
		function updateToday() { today = localCalendarDate(new Date()); }
		updateToday();
		const timer = window.setInterval(updateToday, 60_000);
		window.addEventListener('focus', updateToday);
		void load();
		return () => { window.clearInterval(timer); window.removeEventListener('focus', updateToday); };
	});
	onDestroy(() => { request++; });
</script>

<Dialog.Root bind:open onOpenChange={(value) => { if (!value) close(); }}>
	<Dialog.Content showCloseButton={false} class="max-h-[90svh] gap-4 overflow-y-auto border border-border sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>Купоны облигации</Dialog.Title>
			<Dialog.Description class="break-words">{bondName} · {bondCode}</Dialog.Description>
		</Dialog.Header>
		<!-- TEMP -->
		<p class="text-xs text-muted-foreground">Демонстрационный график выплат. Данные не являются прогнозом.</p>
		<div class="flex flex-wrap items-center justify-between gap-2">
			<span class="text-sm text-muted-foreground">Валюта отображения</span>
			<div role="group" aria-label="Валюта отображения" class="flex gap-1">
				<Button size="sm" variant={rubles ? 'outline' : 'secondary'} aria-pressed={!rubles} onclick={() => rubles = false}>Исходная</Button>
				<Button size="sm" variant={rubles ? 'secondary' : 'outline'} aria-pressed={rubles} onclick={() => rubles = true}>RUB</Button>
			</div>
		</div>
		<div aria-live="polite" aria-busy={loading}>
			{#if loading}
				<p role="status" class="py-6 text-sm text-muted-foreground">Загрузка графика купонов…</p>
			{:else if error}
				<div class="space-y-3 py-4"><p role="alert" class="text-sm text-destructive">Не удалось загрузить график купонов.</p><Button variant="outline" onclick={load}>Попробовать снова</Button></div>
			{:else if schedule}
				<div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
					<div class="rounded-xl bg-muted p-3"><p class="text-xs text-muted-foreground">Всего выплат</p><p class="mt-1 font-medium">{summary.sorted.length}</p></div>
					<div class="rounded-xl bg-muted p-3"><p class="text-xs text-muted-foreground">Предстоящих</p><p class="mt-1 font-medium">{summary.remaining}</p></div>
					<div class="rounded-xl bg-muted p-3"><p class="text-xs text-muted-foreground">На облигацию</p><p class="mt-1 break-words font-medium">{displayMoney(summary.perBond)}</p></div>
					<div class="rounded-xl bg-muted p-3"><p class="text-xs text-muted-foreground">На позицию ({quantity} шт.)</p><p class="mt-1 break-words font-medium">{displayMoney(summary.holding)}</p></div>
				</div>
				<p class="mt-3 text-xs text-muted-foreground">Сумма предстоящих выплат, включая сегодня.</p>
				{#if !summary.sorted.length}
					<p class="py-6 text-sm text-muted-foreground">График купонов пуст.</p>
				{:else}
				<!-- svelte-ignore a11y_no_noninteractive_tabindex: the scroll region must be keyboard accessible. -->
				<div bind:this={list} role="region" aria-label="График купонных выплат" tabindex="0" class="mt-3 max-h-64 overflow-y-auto rounded-xl border border-border focus-visible:outline-2 focus-visible:outline-ring">
					<ul class="divide-y divide-border">
						{#if summary.remaining < summary.sorted.length}<li class="bg-muted px-4 py-2 text-xs font-medium text-muted-foreground">Прошедшие выплаты</li>{/if}
						{#each summary.sorted as coupon, index}
							{#if index === summary.sorted.length - summary.remaining}<li bind:this={boundary} class="bg-muted px-4 py-2 text-xs font-medium text-muted-foreground">Предстоящие выплаты</li>{/if}
							<li class="flex flex-wrap items-center justify-between gap-2 px-4 py-3"><time datetime={couponDate(coupon)}>{displayDate(couponDate(coupon))}</time><span class="font-medium">{displayMoney(coupon.amount)}</span></li>
						{/each}
						{#if summary.remaining === 0}<li bind:this={boundary} class="bg-muted px-4 py-2 text-xs text-muted-foreground">Предстоящих выплат нет</li>{/if}
					</ul>
				</div>
				{/if}
			{/if}
		</div>
		<Dialog.Footer><Button variant="outline" onclick={close}>Закрыть</Button></Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
