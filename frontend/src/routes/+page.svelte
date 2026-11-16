<script lang="ts">
	import { onMount } from 'svelte';
	import LoginPage from '#lib/components/login-page.svelte';
	import { ApiClient, HttpError, REFRESH_TOKEN_KEY } from '#lib/api/client.js';
	import { portfolioApi } from '#lib/api/portfolio-api.js';
	import { loadPortfolioState, type PortfolioState } from '#lib/logic/portfolio-state.js';
	import PortfolioApp from '#lib/components/portfolio-app.svelte';
	import type { Profile } from '#lib/api/contracts.js';
	import type { GroupWeight } from '#lib/logic/tabs/portfolio.js';
	import type { DraftGroup } from '#lib/logic/tabs/portfolio.js';
	import { saveLiveStocks, searchLiveStocks } from '#lib/logic/live-stocks.js';
	import { saveLiveBonds, searchLiveBonds } from '#lib/logic/live-bonds.js';
	import { bondView } from '#lib/api/bond-view.js';
	import type { DraftBondGroup } from '#lib/logic/tabs/bonds.js';
	import { saveLiveWeights } from '#lib/logic/live-weights.js';
	import { historyLoader } from '#lib/api/transaction-history.js';
	import { couponSchedule } from '#lib/api/coupons.js';
	import type { HistoryCursor, HistoryKind } from '#lib/logic/tabs/transaction-history.js';
	import { saveLiveCrypto } from '#lib/logic/live-crypto.js';
	import type { CryptoCommand } from '#lib/logic/tabs/crypto.js';

	let authState = $state<'checking' | 'login' | 'authenticated'>('checking');
	let busy = $state(false), startupError = $state('');
	let portfolio = $state<PortfolioState | null>(null);
	let portfolioStatus = $state<'loading' | 'ready' | 'no-profile' | 'error'>('loading');
	let profiles = $state<Profile[]>([]), profilesKnown = $state(false), profileBusy = $state(false);
	let portfolioGeneration = 0;
	let client: ApiClient | undefined;
	async function findStocks(query: string) {
		if (!client) throw new Error('Сессия недоступна.');
		return searchLiveStocks(portfolioApi(client).stocks, query);
	}
	async function saveStocks(draft: DraftGroup[]) {
		if (!client || !portfolio || profileBusy) throw new Error('Портфель недоступен.');
		const generation = portfolioGeneration;
		await saveLiveStocks(portfolioApi(client).stocks, portfolio, draft, updated => {
			if (generation === portfolioGeneration) portfolio = updated;
		});
	}
	async function findBonds(query: string) {
		if (!client || !portfolio) throw new Error('Портфель недоступен.');
		const generation = portfolioGeneration;
		const api = portfolioApi(client);
		const results = await searchLiveBonds(api.bonds, query);
		const currencies = [...new Set(results.flatMap(bond => [bond.cleanPrice.currency, bond.accruedInterest.currency, bond.maturityPayment?.currency]).filter((currency): currency is string => !!currency && currency !== 'RUB' && !((portfolio?.rates[currency] ?? 0) > 0)))];
		const fetched = await Promise.allSettled(currencies.map(currency => api.currencyRate(currency, 'RUB')));
		if (generation !== portfolioGeneration || !portfolio) throw new Error('Профиль изменился. Повторите поиск.');
		const rates = { ...portfolio.rates };
		for (const [index, result] of fetched.entries()) {
			if (result.status === 'fulfilled' && result.value.from === currencies[index] && result.value.to === 'RUB' && Number.isFinite(result.value.rate) && result.value.rate > 0) rates[currencies[index]] = result.value.rate;
		}
		portfolio = { ...portfolio, rates, bonds: bondView(portfolio.data, rates, portfolio.loadedAt), unavailableRates: portfolio.unavailableRates.filter(currency => !(rates[currency] > 0)) };
		return results;
	}
	async function saveBonds(draft: DraftBondGroup[]) {
		if (!client || !portfolio || profileBusy) throw new Error('Портфель недоступен.');
		const generation = portfolioGeneration;
		await saveLiveBonds(portfolioApi(client).bonds, portfolio, draft, updated => {
			if (generation === portfolioGeneration) portfolio = updated;
		});
	}
	async function saveCrypto(commands: CryptoCommand[]) {
		if (!client || !portfolio || profileBusy) throw new Error('Портфель недоступен.');
		if (!commands.length) return;
		await saveLiveCrypto(portfolioApi(client).crypto, commands);
		await loadPortfolio();
	}
	async function loadHistory(kind: HistoryKind, profileId: string, cursor?: HistoryCursor) {
		if (!client) throw new Error('Сессия недоступна.');
		return historyLoader(portfolioApi(client))(kind, profileId, cursor);
	}
	async function loadCoupons(code: string) {
		if (!client) throw new Error('Сессия недоступна.');
		return couponSchedule(code, await portfolioApi(client).bonds.coupons(code));
	}
	async function downloadSnapshot() {
		if (!client) throw new Error('Сессия недоступна.');
		const data = await portfolioApi(client).snapshot();
		const url = URL.createObjectURL(new Blob([data], { type: 'application/json' }));
		try { const anchor = document.createElement('a'); anchor.href = url; anchor.download = 'snapshot.json'; anchor.click(); }
		finally { setTimeout(() => URL.revokeObjectURL(url), 1000); }
	}
	async function installSnapshot(snapshot: unknown) {
		if (!client || profileBusy) throw new Error('Портфель недоступен.');
		profileBusy = true;
		try {
			await portfolioApi(client).installSnapshot(snapshot);
			await loadPortfolio();
			return portfolioStatus !== 'error';
		} finally { profileBusy = false; }
	}

	async function loadPortfolio() {
		if (!client) return;
		const generation = ++portfolioGeneration;
		portfolioStatus = 'loading'; portfolio = null;
		try {
			const result = await loadPortfolioState(portfolioApi(client));
			if (generation !== portfolioGeneration) return;
			portfolio = result; profiles = result.data.profiles; profilesKnown = true; portfolioStatus = 'ready';
		} catch (error) {
			if (generation !== portfolioGeneration) return;
			if (error instanceof HttpError && error.status === 404 && error.message === 'Default profile not found') {
				try {
					const list = await portfolioApi(client).profiles.list();
					if (generation !== portfolioGeneration) return;
					profiles = list; profilesKnown = true; portfolioStatus = 'no-profile';
				} catch { if (generation === portfolioGeneration) portfolioStatus = 'error'; }
			} else portfolioStatus = 'error';
		}
	}
	function setProfiles(next: Profile[]) {
		profiles = next;
		if (portfolio) portfolio = { ...portfolio, data: { ...portfolio.data, profiles: next } };
	}
	async function saveWeights(kind: 'stock' | 'bond' | 'crypto', weights: GroupWeight[]) {
		if (!client || !portfolio || profileBusy) throw new Error('Портфель недоступен.');
		const generation = portfolioGeneration;
		const result = await saveLiveWeights(portfolioApi(client), portfolio, kind, weights);
		if (generation === portfolioGeneration) portfolio = result;
	}
	async function selectProfile(profile: Profile) {
		if (!client || profileBusy) return;
		profileBusy = true;
		const generation = portfolioGeneration;
		try {
			await portfolioApi(client).profiles.setDefault(profile);
			if (generation !== portfolioGeneration) return;
			setProfiles(profiles.map(item => ({ ...item, default: item.name === profile.name })));
			await loadPortfolio();
		} catch { throw new Error('Не удалось сменить профиль. Попробуйте ещё раз.'); }
		finally { profileBusy = false; }
	}
	async function createProfile(name: string) {
		if (!client || profileBusy) return;
		profileBusy = true;
		const generation = portfolioGeneration;
		const profile = { name, default: !profiles.some(item => item.default) };
		try {
			await portfolioApi(client).profiles.create(profile);
			if (generation !== portfolioGeneration) return;
			setProfiles([...profiles, profile]);
			if (profile.default) await loadPortfolio();
		} catch (error) {
			if (error instanceof HttpError && error.status === 409) throw new Error('Профиль с таким названием уже существует.');
			throw new Error('Не удалось создать профиль. Попробуйте ещё раз.');
		} finally { profileBusy = false; }
	}
	async function deleteProfile(profile: Profile) {
		if (!client || profileBusy) return;
		profileBusy = true;
		const generation = portfolioGeneration;
		try {
			await portfolioApi(client).profiles.remove(profile);
			if (generation !== portfolioGeneration) return;
			setProfiles(profiles.filter(item => item.name !== profile.name));
			if (profile.default) { portfolioGeneration++; portfolio = null; portfolioStatus = 'no-profile'; }
		} catch { throw new Error('Не удалось удалить профиль. Попробуйте ещё раз.'); }
		finally { profileBusy = false; }
	}
	async function renameProfile(profile: Profile, name: string) {
		if (!client || profileBusy) throw new Error('Портфель недоступен.');
		profileBusy = true;
		const generation = portfolioGeneration;
		try {
			await portfolioApi(client).profiles.rename(profile.name, name);
			if (generation !== portfolioGeneration) return;
			setProfiles(profiles.map(item => item.name === profile.name ? { ...item, name } : item));
		} catch (error) {
			if (error instanceof HttpError && error.status === 409) throw new Error('Профиль с таким названием уже существует.');
			if (error instanceof HttpError && error.status === 404) throw new Error('Профиль не найден. Обновите страницу.');
			throw new Error('Не удалось переименовать профиль. Попробуйте ещё раз.');
		} finally { profileBusy = false; }
	}
	function authenticated() { authState = 'authenticated'; void loadPortfolio(); }

	onMount(() => {
		let active = true;
		try {
			const storage = window.localStorage;
			client = new ApiClient({ baseUrl: '/api', storage, onSessionExpired: () => {
				if (active) { portfolioGeneration++; portfolio = null; profiles = []; profilesKnown = false; startupError = 'Сессия завершена. Войдите снова.'; authState = 'login'; }
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

	function logout(): string | null {
		portfolioGeneration++; portfolio = null; profiles = []; profilesKnown = false;
		try { client?.logout(); startupError = ''; }
		catch { startupError = 'Не удалось удалить сессию из хранилища браузера.'; }
		authState = 'login';
		return null;
	}
</script>

{#if authState === 'checking'}
	<main class="flex min-h-svh items-center justify-center bg-background p-4"><p role="status" class="text-sm text-muted-foreground">Проверка сессии…</p></main>
{:else if authState === 'login'}
	<LoginPage {busy} error={startupError} onauthenticate={authenticate} />
{:else}
	<PortfolioApp {portfolio} {portfolioStatus} {profiles} {profilesKnown} {profileBusy} onselect={selectProfile} oncreate={createProfile} ondelete={deleteProfile} onrename={renameProfile} onretry={loadPortfolio} onlogout={logout} onsaveweights={saveWeights} {loadHistory} {loadCoupons} {downloadSnapshot} {installSnapshot} {findStocks} {saveStocks} {findBonds} {saveBonds} {saveCrypto} />
{/if}
