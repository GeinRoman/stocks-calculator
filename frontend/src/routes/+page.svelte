<script lang="ts">
	import { onMount } from 'svelte';
	import LoginPage from '#lib/components/login-page.svelte';
	import PortfolioApp from '#lib/components/portfolio-app.svelte';
	import { DEMO_REFRESH_KEY, demoAuthenticate, demoLogout, demoRefresh } from '#lib/logic/demo-auth.js';
	let authState = $state<'checking' | 'login' | 'authenticated'>('checking');
	let busy = $state(false), startupError = $state('');
	onMount(() => {
		let active = true;
		async function refresh() {
			try { const valid = await demoRefresh(window.localStorage); if (active) authState = valid ? 'authenticated' : 'login'; }
			catch { if (active) { startupError = 'Хранилище браузера недоступно. Включите его для демонстрационного входа.'; authState = 'login'; } }
		}
		void refresh();
		function onStorage(event: StorageEvent) {
			if (event.storageArea === window.localStorage && (event.key === DEMO_REFRESH_KEY || event.key === null)) {
				void refresh();
			}
		}
		window.addEventListener('storage', onStorage);
		return () => { active = false; window.removeEventListener('storage', onStorage); };
	});
	async function authenticate(user: string, pass: string) {
		if (busy) return;
		busy = true;
		try { await demoAuthenticate(user, pass, window.localStorage); startupError = ''; authState = 'authenticated'; }
		finally { busy = false; }
	}
	function logout(): string | null {
		try { demoLogout(window.localStorage); authState = 'login'; return null; }
		catch { return 'Не удалось удалить демонстрационную сессию. Проверьте настройки браузера и повторите попытку.'; }
	}
</script>

{#if authState === 'checking'}
	<main class="flex min-h-svh items-center justify-center bg-background p-4"><p role="status" class="text-sm text-muted-foreground">Проверка демонстрационной сессии…</p></main>
{:else if authState === 'login'}
	<LoginPage {busy} error={startupError} onauthenticate={authenticate} />
{:else}
	<PortfolioApp onlogout={logout} />
{/if}
