<script lang="ts">
	import * as Tabs from '#lib/components/ui/tabs/index.js';
	import GeneralInformation from '#lib/components/tabs/general-information.svelte';
	import ProfilePanel from '#lib/components/tabs/profile.svelte';
	import Stocks from '#lib/components/tabs/stocks.svelte';
	import Bonds from '#lib/components/tabs/bonds.svelte';
	import { demoProfiles, demoPortfolios, emptyPortfolio } from '#lib/logic/tabs/demo-portfolio.js';
	import { createDemoBondPortfolios, createEmptyBondPortfolio } from '#lib/logic/tabs/demo-bonds.js';
	import type { BondPortfolioSnapshot } from '#lib/logic/tabs/bonds.js';
	import type { PortfolioSnapshot, Profile } from '#lib/logic/tabs/portfolio.js';

	let activeTab = $state('general');
	let profiles = $state<Profile[]>(demoProfiles.map((profile) => ({ ...profile })));
	let selectedId = $state('main');
	let portfolios = $state<Record<string, PortfolioSnapshot>>(structuredClone(demoPortfolios));
	let stocksBusy = $state(false);
	let bondsBusy = $state(false);
	let bondPortfolios = $state<Record<string, BondPortfolioSnapshot>>(createDemoBondPortfolios());
	const emptyBonds = createEmptyBondPortfolio();
	const selectedProfile = $derived(profiles.find((profile) => profile.id === selectedId));
	const portfolio = $derived(portfolios[selectedId] ?? emptyPortfolio);
	const bonds = $derived(bondPortfolios[selectedId] ?? emptyBonds);
	const tabClass =
		'group-data-vertical/tabs:h-auto group-data-vertical/tabs:rounded-none group-data-vertical/tabs:px-4 group-data-vertical/tabs:py-5 whitespace-normal text-left leading-relaxed text-sidebar-foreground/70 hover:bg-sidebar-accent data-active:bg-sidebar-accent data-active:text-primary after:bg-primary group-data-vertical/tabs:after:right-0 group-data-vertical/tabs:after:w-1 sm:group-data-vertical/tabs:px-6';
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
				<p class="mt-1 text-xs text-muted-foreground">Личный кабинет</p>
			</div>
			<Tabs.List aria-label="Разделы портфеля" variant="line" class="w-full gap-0 p-0 pt-4">
				<Tabs.Trigger value="profile" class={tabClass} disabled={stocksBusy || bondsBusy}>Профиль</Tabs.Trigger>
				<Tabs.Trigger value="general" class={tabClass}>Общая информация</Tabs.Trigger>
				<Tabs.Trigger value="stocks" class={tabClass}>Акции</Tabs.Trigger>
				<Tabs.Trigger value="bonds" class={tabClass}>Облигации</Tabs.Trigger>
			</Tabs.List>
			{#if stocksBusy || bondsBusy}<p class="px-4 py-4 text-xs leading-relaxed text-muted-foreground sm:px-6">Завершите редактирование активов, чтобы сменить профиль.</p>{/if}
		</aside>

		<div class="min-w-0 flex-1 p-4 pt-8 sm:p-8 lg:p-12">
			<Tabs.Content value="profile" class="min-h-full">
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
						selectedId = profiles[0].id;
					}}
				/>
			</Tabs.Content>
			<Tabs.Content value="general" class="min-h-full">
				{#key selectedId}<GeneralInformation values={{ ...portfolio.assetValues, bonds: bonds.totalBondsValue }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="bonds" class="min-h-full">
				{#key selectedId}<Bonds portfolio={bonds} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => bondsBusy = busy} oncommit={(snapshot) => bondPortfolios = { ...bondPortfolios, [selectedId]: snapshot }} />{/key}
			</Tabs.Content>
			<Tabs.Content value="stocks" class="min-h-full">
				{#key selectedId}<Stocks {portfolio} profileName={selectedProfile?.name ?? ''} onbusychange={(busy) => stocksBusy = busy} oncommit={(snapshot) => portfolios = { ...portfolios, [selectedId]: snapshot }} />{/key}
			</Tabs.Content>
		</div>
	</Tabs.Root>
</main>
