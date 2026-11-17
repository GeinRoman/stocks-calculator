<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	let { mode, busy, onauthenticate }: { mode: 'login' | 'register'; busy: boolean; onauthenticate: (mode: 'login' | 'register', user: string, pass: string) => Promise<void> } = $props();
	const id = $props.id();
	let user = $state(''), pass = $state(''), error = $state(''), submitting = $state(false);
	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (busy || submitting) return;
		error = !user || !pass ? 'Введите имя пользователя и пароль.' : new TextEncoder().encode(pass).length > 64 ? 'Пароль не должен превышать 64 байта UTF-8.' : '';
		if (error) return;
		submitting = true;
		try { await onauthenticate(mode, user, pass); pass = ''; }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось выполнить вход. Попробуйте ещё раз.'; }
		finally { submitting = false; }
	}
</script>

<section aria-labelledby={`${id}-heading`} class="w-full min-w-0 rounded-xl border border-border bg-card p-5 text-card-foreground sm:p-6">
	<h2 id={`${id}-heading`} class="text-center text-lg font-semibold">{mode === 'login' ? 'Войти' : 'Создать пользователя'}</h2>
	<form onsubmit={submit} class="mt-6 space-y-4">
		<div class="space-y-2"><Label for={`${id}-user`}>Имя пользователя</Label><Input id={`${id}-user`} bind:value={user} autocomplete="username" required disabled={busy || submitting} /></div>
		<div class="space-y-2"><Label for={`${id}-pass`}>Пароль</Label><Input id={`${id}-pass`} type="password" bind:value={pass} autocomplete={mode === 'login' ? 'current-password' : 'new-password'} required disabled={busy || submitting} /></div>
		{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
		<div class="flex justify-end"><Button type="submit" disabled={busy || submitting}>{submitting ? 'Подождите…' : mode === 'login' ? 'Войти' : 'Создать'}</Button></div>
	</form>
</section>
