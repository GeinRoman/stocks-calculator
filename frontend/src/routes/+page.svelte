<script lang="ts">
	import { onMount } from 'svelte';
	import * as Tabs from '#lib/components/ui/tabs/index.js';
	import GeneralInformation from '#lib/components/tabs/general-information.svelte';
	import ProfilePanel from '#lib/components/tabs/profile.svelte';
	import Stocks from '#lib/components/tabs/stocks.svelte';
	import Bonds from '#lib/components/tabs/bonds.svelte';
	import Rebalance from '#lib/components/tabs/rebalance.svelte';
	import PersonalAssets from '#lib/components/tabs/personal-assets.svelte';
	import { demoProfiles, demoPortfolios, emptyPortfolio } from '#lib/logic/tabs/demo-portfolio.js';
	import { createDemoBondPortfolios, createEmptyBondPortfolio } from '#lib/logic/tabs/demo-bonds.js';
	import type { BondPortfolioSnapshot } from '#lib/logic/tabs/bonds.js';
	import type { PortfolioSnapshot, Profile } from '#lib/logic/tabs/portfolio.js';
	import { createDemoPersonalAssets } from '#lib/logic/tabs/personal-assets-demo.js';
	import { emptyPersonalAssets, summarizePersonalAssets, type PersonalAssetsSnapshot } from '#lib/logic/tabs/personal-assets.js';

	let activeTab = $state('general');
	// TEMP
	let profiles = $state<Profile[]>(demoProfiles.map((profile) => ({ ...profile })));
	let selectedId = $state('main');
	let portfolios = $state<Record<string, PortfolioSnapshot>>(structuredClone(demoPortfolios));
	let stocksBusy = $state(false);
	let bondsBusy = $state(false);
	let rebalanceBusy = $state(false);
	let personalAssetsBusy = $state(false);
	// TEMP
	let generalLoading = $state(true);
	// TEMP
	let bondPortfolios = $state<Record<string, BondPortfolioSnapshot>>(createDemoBondPortfolios());
	// TEMP
	let personalPortfolios = $state<Record<string, PersonalAssetsSnapshot>>(createDemoPersonalAssets());
	const emptyPersonal = emptyPersonalAssets();
	const emptyBonds = createEmptyBondPortfolio();
	const selectedProfile = $derived(profiles.find((profile) => profile.id === selectedId));
	const portfolio = $derived(portfolios[selectedId] ?? emptyPortfolio);
	const bonds = $derived(bondPortfolios[selectedId] ?? emptyBonds);
	const personal = $derived(personalPortfolios[selectedId] ?? emptyPersonal);
	const personalTotals = $derived(summarizePersonalAssets(personal));
	const tabClass =
		'group-data-vertical/tabs:h-auto group-data-vertical/tabs:rounded-none group-data-vertical/tabs:px-4 group-data-vertical/tabs:py-5 whitespace-normal text-left leading-relaxed text-sidebar-foreground/70 hover:bg-sidebar-accent data-active:bg-sidebar-accent data-active:text-primary after:bg-primary group-data-vertical/tabs:after:right-0 group-data-vertical/tabs:after:w-1 sm:group-data-vertical/tabs:px-6';

	// TEMP
	onMount(() => {
		const timer = setTimeout(() => generalLoading = false, 2000);
		return () => clearTimeout(timer);
	});
</script>

<svelte:head>
	<title>Портфель</title>
	<meta name="description" content="Обзор состава и стоимости инвестиционного портфеля" />
</svelte:head>

<main class="min-h-svh bg-background">
	<Tabs.Root orientation="vertical" bind:value={activeTab} class="min-h-svh gap-0">
		<aside class="sticky top-0 h-svh w-32 shrink-0 self-start overflow-y-auto border-r border-sidebar-border bg-sidebar sm:w-56 lg:w-64">
			<div class="border-b border-sidebar-border px-4 py-8 sm:px-6">
				<p class="text-sm font-semibold tracking-tight text-sidebar-foreground sm:text-base">Портфель</p>
				<p class="mt-1 break-words text-xs text-muted-foreground">{selectedProfile?.name ?? ''}</p>
			</div>
			<Tabs.List aria-label="Разделы портфеля" variant="line" class="w-full gap-0 p-0 pt-4">
				<Tabs.Trigger value="profile" class={tabClass} disabled={generalLoading || stocksBusy || bondsBusy || rebalanceBusy || personalAssetsBusy}>Профиль</Tabs.Trigger>
				<Tabs.Trigger value="general" class={tabClass}>Общая информация</Tabs.Trigger>
				<Tabs.Trigger value="stocks" class={tabClass} disabled={generalLoading || rebalanceBusy}>Акции</Tabs.Trigger>
				<Tabs.Trigger value="bonds" class={tabClass} disabled={generalLoading || rebalanceBusy}>Облигации</Tabs.Trigger>
				<Tabs.Trigger value="personal" class={tabClass} disabled={generalLoading || rebalanceBusy}>Вклады и страхование</Tabs.Trigger>
				<Tabs.Trigger value="rebalance" class={tabClass} disabled={generalLoading || stocksBusy || bondsBusy || personalAssetsBusy}>Ребалансировка</Tabs.Trigger>
			</Tabs.List>
			{#if stocksBusy || bondsBusy || rebalanceBusy || personalAssetsBusy}<p class="px-4 py-4 text-xs leading-relaxed text-muted-foreground sm:px-6">Завершите текущие изменения, чтобы перейти к другим операциям.</p>{/if}
		</aside>

		<div class="min-w-0 flex-1 p-4 pt-8 sm:p-8 lg:p-12">
			<Tabs.Content value="profile" class="min-h-full">
				<!-- TEMP -->
				<ProfilePanel
					{profiles}
					{selectedId}
					username="Иван Иванов"
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
					{#key selectedId}<GeneralInformation values={{ stocks: portfolio.assetValues.stocks, bonds: bonds.totalBondsValueRub, deposits: personalTotals.deposits, insurances: personalTotals.insurances }} />{/key}
				{/if}
			</Tabs.Content>
			<Tabs.Content value="bonds" class="min-h-full">
				{#key selectedId}<Bonds portfolio={bonds} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => bondsBusy = busy} oncommit={(snapshot) => bondPortfolios = { ...bondPortfolios, [selectedId]: snapshot }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="stocks" class="min-h-full">
				{#key selectedId}<Stocks {portfolio} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => stocksBusy = busy} oncommit={(snapshot) => portfolios = { ...portfolios, [selectedId]: snapshot }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="rebalance" class="min-h-full">
				{#key selectedId}<Rebalance profileId={selectedId} profileName={selectedProfile?.name ?? ''} snapshots={{ stocks: portfolio, bonds }} onbusychange={(busy) => rebalanceBusy = busy} oncommit={(updated) => { portfolios = { ...portfolios, [selectedId]: updated.stocks }; bondPortfolios = { ...bondPortfolios, [selectedId]: updated.bonds }; }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="personal" class="min-h-full">
				{#key selectedId}<PersonalAssets snapshot={personal} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => personalAssetsBusy = busy} oncommit={(updated) => personalPortfolios = { ...personalPortfolios, [selectedId]: updated }} />{/key}
			</Tabs.Content>
		</div>
	</Tabs.Root>
</main>
