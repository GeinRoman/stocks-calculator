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
	import ProfilePanel from '#lib/components/tabs/profile.svelte';
	import Stocks from '#lib/components/tabs/stocks.svelte';
	import Bonds from '#lib/components/tabs/bonds.svelte';
	import Rebalance from '#lib/components/tabs/rebalance.svelte';
	import PersonalAssets from '#lib/components/tabs/personal-assets.svelte';
	import CreditCards from '#lib/components/tabs/credit-cards.svelte';
	import Crypto from '#lib/components/tabs/crypto.svelte';
	import TransactionHistory from '#lib/components/tabs/transaction-history.svelte';
	import { demoCrypto, type CryptoSnapshot } from '#lib/logic/tabs/crypto.js';
	import { demoProfiles, demoPortfolios, emptyPortfolio } from '#lib/logic/tabs/demo-portfolio.js';
	import { createDemoBondPortfolios, createEmptyBondPortfolio } from '#lib/logic/tabs/demo-bonds.js';
	import type { BondPortfolioSnapshot } from '#lib/logic/tabs/bonds.js';
	import type { PortfolioSnapshot, Profile } from '#lib/logic/tabs/portfolio.js';
	import { createDemoPersonalAssets } from '#lib/logic/tabs/personal-assets-demo.js';
	import { emptyPersonalAssets, summarizePersonalAssets, type PersonalAssetsSnapshot } from '#lib/logic/tabs/personal-assets.js';
	import { createDemoCreditCards } from '#lib/logic/tabs/credit-card-demo.js';
	import { emptyCreditCards, type CreditCardsSnapshot } from '#lib/logic/tabs/credit-cards.js';
	import type { DemoBackup } from '#lib/logic/tabs/demo-backup.js';
	let { onlogout }: { onlogout: () => string | null } = $props();
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
	// TEMP
	let profiles = $state<Profile[]>(demoProfiles.map((profile) => ({ ...profile })));
	let selectedId = $state('main');
	let portfolios = $state<Record<string, PortfolioSnapshot>>(structuredClone(demoPortfolios));
	let stocksBusy = $state(false);
	let bondsBusy = $state(false);
	let rebalanceBusy = $state(false);
	let personalAssetsBusy = $state(false);
	let creditCardsBusy = $state(false);
	let cryptoBusy = $state(false);
	let backupBusy = $state(false);
	let restoreGeneration = $state(0);
	// TEMP
	let cryptoPortfolios = $state<Record<string, CryptoSnapshot>>(demoCrypto());
	const emptyCrypto: CryptoSnapshot = { groups: [] };
	const cryptoPortfolio = $derived(cryptoPortfolios[selectedId] ?? emptyCrypto);
	// TEMP
	let generalLoading = $state(true);
	// TEMP
	let bondPortfolios = $state<Record<string, BondPortfolioSnapshot>>(createDemoBondPortfolios());
	// TEMP
	let personalPortfolios = $state<Record<string, PersonalAssetsSnapshot>>(createDemoPersonalAssets());
	// TEMP
	let creditCardPortfolios = $state<Record<string, CreditCardsSnapshot>>(createDemoCreditCards());
	const emptyCards = emptyCreditCards();
	const emptyPersonal = emptyPersonalAssets();
	const emptyBonds = createEmptyBondPortfolio();
	const selectedProfile = $derived(profiles.find((profile) => profile.id === selectedId));
	const portfolio = $derived(portfolios[selectedId] ?? emptyPortfolio);
	const bonds = $derived(bondPortfolios[selectedId] ?? emptyBonds);
	const personal = $derived(personalPortfolios[selectedId] ?? emptyPersonal);
	const creditCards = $derived(creditCardPortfolios[selectedId] ?? emptyCards);
	const personalTotals = $derived(summarizePersonalAssets(personal));
	// TEMP: The demo backup captures only in-memory data, not backend snapshots or history.
	function getBackup(): DemoBackup {
		return { format: 'stocks-calculator-demo-backup', version: 1, selectedId, profiles: profiles.map(profile => ({ profile, stocks: portfolios[profile.id] ?? emptyPortfolio, bonds: bondPortfolios[profile.id] ?? emptyBonds, crypto: cryptoPortfolios[profile.id] ?? emptyCrypto, personal: personalPortfolios[profile.id] ?? emptyPersonal, creditCards: creditCardPortfolios[profile.id] ?? emptyCards })) };
	}
	function restoreBackup(backup: DemoBackup) {
		profiles = backup.profiles.map(entry => entry.profile);
		portfolios = Object.fromEntries(backup.profiles.map(entry => [entry.profile.id, entry.stocks]));
		bondPortfolios = Object.fromEntries(backup.profiles.map(entry => [entry.profile.id, entry.bonds]));
		cryptoPortfolios = Object.fromEntries(backup.profiles.map(entry => [entry.profile.id, entry.crypto]));
		personalPortfolios = Object.fromEntries(backup.profiles.map(entry => [entry.profile.id, entry.personal]));
		creditCardPortfolios = Object.fromEntries(backup.profiles.map(entry => [entry.profile.id, entry.creditCards]));
		selectedId = backup.selectedId; restoreGeneration++;
	}
	const tabClass =
		'h-auto w-full flex-none justify-start rounded-none px-4 py-5 whitespace-normal text-left leading-relaxed text-sidebar-foreground/70 hover:bg-sidebar-accent data-active:bg-sidebar-accent data-active:text-primary after:inset-y-0 after:right-0 after:w-1 after:bg-primary sm:px-6';

	// TEMP
	onMount(() => {
		const timer = setTimeout(() => generalLoading = false, 2000);
		const stop = observeSidebarViewport(window.matchMedia('(max-width: 767px)'), (isNarrow) => { narrow = isNarrow; mobileOpen = false; responsiveReady = true; });
		return () => { clearTimeout(timer); stop(); };
	});
</script>

<svelte:head>
	<title>Портфель</title>
	<meta name="description" content="Обзор состава и стоимости инвестиционного портфеля" />
</svelte:head>

{#snippet navigation()}
			<div class="border-b border-sidebar-border px-4 py-6 sm:px-6">
				<div class="flex items-center justify-between gap-1"><p class="min-w-0 flex-1 text-sm font-semibold tracking-tight text-sidebar-foreground sm:text-base">Портфель</p><Button variant="ghost" size="icon-sm" aria-label="Выйти" title="Выйти" onclick={() => { mobileOpen = false; logoutError = ''; logoutOpen = true; }}><LogOutIcon aria-hidden="true" /></Button><Button bind:ref={navigationClose} variant="ghost" size="icon-sm" aria-label="Закрыть меню" aria-expanded="true" aria-controls="portfolio-navigation" title="Закрыть меню" onclick={closeNavigation}><PanelLeftCloseIcon aria-hidden="true" /></Button></div>
				<p class="mt-1 break-words text-xs text-muted-foreground">{selectedProfile?.name ?? ''}</p>
			</div>
			<Tabs.List aria-label="Разделы портфеля" variant="line" class="flex w-full flex-col items-stretch gap-0 p-0 pt-4" onclick={(event) => { const target = event.target as HTMLElement; const trigger = target.closest('[data-slot="tabs-trigger"]'); if (narrow && trigger?.getAttribute('data-value') === activeTab && !trigger.hasAttribute('disabled')) queueMicrotask(() => mobileOpen = false); }}>
				<Tabs.Trigger value="profile" class={tabClass} disabled={generalLoading || stocksBusy || bondsBusy || rebalanceBusy || personalAssetsBusy || creditCardsBusy || cryptoBusy}>Профиль</Tabs.Trigger>
				<Tabs.Trigger value="general" class={tabClass} disabled={backupBusy}>Общая информация</Tabs.Trigger>
				<Tabs.Trigger value="stocks" class={tabClass} disabled={generalLoading || rebalanceBusy || backupBusy}>Акции</Tabs.Trigger>
				<Tabs.Trigger value="bonds" class={tabClass} disabled={generalLoading || rebalanceBusy || backupBusy}>Облигации</Tabs.Trigger>
				<Tabs.Trigger value="crypto" class={tabClass} disabled={generalLoading || rebalanceBusy || backupBusy}>Криптовалюта</Tabs.Trigger>
				<Tabs.Trigger value="personal" class={tabClass} disabled={generalLoading || rebalanceBusy || backupBusy}>Вклады и страхование</Tabs.Trigger>
				<Tabs.Trigger value="credit-cards" class={tabClass} disabled={generalLoading || rebalanceBusy || backupBusy}>Кредитные карты</Tabs.Trigger>
				<Tabs.Trigger value="rebalance" class={tabClass} disabled={generalLoading || stocksBusy || bondsBusy || personalAssetsBusy || creditCardsBusy || cryptoBusy || backupBusy}>Ребалансировка</Tabs.Trigger>
			</Tabs.List>
			{#if stocksBusy || bondsBusy || rebalanceBusy || personalAssetsBusy || creditCardsBusy || cryptoBusy}<p class="px-4 py-4 text-xs leading-relaxed text-muted-foreground sm:px-6">Завершите текущие изменения, чтобы перейти к другим операциям.</p>{/if}
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
				<!-- TEMP -->
				<ProfilePanel
					{profiles}
					{selectedId}
					username="user"
					{getBackup}
					{backupBusy}
					onrestore={restoreBackup}
					onbackupbusychange={(busy) => backupBusy = busy}
					onselect={(id) => selectedId = id}
					oncreate={(name) => {
						const id = crypto.randomUUID();
						profiles = [...profiles, { id, name }];
						selectedId = id;
					}}
					onrename={(id, name) => profiles = profiles.map((profile) => profile.id === id ? { ...profile, name } : profile)}
					ondelete={(id) => {
						if (profiles.length <= 1) return;
						profiles = profiles.filter((profile) => profile.id !== id);
						delete portfolios[id];
						delete bondPortfolios[id];
						delete personalPortfolios[id];
						delete creditCardPortfolios[id];
						delete cryptoPortfolios[id];
						selectedId = profiles[0].id;
					}}
				/>
			</Tabs.Content>
			<!-- TEMP -->
			<Tabs.Content value="general" class="min-h-full">
				{#if generalLoading}
					<!-- TEMP -->
					<div role="status" aria-live="polite" class="mx-auto flex max-w-6xl items-center gap-3 text-sm text-muted-foreground">
						<span aria-hidden="true" class="size-5 animate-spin rounded-full border-2 border-border border-t-primary motion-reduce:animate-none"></span>
						Загрузка общей информации…
					</div>
				{:else}
					{#key `${selectedId}:${restoreGeneration}`}<GeneralInformation values={{ stocks: portfolio.assetValues.stocks, bonds: bonds.totalBondsValueRub, deposits: personalTotals.deposits, insurances: personalTotals.insurances }} />{/key}
				{/if}
			</Tabs.Content>
			<Tabs.Content value="bonds" class="min-h-full">
				{#key `${selectedId}:${restoreGeneration}`}<Bonds portfolio={bonds} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => bondsBusy = busy} oncommit={(snapshot) => bondPortfolios = { ...bondPortfolios, [selectedId]: snapshot }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="stocks" class="min-h-full">
				{#key `${selectedId}:${restoreGeneration}`}<Stocks {portfolio} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => stocksBusy = busy} oncommit={(snapshot) => portfolios = { ...portfolios, [selectedId]: snapshot }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="crypto" class="min-h-full">
				{#key `${selectedId}:${restoreGeneration}`}<Crypto snapshot={cryptoPortfolio} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => cryptoBusy = busy} oncommit={(updated) => cryptoPortfolios = { ...cryptoPortfolios, [selectedId]: updated }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="rebalance" class="min-h-full">
				{#key `${selectedId}:${restoreGeneration}`}<Rebalance profileId={selectedId} profileName={selectedProfile?.name ?? ''} snapshots={{ stocks: portfolio, bonds }} onbusychange={(busy) => rebalanceBusy = busy} oncommit={(updated) => { portfolios = { ...portfolios, [selectedId]: updated.stocks }; bondPortfolios = { ...bondPortfolios, [selectedId]: updated.bonds }; }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="personal" class="min-h-full">
				{#key `${selectedId}:${restoreGeneration}`}<PersonalAssets snapshot={personal} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => personalAssetsBusy = busy} oncommit={(updated) => personalPortfolios = { ...personalPortfolios, [selectedId]: updated }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="credit-cards" class="min-h-full">
				{#key `${selectedId}:${restoreGeneration}`}<CreditCards snapshot={creditCards} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => creditCardsBusy = busy} oncommit={(updated) => creditCardPortfolios = { ...creditCardPortfolios, [selectedId]: updated }} />{/key}
			</Tabs.Content>
		</div>
	</Tabs.Root>
	{#if !generalLoading && !backupBusy && !rebalanceBusy}
		{#key `${selectedId}:${restoreGeneration}`}
			{#if activeTab === 'stocks'}<TransactionHistory kind="stock" profileId={selectedId} />
			{:else if activeTab === 'bonds'}<TransactionHistory kind="bond" profileId={selectedId} />
			{:else if activeTab === 'crypto'}<TransactionHistory kind="crypto" profileId={selectedId} />
			{:else if activeTab === 'personal'}<TransactionHistory kind="deposit" profileId={selectedId} stacked /><TransactionHistory kind="insurance" profileId={selectedId} />{/if}
		{/key}
	{/if}
</main>

<AlertDialog.Root bind:open={logoutOpen}>
	<AlertDialog.Content><AlertDialog.Header><AlertDialog.Title>Выйти из приложения?</AlertDialog.Title><AlertDialog.Description>Демонстрационные изменения, включая несохранённые правки, будут потеряны. После выхода потребуется снова войти.</AlertDialog.Description></AlertDialog.Header>
		{#if logoutError}<p role="alert" class="text-sm text-destructive">{logoutError}</p>{/if}
		<AlertDialog.Footer><AlertDialog.Cancel>Отмена</AlertDialog.Cancel><Button variant="destructive" onclick={() => { logoutError = onlogout() ?? ''; if (!logoutError) logoutOpen = false; }}>Выйти</Button></AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
