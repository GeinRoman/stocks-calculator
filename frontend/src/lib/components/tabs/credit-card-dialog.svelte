<script lang="ts">
	import { untrack } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import type { CreditCard, CreditCardInput } from '#lib/logic/tabs/credit-cards.js';
	import PersonalAssetDateInput from './personal-asset-date-input.svelte';

	let { card, saving, onsave, onclose, blocked = false }: {
		card: CreditCard | null; saving: boolean; blocked?: boolean;
		onsave: (input: CreditCardInput) => Promise<void>; onclose: () => void;
	} = $props();
	const id = $props.id();
	const initial = untrack(() => card);
	let name = $state(initial?.name ?? '');
	let bankName = $state(initial?.bankName ?? '');
	let paymentDate = $state(initial?.paymentDate ?? '');
	let dateValid = $state(false);
	let amount = $state(initial ? String(initial.paymentAmountRub) : '');
	let limit = $state(initial ? String(initial.overallLimitRub) : '');
	let error = $state('');

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (saving) return;
		const input: CreditCardInput = { name: name.trim(), bankName: bankName.trim(), paymentDate, paymentAmountRub: Number(amount.replace(',', '.')), overallLimitRub: Number(limit.replace(',', '.')) };
		error = !input.name || !input.bankName ? 'Укажите название и банк.' : !dateValid ? 'Укажите корректную дату платежа.' : !amount.trim() || !limit.trim() || !Number.isFinite(input.paymentAmountRub) || !Number.isFinite(input.overallLimitRub) || input.paymentAmountRub < 0 || input.overallLimitRub < 0 ? 'Укажите неотрицательные суммы платежа и лимита.' : '';
		if (error) return;
		try { await onsave(input); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось сохранить запись. Попробуйте ещё раз.'; }
	}
</script>

<Dialog.Root open={true} onOpenChange={(open) => { if (!open && !saving) onclose(); }}>
	<Dialog.Content showCloseButton={!saving} class="max-h-[90vh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>{card ? 'Изменить карту' : 'Добавить карту'}</Dialog.Title>
			<Dialog.Description>Укажите данные кредитной карты и обязательного платежа.</Dialog.Description>
		</Dialog.Header>
		<form onsubmit={submit} class="space-y-4">
			<fieldset disabled={saving} class="space-y-4">
				<div class="space-y-2"><Label for={`${id}-name`}>Название</Label><Input id={`${id}-name`} bind:value={name} required /></div>
				<div class="space-y-2"><Label for={`${id}-bank`}>Банк</Label><Input id={`${id}-bank`} bind:value={bankName} required /></div>
				<PersonalAssetDateInput id={`${id}-date`} label="Дата обязательного платежа" bind:value={paymentDate} bind:valid={dateValid} required disabled={saving} />
				<div class="space-y-2"><Label for={`${id}-amount`}>Обязательный платёж, ₽</Label><Input id={`${id}-amount`} type="text" inputmode="decimal" bind:value={amount} required /></div>
				<div class="space-y-2"><Label for={`${id}-limit`}>Общий кредитный лимит, ₽</Label><Input id={`${id}-limit`} type="text" inputmode="decimal" bind:value={limit} required /></div>
			</fieldset>
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button type="button" variant="outline" disabled={saving} onclick={onclose}>Отмена</Button><Button type="submit" disabled={saving || blocked}>{saving ? 'Сохранение…' : 'Сохранить'}</Button></Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
