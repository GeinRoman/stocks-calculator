<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { daysUntilPersonalAssetEnd, getPersonalAssetStatus, localCalendarDate } from '#lib/logic/tabs/personal-assets.js';
	import { formatDateISO } from '#lib/logic/tabs/personal-asset-date.js';
	import { sortCreditCards, type CreditCard, type CreditCardInput, type CreditCardsSnapshot } from '#lib/logic/tabs/credit-cards.js';
	import { createCreditCard, updateCreditCard, deleteCreditCard } from '#lib/logic/tabs/credit-card-service.js';
	import CreditCardDialog from './credit-card-dialog.svelte';
	import type { CreditCardOperation } from '#lib/logic/live-credit-cards.js';

	let { snapshot, profileName, oncommit, onbusychange, readOnly = false, onmutate, blocked = false }: { snapshot: CreditCardsSnapshot; profileName: string; oncommit: (updated: CreditCardsSnapshot) => void; onbusychange: (busy: boolean) => void; readOnly?: boolean; onmutate?: (operation: CreditCardOperation) => Promise<void>; blocked?: boolean } = $props();
	let dialog = $state<{ kind: 'edit'; card: CreditCard | null } | { kind: 'delete'; card: CreditCard } | null>(null);
	let saving = $state(false);
	let error = $state('');
	let today = $state('');
	$effect(() => onbusychange(dialog !== null || saving));
	onDestroy(() => onbusychange(false));
	onMount(() => {
		const refresh = () => today = localCalendarDate(new Date());
		refresh();
		window.addEventListener('focus', refresh);
		const timer = window.setInterval(refresh, 60_000);
		return () => { window.removeEventListener('focus', refresh); window.clearInterval(timer); };
	});
	const money = (value: number) => new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value);
	const plurals = new Intl.PluralRules('ru-RU');
	const remainingDays = (days: number) => `${days} ${{ zero: 'дней', one: 'день', two: 'дня', few: 'дня', many: 'дней', other: 'дня' }[plurals.select(days)]}`;
	async function save(input: CreditCardInput) {
		if (!dialog || dialog.kind !== 'edit' || saving) return;
		saving = true;
		try {
			if (onmutate) await onmutate(dialog.card ? { kind: 'update', id: dialog.card.id, input } : { kind: 'create', input });
			else { const updated = dialog.card ? await updateCreditCard(snapshot, dialog.card.id, input) : await createCreditCard(snapshot, input); oncommit(updated); }
			dialog = null;
		} finally { saving = false; }
	}
	async function remove() {
		if (!dialog || dialog.kind !== 'delete' || saving) return;
		saving = true; error = '';
		try { if (onmutate) await onmutate({ kind: 'remove', id: dialog.card.id }); else { const updated = await deleteCreditCard(snapshot, dialog.card.id); oncommit(updated); } dialog = null; }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось удалить запись. Попробуйте ещё раз.'; }
		finally { saving = false; }
	}
</script>

<section class="mx-auto max-w-6xl space-y-6" aria-label="Кредитные карты">
	<!-- svelte-ignore a11y_no_noninteractive_tabindex (Scrollable header must be keyboard accessible.) -->
	<header tabindex="0" class="sticky top-0 z-20 max-h-[50svh] overflow-y-auto border-b border-border bg-background py-3 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><h1 class="text-xl font-semibold sm:text-2xl">Кредитные карты</h1><p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
		{#if !readOnly}<!-- TEMP --><p class="mt-2 text-xs text-muted-foreground">Демонстрационные данные: изменения сохраняются только в памяти и пропадают при перезагрузке.</p>{/if}
		<div class="mt-4 flex flex-wrap items-center gap-4"><Button size="sm" disabled={(readOnly && !onmutate) || blocked} onclick={() => dialog = { kind: 'edit', card: null }}>Добавить карту</Button>{#if blocked}<Button size="sm" variant="outline" onclick={() => window.location.reload()}>Обновить страницу</Button>{/if}</div>
		{#if blocked}<p role="alert" class="mt-2 text-sm text-destructive">Обновите страницу перед повторным редактированием.</p>{/if}
	</header>
	<section class="space-y-4" aria-label="Список кредитных карт">
		<div class="space-y-3">
			{#each sortCreditCards(snapshot.cards) as card (card.id)}
				{@const status = today ? getPersonalAssetStatus(card.paymentDate, today) : 'normal'}
				<article class="min-w-0 rounded-xl border border-border p-4 {status === 'ending' ? 'bg-destructive/10' : status === 'expired' ? 'bg-muted' : 'bg-card'}">
					<div class="flex flex-wrap items-start justify-between gap-4">
						<div class="min-w-0 space-y-2 break-words">
							<h2 class="font-semibold">{card.name}</h2>
							<p class="text-sm">Банк: {card.bankName}</p>
							<p class="text-sm">Обязательный платёж: {money(card.paymentAmountRub)} · Дата: {formatDateISO(card.paymentDate)}</p>
							<p class="text-sm">Общий кредитный лимит: {money(card.overallLimitRub)}</p>
							{#if status === 'expired'}<p class="text-sm font-medium text-muted-foreground">Срок платежа истёк</p>{:else if today}<p class="text-sm font-medium {status === 'ending' ? 'text-destructive' : 'text-muted-foreground'}">До платежа: {remainingDays(daysUntilPersonalAssetEnd(card.paymentDate, today))}</p>{/if}
						</div>
						<div class="flex flex-wrap gap-2"><Button size="sm" variant="outline" disabled={(readOnly && !onmutate) || blocked} onclick={() => dialog = { kind: 'edit', card }}>Изменить</Button><Button size="sm" variant="destructive" disabled={(readOnly && !onmutate) || blocked} onclick={() => { error = ''; dialog = { kind: 'delete', card }; }}>Удалить</Button></div>
					</div>
				</article>
			{:else}<p class="text-sm text-muted-foreground">Карт пока нет.</p>{/each}
		</div>
	</section>
	<footer class="border-t border-border pt-4 text-xs text-muted-foreground">
		<div aria-label="Обозначения сроков" class="flex flex-wrap gap-x-6 gap-y-2">
			<p class="flex items-center gap-2"><span aria-hidden="true" class="size-3 shrink-0 rounded-sm border border-destructive/30 bg-destructive/10"></span>Платёж сегодня или в ближайшие 14 дней</p>
			<p class="flex items-center gap-2"><span aria-hidden="true" class="size-3 shrink-0 rounded-sm border border-border bg-muted"></span>Срок платежа истёк</p>
		</div>
	</footer>
</section>

{#if dialog?.kind === 'edit'}
	<CreditCardDialog card={dialog.card} {saving} {blocked} onsave={save} onclose={() => dialog = null} />
{/if}
<AlertDialog.Root open={dialog?.kind === 'delete'} onOpenChange={(open) => { if (!open && !saving) dialog = null; }}>
	{#if dialog?.kind === 'delete'}
		<AlertDialog.Content>
			<AlertDialog.Header><AlertDialog.Title>Удалить запись?</AlertDialog.Title><AlertDialog.Description>Будет удалена только запись «{dialog.card.name}». Банковская карта не будет закрыта, а задолженность не будет погашена.</AlertDialog.Description></AlertDialog.Header>
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<AlertDialog.Footer><Button variant="outline" disabled={saving} onclick={() => dialog = null}>Отмена</Button><Button variant="destructive" disabled={saving || blocked} onclick={remove}>{saving ? 'Удаление…' : 'Удалить запись'}</Button></AlertDialog.Footer>
		</AlertDialog.Content>
	{/if}
</AlertDialog.Root>
