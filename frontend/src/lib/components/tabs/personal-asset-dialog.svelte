<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import PersonalAssetDateInput from './personal-asset-date-input.svelte';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { untrack } from 'svelte';
	import { validatePersonalAsset, type PersonalAsset, type PersonalAssetInput, type PersonalAssetType } from '#lib/logic/tabs/personal-assets.js';

	let { type, asset, saving, onsave, onclose }: {
		type: PersonalAssetType; asset: PersonalAsset | null; saving: boolean;
		onsave: (input: PersonalAssetInput) => Promise<void>; onclose: () => void;
	} = $props();
	const id = $props.id();
	const initial = untrack(() => asset);
	let name = $state(initial?.name ?? '');
	let provider = $state(initial?.provider ?? '');
	let amount = $state(initial ? String(initial.amountRub) : '');
	let interest = $state(initial ? String(initial.annualInterest) : '');
	let startDate = $state(initial?.startDate ?? '');
	let endDate = $state(initial?.endDate ?? '');
	let startValid = $state(false);
	let endValid = $state(false);
	let error = $state('');
	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (saving) return;
		const input: PersonalAssetInput = { name: name.trim(), provider: provider.trim(), amountRub: Number(amount.replace(',', '.')), annualInterest: Number(interest.replace(',', '.')), startDate, endDate };
		error = !amount.trim() || !interest.trim() ? 'Укажите сумму и процентную ставку.' : !startValid || !endValid ? 'Укажите корректные даты начала и окончания.' : endDate < startDate ? 'Дата окончания не может быть раньше даты начала.' : validatePersonalAsset(input) ?? '';
		if (error) return;
		try { await onsave(input); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось сохранить запись. Попробуйте ещё раз.'; }
	}
</script>

<Dialog.Root open={true} onOpenChange={(open) => { if (!open && !saving) onclose(); }}>
	<Dialog.Content showCloseButton={!saving} class="max-h-[90vh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>{asset ? 'Изменить' : 'Добавить'} {type === 'deposit' ? 'вклад' : 'страхование'}</Dialog.Title>
			<Dialog.Description>Укажите условия договора. Брокерские поручения не создаются.</Dialog.Description>
		</Dialog.Header>
		<form onsubmit={submit} class="space-y-4">
			<fieldset disabled={saving} class="space-y-4">
				<div class="space-y-2"><Label for={`${id}-name`}>Название</Label><Input id={`${id}-name`} bind:value={name} required /></div>
				<div class="space-y-2"><Label for={`${id}-provider`}>{type === 'deposit' ? 'Банк' : 'Страховая компания'}</Label><Input id={`${id}-provider`} bind:value={provider} required /></div>
				<div class="space-y-2"><Label for={`${id}-amount`}>Первоначальная сумма, ₽</Label><Input id={`${id}-amount`} type="text" inputmode="decimal" bind:value={amount} required /></div>
				<div class="space-y-2"><Label for={`${id}-interest`}>Годовая ставка, %</Label><Input id={`${id}-interest`} type="text" inputmode="decimal" bind:value={interest} required /></div>
				<div class="grid gap-4 sm:grid-cols-2">
					<PersonalAssetDateInput id={`${id}-start`} label="Дата начала" bind:value={startDate} bind:valid={startValid} required disabled={saving} />
					<PersonalAssetDateInput id={`${id}-end`} label="Дата окончания" bind:value={endDate} bind:valid={endValid} required disabled={saving} />
				</div>
			</fieldset>
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button type="button" variant="outline" disabled={saving} onclick={onclose}>Отмена</Button><Button type="submit" disabled={saving}>{saving ? 'Сохранение…' : 'Сохранить'}</Button></Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
