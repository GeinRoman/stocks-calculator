<script lang="ts">
	import { onMount } from 'svelte';
	import LoginPage from '#lib/components/login-page.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { ApiClient, HttpError, REFRESH_TOKEN_KEY } from '#lib/api/client.js';

	let authState = $state<'checking' | 'login' | 'authenticated'>('checking');
	let busy = $state(false), startupError = $state('');
	let client: ApiClient | undefined;

	onMount(() => {
		let active = true;
		try {
			const storage = window.localStorage;
			client = new ApiClient({ baseUrl: '/api', storage, onSessionExpired: () => {
				if (active) { startupError = 'Сессия завершена. Войдите снова.'; authState = 'login'; }
			} });
			if (!storage.getItem(REFRESH_TOKEN_KEY)) authState = 'login';
			else void client.refresh().then(() => { if (active) authState = 'authenticated'; }).catch(() => {
				if (active) { startupError = 'Не удалось восстановить сессию. Проверьте соединение и войдите снова.'; authState = 'login'; }
			});
		} catch {
			startupError = 'Не удалось получить доступ к хранилищу браузера.';
			authState = 'login';
		}
		return () => { active = false; };
	});

	async function authenticate(mode: 'login' | 'register', user: string, pass: string) {
		if (busy) return;
		if (!client) throw new Error('Хранилище браузера недоступно.');
		busy = true;
		try {
			await client.authenticate(mode === 'login' ? '/login' : '/createuser', user, pass);
			startupError = ''; authState = 'authenticated';
		} catch (error) {
			if (error instanceof HttpError && error.status === 401) throw new Error('Неверное имя пользователя или пароль.');
			if (error instanceof HttpError && error.status === 409) throw new Error('Пользователь с таким именем уже существует.');
			throw new Error('Не удалось связаться с сервером или сохранить сессию. Попробуйте ещё раз.');
		} finally { busy = false; }
	}

	function logout() {
		try { client?.logout(); startupError = ''; }
		catch { startupError = 'Не удалось удалить сессию из хранилища браузера.'; }
		authState = 'login';
	}
</script>

{#if authState === 'checking'}
	<main class="flex min-h-svh items-center justify-center bg-background p-4"><p role="status" class="text-sm text-muted-foreground">Проверка сессии…</p></main>
{:else if authState === 'login'}
	<LoginPage {busy} error={startupError} onauthenticate={authenticate} />
{:else}
	<!-- TEMP: Live portfolio loading is the next integration step; never show demo data as a real account. -->
	<main class="flex min-h-svh items-center justify-center bg-background p-4"><div class="max-w-md space-y-4 text-center"><h1 class="text-xl font-semibold">Вход выполнен</h1><p class="text-sm text-muted-foreground">Подключение данных портфеля ещё не завершено.</p><Button variant="outline" onclick={logout}>Выйти</Button></div></main>
{/if}
