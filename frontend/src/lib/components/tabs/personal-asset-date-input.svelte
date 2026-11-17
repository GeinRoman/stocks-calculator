<script lang="ts">
	import CalendarDays from '@lucide/svelte/icons/calendar-days';
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import { untrack } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import { dateISO, formatDateISO, monthDays, parseDateText } from '#lib/logic/tabs/personal-asset-date.js';

	let { id, label, value = $bindable(''), valid = $bindable(false), required = false, disabled = false }: {
		id: string; label: string; value?: string; valid?: boolean; required?: boolean; disabled?: boolean;
	} = $props();

	let text = $state(untrack(() => formatDateISO(value)));
	let lastValue = untrack(() => value);
	let input = $state<HTMLInputElement | null>(null);
	let open = $state(false);
	let year = $state(1);
	let month = $state(1);
	let todayISO = $state('');
	const months = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь', 'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'];
	const weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
	let error = $derived(text !== '' && parseDateText(text) === null);

	$effect(() => {
		if (value !== lastValue) {
			text = formatDateISO(value);
			lastValue = value;
		}
	});
	$effect(() => {
		valid = !error && (text !== '' || !required) && (value === '' || formatDateISO(value) !== '');
		input?.setCustomValidity(error || (value !== '' && !formatDateISO(value)) ? 'Введите корректную дату в формате дд/мм/гггг' : '');
	});

	function updateText(next: string) {
		text = next;
		lastValue = parseDateText(next) ?? '';
		value = lastValue;
		valid = next === '' ? !required : lastValue !== '';
		input?.setCustomValidity(next !== '' && !lastValue ? 'Введите корректную дату в формате дд/мм/гггг' : '');
	}

	function openCalendar() {
		const selected = parseDateText(text);
		const today = new Date();
		todayISO = dateISO(today.getFullYear(), today.getMonth() + 1, today.getDate());
		year = selected ? Number(selected.slice(0, 4)) : today.getFullYear();
		month = selected ? Number(selected.slice(5, 7)) : today.getMonth() + 1;
		open = true;
	}

	function moveMonth(delta: number) {
		month += delta;
		if (month === 0) { month = 12; year--; }
		if (month === 13) { month = 1; year++; }
	}
</script>

<div class="grid gap-2">
	<Label for={id}>{label}</Label>
	<div class="flex gap-2">
		<Input bind:ref={input} {id} type="text" inputmode="numeric" placeholder="dd/mm/yyyy" {required} {disabled}
			value={text} aria-invalid={error ? 'true' : undefined} aria-describedby={error ? `${id}-error` : undefined}
			oninput={(event) => updateText(event.currentTarget.value)} />
		<Button type="button" variant="outline" size="icon" {disabled} aria-label={`Выбрать дату: ${label}`} onclick={openCalendar}>
			<CalendarDays />
		</Button>
	</div>
	{#if error}<p id={`${id}-error`} class="text-sm text-destructive">Введите корректную дату в формате дд/мм/гггг</p>{/if}
</div>

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Выбрать дату: {label}</Dialog.Title>
			<Dialog.Description>Выбранная дата: {formatDateISO(value) || 'не выбрана'}</Dialog.Description>
		</Dialog.Header>
		<div class="flex items-center justify-between gap-2">
			<Button type="button" variant="outline" size="icon" aria-label="Предыдущий месяц" disabled={year === 1 && month === 1} onclick={() => moveMonth(-1)}><ChevronLeft /></Button>
			<div aria-live="polite" class="font-medium">{months[month - 1]} {year}</div>
			<Button type="button" variant="outline" size="icon" aria-label="Следующий месяц" disabled={year === 9999 && month === 12} onclick={() => moveMonth(1)}><ChevronRight /></Button>
		</div>
		<div class="grid grid-cols-7 gap-1 text-center" aria-label="Календарь">
			{#each weekdays as weekday}<span class="text-sm text-muted-foreground">{weekday}</span>{/each}
			{#each monthDays(year, month) as day}
				{#if day}
					{@const date = dateISO(year, month, day)}
					<Button type="button" variant={value === date ? 'default' : date === todayISO ? 'outline' : 'ghost'} class="w-full px-0" aria-label={`${formatDateISO(date)}${date === todayISO ? ', сегодня' : ''}`} aria-pressed={value === date} onclick={() => { updateText(formatDateISO(date)); open = false; }}>{day}</Button>
				{:else}<span></span>{/if}
			{/each}
		</div>
	</Dialog.Content>
</Dialog.Root>
