<script lang="ts">
	import PlusIcon from '@lucide/svelte/icons/plus';
	import CheckIcon from '@lucide/svelte/icons/check';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import type { Profile } from '#lib/logic/tabs/portfolio.js';

	let {
		profiles,
		selectedId,
		username,
		onselect,
		oncreate,
		onrename,
		ondelete
	}: {
		profiles: Profile[];
		selectedId: string;
		username: string;
		onselect: (id: string) => void;
		oncreate: (name: string) => void;
		onrename: (id: string, name: string) => void;
		ondelete: (id: string) => void;
	} = $props();

	const selectedProfile = $derived(profiles.find((profile) => profile.id === selectedId));
	let open = $state(false);
	let mode = $state<'switch' | 'create' | 'rename' | 'delete'>('switch');
	let name = $state('');
	let error = $state('');
	let notice = $state('');
	const title = $derived({ switch: 'Выбор профиля', create: 'Новый профиль', rename: 'Сменить имя профиля', delete: 'Удалить профиль?' }[mode]);

	function show(nextMode: typeof mode) {
		mode = nextMode;
		name = nextMode === 'rename' ? selectedProfile?.name ?? '' : '';
		error = '';
		open = true;
	}

	function save(event: SubmitEvent) {
		event.preventDefault();
		const trimmed = name.trim();
		if (!trimmed) {
			error = 'Введите название профиля.';
			return;
		}
		if (profiles.some((profile) => profile.name.toLocaleLowerCase('ru-RU') === trimmed.toLocaleLowerCase('ru-RU') && (mode === 'create' || profile.id !== selectedId))) {
			error = 'Профиль с таким названием уже существует.';
			return;
		}
		if (mode === 'create') oncreate(trimmed);
		else onrename(selectedId, trimmed);
		notice = mode === 'create' ? 'Профиль создан.' : 'Название профиля изменено.';
		open = false;
	}
</script>

<section aria-labelledby="profile-heading" class="mx-auto max-w-6xl">
	<header class="mb-8 border-b border-border pb-6">
		<h1 id="profile-heading" class="text-xl font-semibold tracking-tight sm:text-2xl">Профиль</h1>
		<p class="mt-2 text-sm text-muted-foreground">Учётная запись и профили портфеля</p>
	</header>

	<div class="rounded-xl border border-border bg-card p-4 text-card-foreground sm:p-8">
		<div class="mb-8 flex flex-wrap items-center justify-between gap-3">
			<h2 class="text-base font-medium">Данные пользователя</h2>
			<span class="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">Демонстрационные данные</span>
		</div>
		<dl class="space-y-8">
			<div><dt class="text-sm text-muted-foreground">Имя пользователя</dt><dd class="mt-2 break-words text-base font-medium">{username}</dd></div>
			<div class="flex flex-wrap items-start justify-between gap-6 border-t border-border pt-6">
				<div class="min-w-0"><dt class="text-sm text-muted-foreground">Текущий профиль</dt><dd class="mt-2 break-words text-lg font-semibold">{selectedProfile?.name}</dd></div>
				<div class="flex min-w-0 max-w-full flex-wrap gap-3">
					<Button variant="outline" class="h-auto min-h-9 max-w-full whitespace-normal py-2" onclick={() => show('rename')}>Сменить имя профиля</Button>
					<Button variant="destructive" class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={profiles.length <= 1} onclick={() => show('delete')}>Удалить профиль</Button>
				</div>
			</div>
		</dl>
		<div class="mt-8"><Button class="h-auto min-h-9 max-w-full whitespace-normal py-2" onclick={() => show('switch')}>Сменить профиль</Button></div>
		<p class="mt-6 text-xs leading-relaxed text-muted-foreground">Изменения сохраняются только до перезагрузки страницы. Последний профиль удалить нельзя.</p>
		<p role="status" class="mt-3 text-sm text-primary">{notice}</p>
	</div>
</section>

<Dialog.Root bind:open>
	<Dialog.Content showCloseButton={false} class="max-h-[85svh] overflow-y-auto border border-border sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>{title}</Dialog.Title>
			<Dialog.Description>
				{#if mode === 'switch'}Выберите профиль портфеля или создайте новый.
				{:else if mode === 'delete'}Профиль «{selectedProfile?.name}» будет удалён из демонстрационного списка. Это действие нельзя отменить.
				{:else}Укажите название профиля. Изменения применяются только к демонстрационным данным.
				{/if}
			</Dialog.Description>
		</Dialog.Header>

		{#if mode === 'switch'}
			<div class="grid gap-3 sm:grid-cols-2">
				{#each profiles as profile (profile.id)}
					<Button
						variant="outline"
						aria-pressed={profile.id === selectedId}
						class="h-auto min-h-24 min-w-0 justify-between gap-3 whitespace-normal rounded-xl border-2 p-4 text-left aria-pressed:border-primary aria-pressed:bg-primary/10 aria-pressed:text-primary"
						onclick={() => { onselect(profile.id); notice = `Выбран профиль «${profile.name}».`; open = false; }}
					>
						<span class="min-w-0 break-words">{profile.name}</span>
						{#if profile.id === selectedId}<CheckIcon class="size-4" aria-hidden="true" />{/if}
					</Button>
				{/each}
				<Button variant="outline" class="h-auto min-h-24 whitespace-normal rounded-xl border-2 p-4" onclick={() => show('create')}><PlusIcon aria-hidden="true" />Добавить профиль</Button>
			</div>
			<Dialog.Footer><Button variant="outline" onclick={() => open = false}>Закрыть</Button></Dialog.Footer>
		{:else if mode === 'delete'}
			<Dialog.Footer>
				<Button variant="outline" onclick={() => open = false}>Отмена</Button>
				<Button variant="destructive" onclick={() => { ondelete(selectedId); notice = 'Профиль удалён.'; open = false; }}>Удалить</Button>
			</Dialog.Footer>
		{:else}
			<form onsubmit={save} class="min-w-0 space-y-6">
				<div class="space-y-2">
					<Label for="profile-name">Название профиля</Label>
					<Input id="profile-name" bind:value={name} maxlength={80} required placeholder="Например, Основной портфель" aria-invalid={!!error} aria-describedby={error ? 'profile-name-error' : undefined} />
					{#if error}<p id="profile-name-error" role="alert" class="text-sm text-destructive">{error}</p>{/if}
				</div>
				<Dialog.Footer>
					<Button variant="outline" onclick={() => open = false}>Отмена</Button>
					<Button type="submit">{mode === 'create' ? 'Создать' : 'Сохранить'}</Button>
				</Dialog.Footer>
			</form>
		{/if}
	</Dialog.Content>
</Dialog.Root>
