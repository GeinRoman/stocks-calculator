<script lang="ts" generics="T extends { id: string; name: string; targetPercent: number | null }">
	import type { Snippet } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { masonryGroups } from '#lib/logic/tabs/group-layout.js';

	let { groups = $bindable(), profileName, saving, instrumentPlural, validate, createGroup, instruments, prepare, onsave, oncancel, children }: {
		groups: T[];
		profileName: string;
		saving: boolean;
		instrumentPlural: string;
		validate: (groups: T[]) => string | null;
		createGroup: (name: string) => T;
		instruments: (group: T) => { lotCount?: number }[];
		prepare: (groups: T[], removedIds: string[]) => T[];
		onsave: (groups: T[]) => Promise<void>;
		oncancel: () => void;
		children: Snippet<[T, boolean]>;
	} = $props();
	const id = $props.id();
	let groupName = $state('');
	let error = $state('');
	let deletingId = $state<string | null>(null);
	let removed = $state<Record<string, (number | undefined)[]>>({});
	const deleting = $derived(groups.find((group) => group.id === deletingId));
	const validation = $derived(validate(prepare(groups, Object.keys(removed))));
	function removeGroup() {
		if (!deleting) return;
		const items = instruments(deleting);
		removed[deleting.id] = items.map(item => item.lotCount);
		for (const item of items) item.lotCount = 0;
		deletingId = null; error = '';
	}
	function restoreGroup(group: T) {
		const values = removed[group.id];
		if (!values) return;
		instruments(group).forEach((item, index) => item.lotCount = values[index]);
		delete removed[group.id]; error = '';
	}
	function addGroup(event: SubmitEvent) {
		event.preventDefault();
		const updated = [...groups, createGroup(groupName.trim())];
		const invalid = validate(prepare(updated, Object.keys(removed)));
		if (invalid) { error = invalid; return; }
		groups = updated; groupName = ''; error = '';
	}
	async function save() {
		error = '';
		try { await onsave(prepare(groups, Object.keys(removed))); }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось сохранить изменения. Попробуйте ещё раз.'; }
	}
</script>

<section aria-labelledby={`editor-heading-${id}`} class="mx-auto max-w-6xl">
	<!-- svelte-ignore a11y_no_noninteractive_tabindex: The bounded region must be keyboard-scrollable. -->
	<div role="region" tabindex="0" aria-label="Заголовок и действия редактирования состава" class="sticky top-0 z-20 mb-6 max-h-[50svh] overflow-y-auto border-b border-border bg-background px-1 py-3 focus-visible:outline-2 focus-visible:outline-ring">
		<h1 id={`editor-heading-${id}`} class="text-xl font-semibold sm:text-2xl">Редактирование состава</h1>
		<p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
		<p class="mt-3 text-xs leading-relaxed text-muted-foreground">Изменения применятся только после подтверждения. Доли настраиваются отдельно. Для смены профиля подтвердите изменения или отмените редактирование.</p>
		<p class="mt-2 text-xs text-muted-foreground">На узком экране таблицы можно прокручивать по горизонтали.</p>
		<div class="mt-4 flex flex-wrap gap-3">
			<Button disabled={saving || !!validation} class="h-auto min-h-9 max-w-full whitespace-normal py-2" onclick={save}>{saving ? 'Сохранение…' : 'Подтвердить изменения'}</Button>
			<Button variant="outline" disabled={saving} onclick={oncancel}>Отмена</Button>
		</div>
		{#if validation}<p role="alert" class="mt-3 text-sm text-destructive">{validation}</p>{/if}
		{#if error}<p role="alert" class="mt-3 text-sm text-destructive">{error}</p>{/if}
	</div>
	<fieldset disabled={saving} class="min-w-0 space-y-6">
		<legend class="sr-only">Группы и {instrumentPlural} портфеля</legend>
		<div use:masonryGroups class="group-layout grid min-w-0 items-start gap-6 xl:grid-cols-2">
			{#each groups as group (group.id)}
				<section aria-label={`Редактирование группы: ${group.name}`} class="min-w-0 overflow-hidden rounded-xl border border-border bg-card">
					<header class="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-primary/10 p-4">
						<h2 class="min-w-0 break-words text-base font-semibold">{group.name}</h2>
						{#if group.id in removed}<Button variant="outline" size="sm" aria-label={`Восстановить группу ${group.name}`} onclick={() => restoreGroup(group)}>Восстановить группу</Button>{:else}<Button variant="destructive" size="sm" aria-label={`Удалить группу ${group.name}`} onclick={() => deletingId = group.id}>Удалить группу</Button>{/if}
					</header>
					{#if group.id in removed}<p class="px-4 pt-3 text-xs text-muted-foreground">Группа будет удалена после подтверждения (черновик).</p>{/if}
					<fieldset disabled={group.id in removed} class="min-w-0"><legend class="sr-only">Состав группы {group.name}</legend>{@render children(group, group.id in removed)}</fieldset>
				</section>
			{:else}<p data-group-empty class="rounded-xl border border-dashed border-border p-6 text-sm text-muted-foreground xl:col-span-2">Групп пока нет. Создайте первую группу.</p>{/each}
		</div>
		<form onsubmit={addGroup} class="space-y-3 rounded-xl border border-border bg-card p-4">
			<Label for={`new-group-${id}`}>Новая группа</Label>
			<div class="flex flex-wrap gap-3"><Input id={`new-group-${id}`} bind:value={groupName} required placeholder="Название группы" class="min-w-0 flex-1 basis-48" /><Button type="submit" class="h-auto min-h-9 max-w-full whitespace-normal py-2">Добавить группу</Button></div>
			<p class="text-xs text-muted-foreground">Целевая доля новой группы будет не задана.</p>
		</form>
	</fieldset>
</section>

<AlertDialog.Root open={deletingId !== null} onOpenChange={(value) => { if (!value) deletingId = null; }}>
	<AlertDialog.Content class="border border-border">
		<AlertDialog.Header><AlertDialog.Title>Удалить группу?</AlertDialog.Title><AlertDialog.Description>Группа «{deleting?.name}» будет помечена для удаления, а её {instrumentPlural} — для продажи. Изменения вступят в силу только после подтверждения состава.</AlertDialog.Description></AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Отмена</AlertDialog.Cancel>
			<AlertDialog.Action variant="destructive" class="h-auto min-h-9 whitespace-normal py-2" onclick={removeGroup}>Удалить группу и {instrumentPlural}</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
