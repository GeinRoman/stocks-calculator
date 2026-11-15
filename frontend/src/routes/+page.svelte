<script lang="ts">
	import { onMount } from 'svelte';
	import LoginPage from '#lib/components/login-page.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { ApiClient, HttpError, REFRESH_TOKEN_KEY } from '#lib/api/client.js';
	import { portfolioApi } from '#lib/api/portfolio-api.js';
	import { loadPortfolioState, type PortfolioState } from '#lib/logic/portfolio-state.js';

	let authState = $state<'checking' | 'login' | 'authenticated'>('checking');
	let busy = $state(false), startupError = $state('');
	let portfolio = $state<PortfolioState | null>(null);
	let portfolioStatus = $state<'loading' | 'ready' | 'no-profile' | 'error'>('loading');
	let portfolioGeneration = 0;
	let client: ApiClient | undefined;
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
	const money = (value: number | null) => value === null ? 'Нет данных' : currency.format(value);

	async function loadPortfolio() {
		if (!client) return;
		const generation = ++portfolioGeneration;
		portfolioStatus = 'loading'; portfolio = null;
		try {
			const result = await loadPortfolioState(portfolioApi(client));
			if (generation !== portfolioGeneration) return;
			portfolio = result; portfolioStatus = 'ready';
		} catch (error) {
			if (generation !== portfolioGeneration) return;
			portfolioStatus = error instanceof HttpError && error.status === 404 && error.message === 'Default profile not found' ? 'no-profile' : 'error';
		}
	}
	function authenticated() { authState = 'authenticated'; void loadPortfolio(); }

	onMount(() => {
		let active = true;
		try {
			const storage = window.localStorage;
			client = new ApiClient({ baseUrl: '/api', storage, onSessionExpired: () => {
				if (active) { portfolioGeneration++; portfolio = null; startupError = 'Сессия завершена. Войдите снова.'; authState = 'login'; }
			} });
			if (!storage.getItem(REFRESH_TOKEN_KEY)) authState = 'login';
			else void client.refresh().then(() => { if (active) authenticated(); }).catch(() => {
				if (active) { startupError = 'Не удалось восстановить сессию. Проверьте соединение и войдите снова.'; authState = 'login'; }
			});
		} catch {
			startupError = 'Не удалось получить доступ к хранилищу браузера.';
			authState = 'login';
		}
		return () => { active = false; portfolioGeneration++; };
	});

	async function authenticate(mode: 'login' | 'register', user: string, pass: string) {
		if (busy) return;
		if (!client) throw new Error('Хранилище браузера недоступно.');
		busy = true;
		try {
			await client.authenticate(mode === 'login' ? '/login' : '/createuser', user, pass);
			startupError = ''; authenticated();
		} catch (error) {
			if (error instanceof HttpError && error.status === 401) throw new Error('Неверное имя пользователя или пароль.');
			if (error instanceof HttpError && error.status === 409) throw new Error('Пользователь с таким именем уже существует.');
			throw new Error('Не удалось связаться с сервером или сохранить сессию. Попробуйте ещё раз.');
		} finally { busy = false; }
	}

	function logout() {
		portfolioGeneration++; portfolio = null;
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
	<main class="min-h-svh bg-background p-4 sm:p-8">
		<div class="mx-auto max-w-4xl space-y-6">
			<header class="flex flex-wrap items-center justify-between gap-4 border-b border-border pb-4"><h1 class="text-xl font-semibold">Портфель</h1><Button variant="outline" onclick={logout}>Выйти</Button></header>
			{#if portfolioStatus === 'loading'}
				<p role="status" class="text-sm text-muted-foreground">Загрузка данных портфеля…</p>
			{:else if portfolioStatus === 'no-profile'}
				<div class="rounded-xl border border-dashed border-border p-6"><h2 class="font-medium">Нет выбранного профиля</h2><p class="mt-2 text-sm text-muted-foreground">Создайте или выберите профиль по умолчанию. Управление профилями в приложении появится на следующем этапе.</p></div>
			{:else if portfolioStatus === 'error'}
				<div role="alert" class="space-y-3"><p class="text-sm text-destructive">Не удалось загрузить данные портфеля. Проверьте соединение и повторите попытку.</p><Button variant="outline" onclick={loadPortfolio}>Повторить</Button></div>
			{:else if portfolio}
				{@const data = portfolio.data}
				{@const empty = data.stocks.length === 0 && data.bonds.length === 0 && data.crypto_assets.length === 0 && data.personal_assets.deposits.length === 0 && data.personal_assets.insurances.length === 0 && data.personal_assets.credit_cards.length === 0}
				<section aria-labelledby="overview-heading" class="space-y-4"><h2 id="overview-heading" class="text-lg font-semibold">Общая информация</h2>
					{#if empty}<p class="text-sm text-muted-foreground">В выбранном профиле пока нет активов.</p>{/if}
					<dl class="grid grid-cols-2 gap-4 rounded-xl border border-border bg-card p-4 sm:grid-cols-4 [&>div]:min-w-0 [&_dd]:wrap-anywhere">
						<div><dt class="text-xs text-muted-foreground">Акции</dt><dd class="mt-1 text-sm font-medium">{money(portfolio.stocks.total)}</dd></div>
						<div><dt class="text-xs text-muted-foreground">Облигации</dt><dd class="mt-1 text-sm font-medium">{money(portfolio.bonds.totalBondsValueRub)}</dd></div>
						<div><dt class="text-xs text-muted-foreground">Вклады</dt><dd class="mt-1 text-sm font-medium">{money(portfolio.personalTotals.deposits)}</dd></div>
						<div><dt class="text-xs text-muted-foreground">Страхование</dt><dd class="mt-1 text-sm font-medium">{money(portfolio.personalTotals.insurances)}</dd></div>
					</dl>
					{#if portfolio.unavailableRates.length}<p role="status" class="text-sm text-muted-foreground">Нет курса к рублю: {portfolio.unavailableRates.join(', ')}. Стоимость облигаций недоступна.</p>{/if}
					<p class="text-xs text-muted-foreground">Профиль: {data.profiles.find(profile => profile.default)?.name ?? 'Нет данных'}. Загружено: {new Date(portfolio.loadedAt).toLocaleString('ru-RU')}.</p>
				</section>
				<!-- TEMP: Detailed live tabs and mutations are connected in later steps. -->
				<p class="text-sm text-muted-foreground">Подробные разделы и редактирование портфеля ещё подключаются.</p>
			{/if}
		</div>
	</main>
{/if}
