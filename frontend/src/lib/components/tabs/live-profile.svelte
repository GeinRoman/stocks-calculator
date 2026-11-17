<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import CheckIcon from '@lucide/svelte/icons/check';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import type { Profile } from '#lib/api/contracts.js';

	let { profiles, busy, onselect, oncreate, ondelete, onrename, downloadSnapshot, installSnapshot }: {
		profiles: Profile[]; busy: boolean;
		onselect: (profile: Profile) => Promise<void>;
		oncreate: (name: string) => Promise<void>;
		ondelete: (profile: Profile) => Promise<void>;
		onrename: (profile: Profile, name: string) => Promise<void>;
		downloadSnapshot: () => Promise<void>;
		installSnapshot: (snapshot: unknown) => Promise<boolean>;
	} = $props();
	let open = $state(false), mode = $state<'list' | 'create' | 'rename' | 'switch' | 'delete'>('list');
	let selected = $state<Profile | null>(null), name = $state(''), error = $state(''), notice = $state('');
	let downloading = $state(false), downloadError = $state('');
	let snapshot = $state<unknown>(null), snapshotName = $state(''), snapshotError = $state(''), snapshotNotice = $state(''), reading = $state(false);
	async function selectSnapshot(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		snapshotError = ''; snapshotNotice = '';
		if (file.size > 32 * 1024 * 1024) { snapshotError = 'Размер файла не должен превышать 32 МиБ.'; return; }
		reading = true;
		try {
			const value: unknown = JSON.parse(await file.text());
			if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error();
			snapshot = value; snapshotName = file.name;
		} catch { snapshotError = 'Не удалось прочитать JSON-файл резервной копии.'; }
		finally { reading = false; }
	}
	async function restoreSnapshot() {
		if (!snapshot || busy) return;
		snapshotError = '';
		try {
			const refreshed = await installSnapshot(snapshot);
			snapshot = null; snapshotName = '';
			snapshotNotice = refreshed ? 'Резервная копия применена.' : 'Копия применена, но данные не удалось обновить. Повторите загрузку на странице общей информации.';
		} catch { snapshotError = 'Не удалось подтвердить применение копии. Проверьте файл и соединение; если связь прервалась, обновите страницу перед повторной попыткой.'; }
	}
	async function download() {
		if (downloading) return;
		downloading = true; downloadError = '';
		try { await downloadSnapshot(); }
		catch { downloadError = 'Не удалось скачать резервную копию. Попробуйте ещё раз.'; }
		finally { downloading = false; }
	}
	const current = $derived(profiles.find(profile => profile.default));
	const title = $derived({ list: 'Выбор профиля', create: 'Новый профиль', rename: 'Переименовать профиль', switch: 'Сменить профиль?', delete: 'Удалить профиль?' }[mode]);
	function show(next: typeof mode, profile: Profile | null = null) { mode = next; selected = profile; error = ''; open = true; }
	async function create(event: SubmitEvent) {
		event.preventDefault();
		if (busy) return;
		const trimmed = name.trim();
		if (!trimmed) { error = 'Введите название профиля.'; return; }
		try { await oncreate(trimmed); name = ''; notice = 'Профиль создан.'; open = false; }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось создать профиль.'; }
	}
	async function rename(event: SubmitEvent) {
		event.preventDefault();
		if (busy || !selected) return;
		const trimmed = name.trim();
		if (!trimmed) { error = 'Введите название профиля.'; return; }
		if (trimmed === selected.name) { error = 'Введите новое название профиля.'; return; }
		try { await onrename(selected, trimmed); notice = 'Профиль переименован.'; open = false; }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось переименовать профиль.'; }
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
		<div class="flex flex-wrap items-start justify-between gap-6"><div class="min-w-0"><p class="text-sm text-muted-foreground">Текущий профиль</p><p class="mt-2 break-words text-lg font-semibold">{current?.name ?? 'Не выбран'}</p></div><div class="flex min-w-0 max-w-full flex-wrap gap-3"><Button variant="outline" class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={busy || !current} onclick={() => { name = current?.name ?? ''; show('rename', current ?? null); }}>Сменить имя профиля</Button><Button variant="destructive" class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={busy || !current} onclick={() => show('delete', current ?? null)}>Удалить профиль</Button></div></div>
		<div class="mt-8"><Button class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={busy} onclick={() => show('list')}>Сменить профиль</Button></div>
		<p role="status" class="mt-3 text-sm text-primary">{notice}</p>
	</div>
	<section aria-labelledby="snapshot-heading" class="mt-6 rounded-xl border border-border bg-card p-4 sm:p-8"><h2 id="snapshot-heading" class="text-base font-medium">Резервная копия</h2><p class="mt-2 text-sm text-muted-foreground">Скачать данные всех профилей и историю операций в формате JSON.</p><Button class="mt-4" variant="outline" disabled={downloading || busy} onclick={download}>{downloading ? 'Загрузка…' : 'Скачать резервную копию'}</Button>{#if downloadError}<p role="alert" class="mt-3 text-sm text-destructive">{downloadError}</p>{/if}
		<div class="mt-6 space-y-2"><Label for="snapshot-file">Применить резервную копию из JSON-файла</Label><input id="snapshot-file" type="file" accept=".json,application/json" disabled={busy || reading || snapshot !== null} onchange={selectSnapshot} class="block h-9 w-full min-w-0 max-w-full rounded-3xl border border-input bg-input/50 px-3 py-1 text-sm file:mr-2 file:border-0 file:bg-transparent file:text-foreground focus-visible:outline-2 focus-visible:outline-ring disabled:opacity-50" /></div>
		{#if snapshotError && !snapshot}<p role="alert" class="mt-3 text-sm text-destructive">{snapshotError}</p>{/if}
		{#if snapshotNotice}<p role="status" class="mt-3 text-sm text-primary">{snapshotNotice}</p>{/if}
	</section>
</section>

<Dialog.Root bind:open>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto border border-border sm:max-w-2xl">
		<Dialog.Header><Dialog.Title>{title}</Dialog.Title><Dialog.Description>
			{#if mode === 'list'}Выберите профиль портфеля или создайте новый.
			{:else if mode === 'switch'}После смены профиля обновите страницу на других активных устройствах и во вкладках для корректной работы приложения.
			{:else if mode === 'delete'}Профиль «{selected?.name}» и все его данные будут удалены. Это действие нельзя отменить.{selected?.default ? ' После удаления потребуется выбрать другой профиль по умолчанию. Обновите страницу на других активных устройствах и во вкладках.' : ''}
			{:else if mode === 'rename'}Укажите новое название профиля.
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
		{:else if mode === 'create' || mode === 'rename'}
			<form onsubmit={mode === 'create' ? create : rename} class="space-y-4"><div class="space-y-2"><Label for="live-profile-name">Название профиля</Label><Input id="live-profile-name" bind:value={name} required disabled={busy} aria-invalid={!!error} /></div>{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => mode === 'rename' ? open = false : show('list')}>Отмена</Button><Button type="submit" disabled={busy}>{mode === 'rename' ? 'Сохранить' : 'Создать'}</Button></Dialog.Footer></form>
		{:else}
			{#if mode === 'switch'}<p class="break-words text-sm">Выбрать профиль «{selected?.name}»?</p>{/if}
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => show('list')}>Отмена</Button><Button variant={mode === 'delete' ? 'destructive' : 'default'} disabled={busy} onclick={confirm}>{mode === 'delete' ? 'Удалить' : 'Сменить профиль'}</Button></Dialog.Footer>
		{/if}
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root open={snapshot !== null} onOpenChange={(value) => { if (!value && !busy) { snapshot = null; snapshotError = ''; } }}>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto border border-border sm:max-w-2xl">
		<Dialog.Header><Dialog.Title>Применить резервную копию?</Dialog.Title><Dialog.Description>Все текущие профили, их данные и история операций будут заменены данными из файла. Это действие нельзя отменить. Обновите страницу на других активных устройствах и во вкладках.</Dialog.Description></Dialog.Header>
		<p class="break-words text-sm">{snapshotName}</p>
		{#if snapshotError}<p role="alert" class="text-sm text-destructive">{snapshotError}</p>{/if}
		<Dialog.Footer><Button variant="outline" disabled={busy} onclick={() => { snapshot = null; snapshotError = ''; }}>Отмена</Button><Button variant="destructive" disabled={busy} onclick={restoreSnapshot}>{busy ? 'Применение…' : 'Применить копию'}</Button></Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
