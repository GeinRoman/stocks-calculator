<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import StockGroups from './stock-groups.svelte';
	import StockEditor from './stock-editor.svelte';
	import GroupWeightsDialog from './group-weights-dialog.svelte';
	import { createDraftGroups } from '#lib/logic/tabs/stock-editor.js';
	import { saveComposition, saveWeights } from '#lib/logic/tabs/portfolio-service.js';
	import type { DraftGroup, GroupWeight, PortfolioSnapshot } from '#lib/logic/tabs/portfolio.js';

	let { portfolio, profileName, oncommit, onbusychange }: {
		portfolio: PortfolioSnapshot;
		profileName: string;
		oncommit: (snapshot: PortfolioSnapshot) => void;
		onbusychange: (busy: boolean) => void;
	} = $props();
	let editing = $state(false);
	let groups = $state<DraftGroup[]>([]);
	let weightsOpen = $state(false);
	let saving = $state(false);
	let notice = $state('');
	$effect(() => onbusychange(editing || weightsOpen || saving));

	async function commitComposition(draft: DraftGroup[]) {
		saving = true;
		try {
			const updated = await saveComposition(portfolio, draft);
			oncommit(updated);
			editing = false;
			notice = 'Состав портфеля обновлён.';
		} finally { saving = false; }
	}

	async function commitWeights(weights: GroupWeight[]) {
		saving = true;
		try {
			const updated = await saveWeights(portfolio, weights);
			oncommit(updated);
			weightsOpen = false;
			notice = 'Целевые доли сохранены.';
		} finally { saving = false; }
	}
</script>

{#if editing}
	<StockEditor bind:groups {profileName} {saving} onsave={commitComposition} oncancel={() => { editing = false; groups = []; }} />
{:else}
	<StockGroups {portfolio} {profileName}>
		{#snippet controls()}
			<div class="mb-6 flex flex-wrap gap-3">
				<Button class="h-auto min-h-9 max-w-full whitespace-normal py-2" onclick={() => { groups = createDraftGroups(portfolio); editing = true; notice = ''; }}>Редактировать состав</Button>
				<Button variant="outline" class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={!portfolio.groups.length} onclick={() => { weightsOpen = true; notice = ''; }}>Настроить доли</Button>
			</div>
			<p role="status" class="mb-4 text-sm text-primary">{notice}</p>
		{/snippet}
	</StockGroups>
{/if}

{#if weightsOpen}
	<GroupWeightsDialog {portfolio} {saving} onsave={commitWeights} onclose={() => weightsOpen = false} />
{/if}
