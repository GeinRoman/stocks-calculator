<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import CheckIcon from '@lucide/svelte/icons/check';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import type { Profile } from '#lib/api/contracts.js';

	let { profiles, busy, onselect, oncreate, ondelete, downloadSnapshot }: {
		profiles: Profile[]; busy: boolean;
		onselect: (profile: Profile) => Promise<void>;
		oncreate: (name: string) => Promise<void>;
		ondelete: (profile: Profile) => Promise<void>;
		downloadSnapshot: () => Promise<void>;
	} = $props();
	let open = $state(false), mode = $state<'list' | 'create' | 'switch' | 'delete'>('list');
	let selected = $state<Profile | null>(null), name = $state(''), error = $state(''), notice = $state('');
	let downloading = $state(false), downloadError = $state('');
	async function download() {
		if (downloading) return;
		downloading = true; downloadError = '';
		try { await downloadSnapshot(); }
		catch { downloadError = 'Не удалось скачать резервную копию. Попробуйте ещё раз.'; }
		finally { downloading = false; }
	}
	const current = $derived(profiles.find(profile => profile.default));
	const title = $derived({ list: 'Выбор профиля', create: 'Новый профиль', switch: 'Сменить профиль?', delete: 'Удалить профиль?' }[mode]);
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

<section aria-labelledby="live-profile-heading" class="mx-auto max-w-6xl">
	<!-- svelte-ignore a11y_no_noninteractive_tabindex (Scrollable header must be keyboard accessible.) -->
	<header tabindex="0" class="sticky top-0 z-20 mb-8 max-h-[50svh] overflow-y-auto border-b border-border bg-background py-3 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><h1 id="live-profile-heading" class="text-xl font-semibold tracking-tight sm:text-2xl">Профиль</h1><p class="mt-2 text-sm text-muted-foreground">Учётная запись и профили портфеля</p></header>
	<div class="min-w-0 rounded-xl border border-border bg-card p-4 text-card-foreground sm:p-8">
		<h2 class="mb-8 text-base font-medium">Данные пользователя</h2>
		<div class="flex flex-wrap items-start justify-between gap-6"><div class="min-w-0"><p class="text-sm text-muted-foreground">Текущий профиль</p><p class="mt-2 break-words text-lg font-semibold">{current?.name ?? 'Не выбран'}</p></div><div class="flex min-w-0 max-w-full flex-wrap gap-3"><Button variant="outline" class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled>Сменить имя профиля</Button><Button variant="destructive" class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={busy || !current} onclick={() => show('delete', current ?? null)}>Удалить профиль</Button></div></div>
		<div class="mt-8"><Button class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={busy} onclick={() => show('list')}>Сменить профиль</Button></div>
		<p class="mt-6 text-xs text-muted-foreground">Переименование профилей пока не поддерживается сервером.</p>
		<p role="status" class="mt-3 text-sm text-primary">{notice}</p>
	</div>
	<section aria-labelledby="snapshot-heading" class="mt-6 rounded-xl border border-border bg-card p-4 sm:p-8"><h2 id="snapshot-heading" class="text-base font-medium">Резервная копия</h2><p class="mt-2 text-sm text-muted-foreground">Скачать данные всех профилей и историю операций в формате JSON.</p><Button class="mt-4" variant="outline" disabled={downloading} onclick={download}>{downloading ? 'Загрузка…' : 'Скачать резервную копию'}</Button>{#if downloadError}<p role="alert" class="mt-3 text-sm text-destructive">{downloadError}</p>{/if}</section>
</section>

<Dialog.Root bind:open>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto border border-border sm:max-w-2xl">
		<Dialog.Header><Dialog.Title>{title}</Dialog.Title><Dialog.Description>
			{#if mode === 'list'}Выберите профиль портфеля или создайте новый.
			{:else if mode === 'switch'}После смены профиля обновите страницу на других активных устройствах и во вкладках для корректной работы приложения.
			{:else if mode === 'delete'}Профиль «{selected?.name}» и все его данные будут удалены. Это действие нельзя отменить.{selected?.default ? ' После удаления потребуется выбрать другой профиль по умолчанию. Обновите страницу на других активных устройствах и во вкладках.' : ''}
			{:else}Укажите название профиля.{!current ? ' Он станет профилем по умолчанию; обновите страницу на других активных устройствах и во вкладках.' : ' Текущий профиль останется выбранным.'}
			{/if}
		</Dialog.Description></Dialog.Header>
		{#if mode === 'list'}
			<div class="grid gap-3 sm:grid-cols-2">
				{#each profiles as profile (profile.name)}
					<Button variant="outline" aria-pressed={profile.default} disabled={busy} class="h-auto min-h-24 min-w-0 justify-between gap-3 whitespace-normal rounded-xl border-2 p-4 text-left aria-pressed:border-primary aria-pressed:bg-primary/10 aria-pressed:text-primary" onclick={() => { if (profile.default) open = false; else show('switch', profile); }}><span class="min-w-0 break-words">{profile.name}</span>{#if profile.default}<CheckIcon class="size-4" aria-hidden="true" />{/if}</Button>
				{/each}
				<Button variant="outline" disabled={busy} class="h-auto min-h-24 whitespace-normal rounded-xl border-2 p-4" onclick={() => { name = ''; show('create'); }}><PlusIcon aria-hidden="true" />Добавить профиль</Button>
			</div>
			<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => open = false}>Закрыть</Button></Dialog.Footer>
		{:else if mode === 'create'}
			<form onsubmit={create} class="space-y-4"><div class="space-y-2"><Label for="live-profile-name">Название профиля</Label><Input id="live-profile-name" bind:value={name} required disabled={busy} aria-invalid={!!error} /></div>{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => show('list')}>Отмена</Button><Button type="submit" disabled={busy}>Создать</Button></Dialog.Footer></form>
		{:else}
			{#if mode === 'switch'}<p class="break-words text-sm">Выбрать профиль «{selected?.name}»?</p>{/if}
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => show('list')}>Отмена</Button><Button variant={mode === 'delete' ? 'destructive' : 'default'} disabled={busy} onclick={confirm}>{mode === 'delete' ? 'Удалить' : 'Сменить профиль'}</Button></Dialog.Footer>
		{/if}
	</Dialog.Content>
</Dialog.Root>
