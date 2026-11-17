<script lang="ts">
	import { onMount, tick, untrack } from 'svelte';
	import { page } from '$app/state';
	import { providePortfolioBack } from '#lib/logic/portfolio-back.js';
	import { observeSidebarViewport } from '#lib/logic/sidebar-viewport.js';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import MenuIcon from '@lucide/svelte/icons/menu';
	import PanelLeftCloseIcon from '@lucide/svelte/icons/panel-left-close';
	import { Button, buttonVariants } from '#lib/components/ui/button/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import * as Drawer from '#lib/components/ui/drawer/index.js';
	import * as Tabs from '#lib/components/ui/tabs/index.js';
	import GeneralInformation from '#lib/components/tabs/general-information.svelte';
	import LiveProfile from '#lib/components/tabs/live-profile.svelte';
	import StockGroups from '#lib/components/tabs/stock-groups.svelte';
	import StockEditor from '#lib/components/tabs/stock-editor.svelte';
	import { StockSaveError } from '#lib/logic/live-stocks.js';
	import Bonds from '#lib/components/tabs/bonds.svelte';
	import PersonalAssets from '#lib/components/tabs/personal-assets.svelte';
	import CreditCards from '#lib/components/tabs/credit-cards.svelte';
	import Crypto from '#lib/components/tabs/crypto.svelte';
	import GroupWeightsDialog from '#lib/components/tabs/group-weights-dialog.svelte';
	import type { DraftGroup, GroupWeight, StockSearchResult } from '#lib/logic/tabs/portfolio.js';
	import { liveSnapshots } from '#lib/logic/live-snapshots.js';
	import type { Profile } from '#lib/api/contracts.js';
	import type { PortfolioState } from '#lib/logic/portfolio-state.js';
	import TransactionHistory from '#lib/components/tabs/transaction-history.svelte';
	import type { HistoryCursor, HistoryKind, HistoryTransaction } from '#lib/logic/tabs/transaction-history.js';
	import type { CouponSchedule } from '#lib/logic/tabs/coupons.js';
	import type { BondSearchResult, DraftBondGroup } from '#lib/logic/tabs/bonds.js';
	import type { CryptoCommand, CryptoSnapshot } from '#lib/logic/tabs/crypto.js';
	import { PersonalAssetSaveError, type PersonalAssetOperation } from '#lib/logic/live-personal-assets.js';
	import type { CreditCardOperation } from '#lib/logic/live-credit-cards.js';
	let { onlogout, portfolio, portfolioStatus, profiles, profilesKnown, profileBusy, onselect, oncreate, ondelete, onrename, onretry, onsaveweights, loadHistory, loadCoupons, downloadSnapshot, installSnapshot, findStocks, saveStocks, findBonds, saveBonds, saveCrypto, savePersonalAsset, saveCreditCard }: {
		onlogout: () => string | null; portfolio: PortfolioState | null;
		portfolioStatus: 'loading' | 'ready' | 'no-profile' | 'error'; profiles: Profile[]; profilesKnown: boolean; profileBusy: boolean;
		onselect: (profile: Profile) => Promise<void>; oncreate: (name: string) => Promise<void>; ondelete: (profile: Profile) => Promise<void>; onretry: () => void;
		onrename: (profile: Profile, name: string) => Promise<void>;
		onsaveweights: (kind: 'stock' | 'bond' | 'crypto', weights: GroupWeight[]) => Promise<void>;
		loadHistory: (kind: HistoryKind, profileId: string, cursor?: HistoryCursor) => Promise<{ items: HistoryTransaction[] }>;
		loadCoupons: (code: string) => Promise<CouponSchedule>; downloadSnapshot: () => Promise<void>; installSnapshot: (snapshot: unknown) => Promise<boolean>;
		findStocks: (query: string) => Promise<StockSearchResult[]>; saveStocks: (draft: DraftGroup[]) => Promise<void>;
		findBonds: (query: string) => Promise<BondSearchResult[]>; saveBonds: (draft: DraftBondGroup[]) => Promise<void>;
		saveCrypto: (commands: CryptoCommand[]) => Promise<void>;
		savePersonalAsset: (operation: PersonalAssetOperation) => Promise<void>;
		saveCreditCard: (operation: CreditCardOperation) => Promise<void>;
	} = $props();
	let stockEditing = $state(false), stockSaving = $state(false), stockBlocked = $state(false), stockDraft = $state<DraftGroup[]>([]);
	let bondEditing = $state(false), bondBlocked = $state(false), cryptoEditing = $state(false), cryptoBlocked = $state(false);
	let personalBusy = $state(false), personalBlocked = $state(false);
	let cardBusy = $state(false), cardBlocked = $state(false);
	async function commitPersonalAsset(operation: PersonalAssetOperation) {
		try { await savePersonalAsset(operation); }
		catch (error) { if (error instanceof PersonalAssetSaveError && error.uncertain) personalBlocked = true; throw error; }
	}
	async function commitCreditCard(operation: CreditCardOperation) {
		try { await saveCreditCard(operation); }
		catch (error) { if (error instanceof PersonalAssetSaveError && error.uncertain) cardBlocked = true; throw error; }
	}
	let cryptoPrices = $state<Record<string, number | null>>({}), cryptoCurrency = $state<CryptoSnapshot['priceCurrency']>('USDT');
	let priceProfile: string | undefined;
	$effect(() => {
		const name = profiles.find(profile => profile.default)?.name;
		if (name !== priceProfile) { priceProfile = name; cryptoPrices = {}; cryptoCurrency = 'USDT'; cryptoBlocked = false; }
	});
	function commitCryptoPrices(snapshot: CryptoSnapshot) {
		cryptoCurrency = snapshot.priceCurrency;
		cryptoPrices = Object.fromEntries(snapshot.groups.flatMap(group => group.assets.map(asset => [asset.id, asset.price])));
	}
	async function commitCrypto(commands: CryptoCommand[]) {
		try { await saveCrypto(commands); }
		catch (error) { cryptoBlocked = true; throw error; }
	}
	function editStocks() {
		if (!portfolio || stockBlocked) return;
		stockDraft = portfolio.stocks.groups.map(group => ({ id: group.id, name: group.name, targetPercent: group.targetPercent ?? null, stocks: group.stocks.map(stock => ({ code: stock.code, name: stock.name, price: stock.price, lotSize: stock.lotSize, lotCount: stock.lotCount })) }));
		stockEditing = true;
	}
	async function commitStocks(draft: DraftGroup[]) {
		stockSaving = true;
		try { await saveStocks(draft); stockEditing = false; }
		catch (error) { if (error instanceof StockSaveError && error.uncertain) stockBlocked = true; throw error; }
		finally { stockSaving = false; }
	}
	let weightsKind = $state<'stock' | 'bond' | 'crypto' | null>(null), weightsSaving = $state(false);
	async function saveWeights(weights: GroupWeight[]) {
		if (!weightsKind || weightsSaving) return;
		weightsSaving = true;
		try { await onsaveweights(weightsKind, weights); weightsKind = null; }
		finally { weightsSaving = false; }
	}
	let logoutOpen = $state(false), logoutError = $state('');
	let responsiveReady = $state(false), narrow = $state(false), desktopOpen = $state(true), mobileOpen = $state(false);
	let desktopTrigger = $state<HTMLButtonElement | null>(null), navigationClose = $state<HTMLButtonElement | null>(null);
	let tabsRoot = $state<HTMLElement | null>(null);
	function closeNavigation() {
		if (narrow) mobileOpen = false;
		else { desktopOpen = false; void tick().then(() => desktopTrigger?.focus()); }
	}
	function openDesktopNavigation() { desktopOpen = true; void tick().then(() => navigationClose?.focus()); }
	function selectTab(value: string) {
		if ((weightsKind || stockEditing || bondEditing || cryptoEditing || personalBusy || cardBusy) && value !== activeTab) return;
		if (narrow && mobileOpen) {
			mobileOpen = false;
			void tick().then(() => back.afterSettled(() => activeTab = value));
		} else activeTab = value;
	}

	let activeTab = $state('general');
	const back = providePortfolioBack();
	let previousTab = 'general';
	let releaseTab: (() => void) | undefined;
	$effect(() => {
		const tab = activeTab;
		if (tab === previousTab) return;
		if (previousTab === 'general' && tab !== 'general') releaseTab = untrack(() => back.enter('tab', () => activeTab = 'general'));
		else if (previousTab !== 'general' && tab === 'general') { untrack(() => releaseTab?.()); releaseTab = undefined; }
		previousTab = tab;
	});
	$effect(() => { const ids = page.state.portfolioBack; untrack(() => back.observe(ids)); });
	$effect(() => { if (narrow && mobileOpen) return untrack(() => back.enter('drawer', () => mobileOpen = false)); });
	const selectedProfile = $derived(profiles.find(profile => profile.default));
	const snapshots = $derived(portfolio ? liveSnapshots(portfolio.data) : null);
	const cryptoSnapshot = $derived(snapshots ? { ...snapshots.crypto, priceCurrency: cryptoCurrency, groups: snapshots.crypto.groups.map(group => ({ ...group, assets: group.assets.map(asset => ({ ...asset, price: cryptoPrices[asset.id] ?? null })) })) } : null);
	const tabClass =
		'h-auto w-full flex-none justify-start rounded-none px-4 py-5 whitespace-normal text-left leading-relaxed text-sidebar-foreground/70 hover:bg-sidebar-accent data-active:bg-sidebar-accent data-active:text-primary after:inset-y-0 after:right-0 after:w-1 after:bg-primary sm:px-6';

	onMount(() => {
		const stop = observeSidebarViewport(window.matchMedia('(max-width: 767px)'), (isNarrow) => { narrow = isNarrow; mobileOpen = false; responsiveReady = true; });
		return stop;
	});
</script>

<svelte:head>
	<title>Портфель</title>
	<meta name="description" content="Обзор состава и стоимости инвестиционного портфеля" />
</svelte:head>

{#snippet navigation()}
			<div class="border-b border-sidebar-border px-4 py-6 sm:px-6">
				<div class="flex items-center justify-between gap-1"><p class="min-w-0 flex-1 text-sm font-semibold tracking-tight text-sidebar-foreground sm:text-base">Портфель</p><Button variant="ghost" size="icon-sm" aria-label="Выйти" title="Выйти" disabled={!!weightsKind || stockEditing || bondEditing || cryptoEditing || personalBusy || cardBusy} onclick={() => { mobileOpen = false; logoutError = ''; logoutOpen = true; }}><LogOutIcon aria-hidden="true" /></Button><Button bind:ref={navigationClose} variant="ghost" size="icon-sm" aria-label="Закрыть меню" aria-expanded="true" aria-controls="portfolio-navigation" title="Закрыть меню" onclick={closeNavigation}><PanelLeftCloseIcon aria-hidden="true" /></Button></div>
				<p class="mt-1 break-words text-xs text-muted-foreground">{selectedProfile?.name ?? ''}</p>
			</div>
			<Tabs.List aria-label="Разделы портфеля" variant="line" class="flex w-full flex-col items-stretch gap-0 p-0 pt-4" onclick={(event) => { const target = event.target as HTMLElement; const trigger = target.closest('[data-slot="tabs-trigger"]'); if (narrow && trigger?.getAttribute('data-value') === activeTab && !trigger.hasAttribute('disabled')) queueMicrotask(() => mobileOpen = false); }}>
				<Tabs.Trigger value="profile" class={tabClass} disabled={!profilesKnown || profileBusy || !!weightsKind || stockEditing || bondEditing || cryptoEditing || personalBusy || cardBusy}>Профиль</Tabs.Trigger>
				<Tabs.Trigger value="general" class={tabClass} disabled={!!weightsKind || stockEditing || bondEditing || cryptoEditing || personalBusy || cardBusy}>Общая информация</Tabs.Trigger>
				<Tabs.Trigger value="stocks" class={tabClass} disabled={!portfolio || profileBusy || bondEditing || cryptoEditing || personalBusy || cardBusy || (weightsKind !== null && weightsKind !== 'stock')}>Акции</Tabs.Trigger>
				<Tabs.Trigger value="bonds" class={tabClass} disabled={!portfolio || profileBusy || stockEditing || cryptoEditing || personalBusy || cardBusy || (weightsKind !== null && weightsKind !== 'bond')}>Облигации</Tabs.Trigger>
				<Tabs.Trigger value="crypto" class={tabClass} disabled={!portfolio || profileBusy || stockEditing || bondEditing || personalBusy || cardBusy || (weightsKind !== null && weightsKind !== 'crypto')}>Криптовалюта</Tabs.Trigger>
				<Tabs.Trigger value="personal" class={tabClass} disabled={!portfolio || profileBusy || !!weightsKind || stockEditing || bondEditing || cryptoEditing || cardBusy}>Вклады и страхование</Tabs.Trigger>
				<Tabs.Trigger value="credit-cards" class={tabClass} disabled={!portfolio || profileBusy || !!weightsKind || stockEditing || bondEditing || cryptoEditing || personalBusy}>Кредитные карты</Tabs.Trigger>
				<Tabs.Trigger value="rebalance" class={tabClass} disabled>Ребалансировка</Tabs.Trigger>
			</Tabs.List>
{/snippet}

<main class="min-h-svh bg-background">
	<Tabs.Root bind:ref={tabsRoot} orientation="vertical" value={activeTab} onValueChange={selectTab} activationMode={narrow ? 'manual' : 'automatic'} class="min-h-svh gap-0">
		{#if responsiveReady && narrow}
			<Drawer.Root direction="left" bind:open={mobileOpen} shouldScaleBackground={false}>
				<Drawer.Trigger aria-label="Открыть меню" title="Открыть меню" class={`${buttonVariants({ size: 'icon-lg' })} fixed bottom-4 left-4 z-30 shadow-lg`}><MenuIcon aria-hidden="true" /></Drawer.Trigger>
				<Drawer.Content id="portfolio-navigation" portalProps={{ to: tabsRoot ?? undefined }} class="h-[100dvh] overflow-y-auto bg-sidebar p-0 before:inset-0 before:rounded-none data-[vaul-drawer-direction=left]:w-[min(85vw,16rem)]">
					<Drawer.Title class="sr-only">Меню портфеля</Drawer.Title><Drawer.Description class="sr-only">Навигация по разделам портфеля</Drawer.Description>
					<div class="group/tabs" data-vertical>{@render navigation()}</div>
				</Drawer.Content>
			</Drawer.Root>
		{:else if responsiveReady && desktopOpen}
			<aside id="portfolio-navigation" aria-label="Меню портфеля" class="sticky top-0 h-svh w-56 shrink-0 self-start overflow-y-auto border-r border-sidebar-border bg-sidebar lg:w-64">{@render navigation()}</aside>
		{:else if responsiveReady}
			<Button bind:ref={desktopTrigger} size="icon-lg" class="fixed bottom-4 left-4 z-30 shadow-lg" aria-label="Открыть меню" aria-expanded="false" title="Открыть меню" onclick={openDesktopNavigation}><MenuIcon aria-hidden="true" /></Button>
		{/if}

		<div class={activeTab === 'stocks' || activeTab === 'bonds' || activeTab === 'crypto' ? 'min-w-0 flex-1 p-4 pt-2 sm:px-8 sm:pb-8 sm:pt-2 lg:px-12 lg:pb-12 lg:pt-2' : 'min-w-0 flex-1 p-4 pt-8 sm:p-8 lg:p-12'}>
			<Tabs.Content value="profile" class="min-h-full">
				<LiveProfile {profiles} busy={profileBusy} {onselect} {oncreate} {ondelete} {onrename} {downloadSnapshot} {installSnapshot} />
			</Tabs.Content>
			<Tabs.Content value="general" class="min-h-full">
				{#if portfolioStatus === 'loading'}
					<div role="status" aria-live="polite" class="mx-auto flex max-w-6xl items-center gap-3 text-sm text-muted-foreground">
						<span aria-hidden="true" class="size-5 animate-spin rounded-full border-2 border-border border-t-primary motion-reduce:animate-none"></span>
						Загрузка общей информации…
					</div>
				{:else if portfolioStatus === 'no-profile'}
					<div class="mx-auto max-w-6xl space-y-4"><h1 class="text-xl font-semibold">Нет выбранного профиля</h1><p class="text-sm text-muted-foreground">Создайте или выберите профиль по умолчанию.</p><Button onclick={() => selectTab('profile')}>Профиль</Button></div>
				{:else if portfolioStatus === 'error'}
					<div role="alert" class="mx-auto max-w-6xl space-y-4"><p class="text-sm text-destructive">Не удалось загрузить данные портфеля. Проверьте соединение и повторите попытку.</p><Button variant="outline" onclick={onretry}>Повторить</Button></div>
				{:else if portfolio}
					<GeneralInformation demo={false} values={{ stocks: portfolio.stocks.total, bonds: portfolio.bonds.totalBondsValueRub, deposits: portfolio.personalTotals.deposits, insurances: portfolio.personalTotals.insurances }} />
					{#if portfolio.unavailableRates.length}<p role="status" class="mx-auto mt-4 max-w-6xl text-sm text-muted-foreground">Нет курса к рублю: {portfolio.unavailableRates.join(', ')}. Стоимость облигаций недоступна.</p>{/if}
				{/if}
			</Tabs.Content>
			{#if portfolio && snapshots}
				<Tabs.Content value="stocks" class="min-h-full">{#if stockEditing}<StockEditor bind:groups={stockDraft} profileName={selectedProfile?.name ?? ''} saving={stockSaving} blocked={stockBlocked} search={findStocks} onsave={commitStocks} oncancel={() => stockEditing = false} />{:else}<StockGroups portfolio={{ totalStocksValue: portfolio.stocks.total, groups: portfolio.stocks.groups }} profileName={selectedProfile?.name ?? ''} demo={false}>{#snippet controls()}<div class="flex flex-wrap gap-2"><Button disabled={stockBlocked} onclick={editStocks}>Редактировать состав</Button><Button variant="outline" disabled={!portfolio.stocks.groups.length || stockBlocked} onclick={() => weightsKind = 'stock'}>Настроить доли</Button>{#if stockBlocked}<Button variant="outline" onclick={() => window.location.reload()}>Обновить страницу</Button>{/if}</div>{#if stockBlocked}<p role="alert" class="mt-2 text-sm text-destructive">Результат запроса неизвестен. Обновите страницу перед повторным редактированием.</p>{/if}{/snippet}</StockGroups>{/if}</Tabs.Content>
				<Tabs.Content value="bonds" class="min-h-full"><Bonds portfolio={portfolio.bonds} rates={portfolio.rates} profileName={selectedProfile?.name ?? ''} readOnly {loadCoupons} search={findBonds} onSaveComposition={saveBonds} bind:blocked={bondBlocked} onEditWeights={() => weightsKind = 'bond'} onbusychange={(busy) => bondEditing = busy} oncommit={() => {}} /></Tabs.Content>
				<Tabs.Content value="crypto" class="min-h-full">{#if cryptoSnapshot}<Crypto snapshot={cryptoSnapshot} profileName={selectedProfile?.name ?? ''} readOnly onEditWeights={() => weightsKind = 'crypto'} onbusychange={(busy) => { cryptoEditing = busy; }} oncommit={commitCryptoPrices} onSaveComposition={commitCrypto} blocked={cryptoBlocked} />{/if}</Tabs.Content>
				<Tabs.Content value="personal" class="min-h-full"><PersonalAssets snapshot={snapshots.personal} profileName={selectedProfile?.name ?? ''} readOnly onbusychange={(busy) => { personalBusy = busy; }} oncommit={() => {}} onmutate={commitPersonalAsset} blocked={personalBlocked} /></Tabs.Content>
				<Tabs.Content value="credit-cards" class="min-h-full"><CreditCards snapshot={snapshots.cards} profileName={selectedProfile?.name ?? ''} readOnly onbusychange={(busy) => { cardBusy = busy; }} oncommit={() => {}} onmutate={commitCreditCard} blocked={cardBlocked} /></Tabs.Content>
			{/if}
			{#if portfolio && selectedProfile && !weightsKind && !stockEditing && !bondEditing && !cryptoEditing && !personalBusy && !cardBusy}
				{#key `${selectedProfile.name}-${activeTab}`}
					{#if activeTab === 'stocks'}<TransactionHistory kind="stock" profileId={selectedProfile.name} loadPage={loadHistory} />
					{:else if activeTab === 'bonds'}<TransactionHistory kind="bond" profileId={selectedProfile.name} loadPage={loadHistory} />
					{:else if activeTab === 'crypto'}<TransactionHistory kind="crypto" profileId={selectedProfile.name} loadPage={loadHistory} />
					{:else if activeTab === 'personal'}<TransactionHistory kind="deposit" profileId={selectedProfile.name} loadPage={loadHistory} /><TransactionHistory kind="insurance" profileId={selectedProfile.name} loadPage={loadHistory} stacked />{/if}
				{/key}
			{/if}
		</div>
	</Tabs.Root>
</main>

{#if weightsKind && portfolio && snapshots}<GroupWeightsDialog groups={weightsKind === 'stock' ? portfolio.stocks.groups : weightsKind === 'bond' ? portfolio.bonds.groups : snapshots.crypto.groups.map(group => ({ id: group.id, name: group.name, targetPercent: group.weight }))} saving={weightsSaving} onsave={saveWeights} onclose={() => { if (!weightsSaving) weightsKind = null; }} />{/if}

<AlertDialog.Root bind:open={logoutOpen}>
	<AlertDialog.Content><AlertDialog.Header><AlertDialog.Title>Выйти из приложения?</AlertDialog.Title><AlertDialog.Description>После выхода потребуется снова войти.</AlertDialog.Description></AlertDialog.Header>
		{#if logoutError}<p role="alert" class="text-sm text-destructive">{logoutError}</p>{/if}
		<AlertDialog.Footer><AlertDialog.Cancel>Отмена</AlertDialog.Cancel><Button variant="destructive" onclick={() => { logoutError = onlogout() ?? ''; if (!logoutError) logoutOpen = false; }}>Выйти</Button></AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
