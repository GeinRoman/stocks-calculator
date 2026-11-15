<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import type { Profile } from '#lib/api/contracts.js';

	let { profiles, busy, onselect, oncreate, ondelete }: {
		profiles: Profile[]; busy: boolean;
		onselect: (profile: Profile) => Promise<void>;
		oncreate: (name: string) => Promise<void>;
		ondelete: (profile: Profile) => Promise<void>;
	} = $props();
	let open = $state(false), mode = $state<'list' | 'create' | 'switch' | 'delete'>('list');
	let selected = $state<Profile | null>(null), name = $state(''), error = $state(''), notice = $state('');
	const current = $derived(profiles.find(profile => profile.default));
	const title = $derived({ list: 'Профили', create: 'Новый профиль', switch: 'Сменить профиль?', delete: 'Удалить профиль?' }[mode]);
	function show(next: typeof mode, profile: Profile | null = null) { mode = next; selected = profile; error = ''; open = true; }
	async function create(event: SubmitEvent) {
		event.preventDefault();
		if (busy) return;
		const trimmed = name.trim();
		if (!trimmed) { error = 'Введите название профиля.'; return; }
		try { await oncreate(trimmed); name = ''; notice = 'Профиль создан.'; open = false; }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось создать профиль.'; }
	}
	async function confirm() {
		if (!selected || busy) return;
		const profile = selected;
		try {
			if (mode === 'switch') { await onselect(profile); notice = `Профиль «${profile.name}» выбран.`; }
			else { await ondelete(profile); notice = `Профиль «${profile.name}» удалён.`; }
			open = false;
		} catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось выполнить операцию.'; }
	}
</script>

<section aria-labelledby="live-profile-heading" class="space-y-3 rounded-xl border border-border bg-card p-4">
	<div class="flex flex-wrap items-center justify-between gap-3"><div class="min-w-0"><h2 id="live-profile-heading" class="font-medium">Профиль</h2><p class="mt-1 break-words text-sm text-muted-foreground">{current ? `Текущий профиль: ${current.name}` : 'Профиль по умолчанию не выбран.'}</p></div><Button variant="outline" disabled={busy} onclick={() => show('list')}>Управление профилями</Button></div>
	<p class="text-xs text-muted-foreground">Переименование профилей пока не поддерживается сервером.</p>
	<p role="status" class="text-sm text-primary">{notice}</p>
</section>

<Dialog.Root bind:open>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto">
		<Dialog.Header><Dialog.Title>{title}</Dialog.Title><Dialog.Description>
			{#if mode === 'list'}Выберите профиль, создайте новый или удалите ненужный.
			{:else if mode === 'switch'}После смены профиля обновите страницу на других активных устройствах и во вкладках для корректной работы приложения.
			{:else if mode === 'delete'}Профиль «{selected?.name}» и все его данные будут удалены. Это действие нельзя отменить.{selected?.default ? ' После удаления потребуется выбрать другой профиль по умолчанию. Обновите страницу на других активных устройствах и во вкладках.' : ''}
			{:else}Укажите название профиля.{!current ? ' Он станет профилем по умолчанию; обновите страницу на других активных устройствах и во вкладках.' : ' Текущий профиль останется выбранным.'}
			{/if}
		</Dialog.Description></Dialog.Header>
		{#if mode === 'list'}
			<div class="space-y-2">
				{#each profiles as profile (profile.name)}
					<div class="flex flex-wrap items-center gap-2 rounded-lg border border-border p-3"><span class="min-w-0 flex-1 break-words text-sm">{profile.name}{profile.default ? ' (текущий)' : ''}</span><Button variant="outline" size="sm" disabled={busy || profile.default} onclick={() => show('switch', profile)}>Выбрать</Button><Button variant="destructive" size="sm" disabled={busy} aria-label={`Удалить профиль ${profile.name}`} onclick={() => show('delete', profile)}>Удалить</Button></div>
				{:else}<p class="text-sm text-muted-foreground">Профилей пока нет.</p>{/each}
			</div>
			<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => open = false}>Закрыть</Button><Button disabled={busy} onclick={() => { name = ''; show('create'); }}>Добавить профиль</Button></Dialog.Footer>
		{:else if mode === 'create'}
			<form onsubmit={create} class="space-y-4"><div class="space-y-2"><Label for="live-profile-name">Название профиля</Label><Input id="live-profile-name" bind:value={name} required disabled={busy} aria-invalid={!!error} /></div>{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => show('list')}>Отмена</Button><Button type="submit" disabled={busy}>Создать</Button></Dialog.Footer></form>
		{:else}
			{#if mode === 'switch'}<p class="break-words text-sm">Выбрать профиль «{selected?.name}»?</p>{/if}
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => show('list')}>Отмена</Button><Button variant={mode === 'delete' ? 'destructive' : 'default'} disabled={busy} onclick={confirm}>{mode === 'delete' ? 'Удалить' : 'Сменить профиль'}</Button></Dialog.Footer>
		{/if}
	</Dialog.Content>
</Dialog.Root>
