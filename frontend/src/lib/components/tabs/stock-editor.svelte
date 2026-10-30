<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as Table from '#lib/components/ui/table/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import AddStockDialog from './add-stock-dialog.svelte';
	import { addStockToGroup, validateDraft } from '#lib/logic/tabs/stock-editor.js';
	import type { DraftGroup } from '#lib/logic/tabs/portfolio.js';

	let { groups = $bindable(), profileName, saving, onsave, oncancel }: {
		groups: DraftGroup[];
		profileName: string;
		saving: boolean;
		onsave: (groups: DraftGroup[]) => Promise<void>;
		oncancel: () => void;
	} = $props();
	let groupName = $state('');
	let error = $state('');
	let addingTo = $state<string | null>(null);
	let deletingId = $state<string | null>(null);
	const deleting = $derived(groups.find(({ id }) => id === deletingId));
	const validation = $derived(validateDraft(groups));
	function addGroup(event: SubmitEvent) {
		event.preventDefault();
		const updated = [...groups, { id: crypto.randomUUID(), name: groupName.trim(), targetPercent: null, stocks: [] }];
		const invalid = validateDraft(updated);
		if (invalid) { error = invalid; return; }
		groups = updated; groupName = ''; error = '';
	}
	async function save() {
		error = '';
		try { await onsave(groups); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось сохранить изменения. Попробуйте ещё раз.'; }
	}
</script>

<section aria-labelledby="editor-heading" class="mx-auto max-w-6xl">
	<header class="mb-6 border-b border-border pb-6">
		<h1 id="editor-heading" class="text-xl font-semibold sm:text-2xl">Редактирование состава</h1>
		<p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
		<p class="mt-3 text-xs leading-relaxed text-muted-foreground">Изменения применятся только после подтверждения. Доли настраиваются отдельно. Для смены профиля подтвердите изменения или отмените редактирование.</p>
		<p class="mt-2 text-xs text-muted-foreground">На узком экране таблицы можно прокручивать по горизонтали.</p>
	</header>
	<fieldset disabled={saving} class="min-w-0 space-y-6">
		<legend class="sr-only">Группы и акции портфеля</legend>
		<div class="flex flex-wrap gap-3">
			<Button disabled={saving || !!validation} class="h-auto min-h-9 max-w-full whitespace-normal py-2" onclick={save}>{saving ? 'Сохранение…' : 'Подтвердить изменения'}</Button>
			<Button variant="outline" onclick={oncancel}>Отмена</Button>
		</div>
		{#if validation}<p role="alert" class="text-sm text-destructive">{validation}</p>{/if}
		{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
		{#each groups as group (group.id)}
			<section aria-label={`Редактирование группы: ${group.name}`} class="min-w-0 overflow-hidden rounded-xl border border-border bg-card">
				<header class="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-primary/10 p-4">
					<h2 class="min-w-0 break-words text-base font-semibold">{group.name}</h2>
					<Button variant="destructive" size="sm" aria-label={`Удалить группу ${group.name}`} onclick={() => deletingId = group.id}>Удалить группу</Button>
				</header>
				<Table.Root class="min-w-max">
					<Table.Caption class="sr-only">Акции группы {group.name}. Таблицу можно прокручивать по горизонтали.</Table.Caption>
					<Table.Header><Table.Row><Table.Head>Название</Table.Head><Table.Head>Код</Table.Head><Table.Head>Размер лота</Table.Head><Table.Head>Лотов</Table.Head><Table.Head><span class="sr-only">Действия</span></Table.Head></Table.Row></Table.Header>
					<Table.Body>
						{#each group.stocks as stock (stock.code)}
							<Table.Row class="odd:bg-card even:bg-muted">
								<Table.Cell class="max-w-64 whitespace-normal font-medium">{stock.name}</Table.Cell>
								<Table.Cell>{stock.code}</Table.Cell>
								<Table.Cell>{stock.lotSize}</Table.Cell>
								<Table.Cell><Input aria-label={`Лотов ${stock.code}`} type="number" min={1} max={Number.MAX_SAFE_INTEGER} step={1} bind:value={stock.lotCount} aria-invalid={!Number.isSafeInteger(stock.lotCount) || !stock.lotCount || stock.lotCount < 1} class="w-28 [&::-webkit-inner-spin-button]:opacity-100" /></Table.Cell>
								<Table.Cell><Button variant="destructive" size="sm" aria-label={`Удалить акцию ${stock.code}`} onclick={() => { group.stocks = group.stocks.filter(({ code }) => code !== stock.code); error = ''; }}>Удалить</Button></Table.Cell>
							</Table.Row>
						{:else}<Table.Row><Table.Cell colspan={5} class="text-muted-foreground">В группе пока нет акций.</Table.Cell></Table.Row>{/each}
					</Table.Body>
				</Table.Root>
				<div class="border-t border-border p-4"><Button variant="outline" size="sm" class="h-auto min-h-8 max-w-full whitespace-normal py-2" aria-label={`Добавить акцию в группу ${group.name}`} onclick={() => addingTo = group.id}>Добавить акцию</Button></div>
			</section>
		{:else}<p class="rounded-xl border border-dashed border-border p-6 text-sm text-muted-foreground">Групп пока нет. Создайте первую группу.</p>{/each}
		<form onsubmit={addGroup} class="space-y-3 rounded-xl border border-border bg-card p-4">
			<Label for="new-group-name">Новая группа</Label>
			<div class="flex flex-wrap gap-3"><Input id="new-group-name" bind:value={groupName} required placeholder="Название группы" class="min-w-0 flex-1 basis-48" /><Button type="submit" class="h-auto min-h-9 max-w-full whitespace-normal py-2">Добавить группу</Button></div>
			<p class="text-xs text-muted-foreground">Целевая доля новой группы будет не задана.</p>
		</form>
	</fieldset>
</section>

{#if addingTo}
	<AddStockDialog groupId={addingTo} {groups} onclose={() => addingTo = null} onadd={(stock, lots) => { groups = addStockToGroup(groups, addingTo!, stock, lots); error = ''; }} />
{/if}

<AlertDialog.Root open={deletingId !== null} onOpenChange={(value) => { if (!value) deletingId = null; }}>
	<AlertDialog.Content class="border border-border">
		<AlertDialog.Header><AlertDialog.Title>Удалить группу?</AlertDialog.Title><AlertDialog.Description>Группа «{deleting?.name}» и все её акции будут удалены из черновика. Изменения вступят в силу только после подтверждения состава.</AlertDialog.Description></AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Отмена</AlertDialog.Cancel>
			<AlertDialog.Action variant="destructive" class="h-auto min-h-9 whitespace-normal py-2" onclick={() => { groups = groups.filter(({ id }) => id !== deletingId); deletingId = null; error = ''; }}>Удалить группу и акции</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
