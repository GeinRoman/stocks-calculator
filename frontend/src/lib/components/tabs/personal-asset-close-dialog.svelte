<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { untrack } from 'svelte';
	import type { PersonalAsset } from '#lib/logic/tabs/personal-assets.js';
	let { asset, saving, onsave, onclose }: { asset: PersonalAsset; saving: boolean; onsave: (amount: number) => Promise<void>; onclose: () => void } = $props();
	const id = $props.id();
	let amount = $state(untrack(() => String(asset.amountRub)));
	let error = $state('');
	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (saving) return;
		const value = Number(amount.replace(',', '.'));
		if (!amount.trim() || !Number.isFinite(value) || value <= 0) { error = 'Укажите положительную сумму закрытия.'; return; }
		error = '';
		try { await onsave(value); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось закрыть запись. Попробуйте ещё раз.'; }
	}
</script>

<Dialog.Root open={true} onOpenChange={(open) => { if (!open && !saving) onclose(); }}>
	<Dialog.Content showCloseButton={!saving}>
		<Dialog.Header><Dialog.Title>Закрыть «{asset.name}»</Dialog.Title><Dialog.Description>Укажите фактически полученную сумму. Будет записан вывод средств, а не удаление приобретения. Брокерские поручения не создаются.</Dialog.Description></Dialog.Header>
		<form onsubmit={submit} class="space-y-4">
			<div class="space-y-2"><Label for={`${id}-amount`}>Сумма закрытия, ₽</Label><Input id={`${id}-amount`} type="text" inputmode="decimal" bind:value={amount} required disabled={saving} /></div>
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button type="button" variant="outline" disabled={saving} onclick={onclose}>Отмена</Button><Button type="submit" disabled={saving}>{saving ? 'Сохранение…' : 'Записать вывод'}</Button></Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
