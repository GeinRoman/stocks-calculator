<script lang="ts">
	import { untrack } from 'svelte';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { Slider } from '#lib/components/ui/slider/index.js';
	import { validateWeights } from '#lib/logic/tabs/stock-editor.js';
	import type { GroupWeight, PortfolioSnapshot } from '#lib/logic/tabs/portfolio.js';

	let { portfolio, saving, onsave, onclose }: {
		portfolio: PortfolioSnapshot;
		saving: boolean;
		onsave: (weights: GroupWeight[]) => Promise<void>;
		onclose: () => void;
	} = $props();
	let open = $state(true);
	let weights = $state<GroupWeight[]>(untrack(() => portfolio.groups.map(({ id, targetPercent }) => ({ id, value: targetPercent ?? null }))));
	let error = $state('');
	const total = $derived(weights.reduce((sum, { value }) => sum + (value ?? 0), 0));
	function update(id: string, value: number | null, input?: HTMLInputElement) {
		const updated = weights.map((weight) => weight.id === id ? { ...weight, value } : weight);
		const invalid = validateWeights(updated);
		if (invalid) {
			error = invalid;
			if (input) input.value = weights.find((weight) => weight.id === id)?.value?.toString() ?? '';
			return;
		}
		weights = updated; error = '';
	}
	async function save(event: SubmitEvent) {
		event.preventDefault(); error = '';
		const invalid = validateWeights(weights, true);
		if (invalid) { error = invalid; return; }
		try { await onsave(weights); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось сохранить доли. Попробуйте ещё раз.'; }
	}
</script>

<Dialog.Root bind:open onOpenChange={(value) => { if (!value) onclose(); }}>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto border border-border sm:max-w-2xl" onInteractOutside={(event) => { if (saving) event.preventDefault(); }} onEscapeKeydown={(event) => { if (saving) event.preventDefault(); }}>
		<Dialog.Header><Dialog.Title>Настроить доли</Dialog.Title><Dialog.Description>Задайте целые проценты для всех групп. Для сохранения сумма должна равняться 100%. Пустое поле означает, что доля не задана; 0% — явно заданное значение.</Dialog.Description></Dialog.Header>
		<form onsubmit={save} class="min-w-0 space-y-6">
			<fieldset disabled={saving} class="min-w-0 space-y-5">
				<legend class="sr-only">Целевые доли групп</legend>
				{#each weights as weight (weight.id)}
					{@const group = portfolio.groups.find(({ id }) => id === weight.id)!}
					{@const remaining = 100 - total + (weight.value ?? 0)}
					<div class="space-y-4 rounded-lg border border-border bg-muted/30 p-4">
						<div class="flex flex-wrap items-center justify-between gap-3">
							<Label for={`weight-${weight.id}`} class="min-w-0 break-words">{group.name}</Label>
							<div class="flex items-center gap-2"><Input id={`weight-${weight.id}`} type="number" min={0} max={remaining} step={1} value={weight.value ?? ''} placeholder="Не задана" class="w-32 [&::-webkit-inner-spin-button]:opacity-100" oninput={(event) => update(weight.id, event.currentTarget.value === '' ? null : event.currentTarget.valueAsNumber, event.currentTarget)} /><span class="text-sm">%</span></div>
						</div>
						<!-- Restrict selectable steps without changing the scale or moving other thumbs. -->
						<Slider type="single" value={weight.value ?? 0} min={0} max={100} step={Array.from({ length: remaining + 1 }, (_, value) => value)} disabled={saving || remaining === 0} aria-label={`Доля группы ${group.name}`} aria-valuetext={weight.value === null ? 'Не задана' : `${weight.value}%`} onValueChange={(value) => update(weight.id, value)} class={weight.value === null ? 'opacity-50' : ''} />
						{#if weight.value === null}<p class="text-xs text-muted-foreground">Не задана. Введите значение или передвиньте ползунок.</p>{/if}
					</div>
				{/each}
			</fieldset>
			<div aria-live="polite" class="grid gap-3 rounded-lg border border-border bg-muted p-4 text-sm sm:grid-cols-2 sm:items-center">
				<div><p>Распределено: <strong>{total}%</strong></p><p class="mt-1 text-muted-foreground">Осталось: {100 - total}%</p></div>
				<p class="text-xs text-muted-foreground" class:invisible={total === 100 && weights.every(({ value }) => value !== null)}>Для сохранения задайте доли всех групп с общей суммой 100%.</p>
			</div>
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button variant="outline" disabled={saving} onclick={onclose}>Отмена</Button><Button type="submit" disabled={saving || !!validateWeights(weights, true)}>{saving ? 'Сохранение…' : 'Сохранить доли'}</Button></Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
