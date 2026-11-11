<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as RadioGroup from '#lib/components/ui/radio-group/index.js';
	import BondGroups from './bond-groups.svelte';
	import BondEditor from './bond-editor.svelte';
	import GroupWeightsDialog from './group-weights-dialog.svelte';
	import { createBondDraftGroups } from '#lib/logic/tabs/bond-editor.js';
	import { saveBondComposition, saveBondWeights } from '#lib/logic/tabs/bond-service.js';
	import type { BondPortfolioSnapshot, DraftBondGroup } from '#lib/logic/tabs/bonds.js';
	import type { GroupWeight } from '#lib/logic/tabs/portfolio.js';
	let { portfolio, profileName, oncommit, onbusychange }: {
		portfolio: BondPortfolioSnapshot; profileName: string;
		oncommit: (snapshot: BondPortfolioSnapshot) => void; onbusychange: (busy: boolean) => void;
	} = $props();
	let editing = $state(false);
	let groups = $state<DraftBondGroup[]>([]);
	let weightsOpen = $state(false);
	let saving = $state(false);
	let notice = $state('');
	let includeAccrued = $state(true);
	let inRub = $state(false);
	$effect(() => onbusychange(editing || weightsOpen || saving));
	async function commitComposition(draft: DraftBondGroup[]) {
		saving = true;
		try {
			const updated = await saveBondComposition(portfolio, draft);
			oncommit(updated); editing = false; notice = 'Состав облигаций обновлён.';
		} finally { saving = false; }
	}
	async function commitWeights(weights: GroupWeight[]) {
		saving = true;
		try {
			const updated = await saveBondWeights(portfolio, weights);
			oncommit(updated); weightsOpen = false; notice = 'Целевые доли облигаций сохранены.';
		} finally { saving = false; }
	}
</script>

{#if editing}
	<BondEditor bind:groups {profileName} {saving} {includeAccrued} {inRub} onsave={commitComposition} oncancel={() => { editing = false; groups = []; }} />
{:else}
	<BondGroups {portfolio} {profileName} bind:includeAccrued bind:inRub>
		{#snippet controls()}
			<div class="mb-6 flex flex-wrap items-center gap-4">
				<div class="flex flex-wrap gap-3"><Button class="h-auto min-h-9 max-w-full whitespace-normal py-2" onclick={() => { groups = createBondDraftGroups(portfolio); editing = true; notice = ''; }}>Редактировать состав</Button><Button variant="outline" class="h-auto min-h-9 max-w-full whitespace-normal py-2" disabled={!portfolio.groups.length} onclick={() => { weightsOpen = true; notice = ''; }}>Настроить доли</Button></div>
				<div class="ml-auto flex min-w-0 flex-wrap items-center justify-end gap-x-6 gap-y-3">
					<RadioGroup.Root aria-label="Цена облигации" value={includeAccrued ? 'dirty' : 'clean'} onValueChange={(value) => includeAccrued = value === 'dirty'} class="flex w-auto flex-col items-start gap-2">
						<div class="flex items-center gap-2"><RadioGroup.Item id="bond-price-clean" value="clean" /><Label for="bond-price-clean">Чистая цена</Label></div>
						<div class="flex items-center gap-2"><RadioGroup.Item id="bond-price-dirty" value="dirty" /><Label for="bond-price-dirty">С НКД</Label></div>
					</RadioGroup.Root>
					<RadioGroup.Root aria-label="Валюта цены облигации" value={inRub ? 'rub' : 'native'} onValueChange={(value) => inRub = value === 'rub'} class="flex w-auto flex-col items-start gap-2">
						<div class="flex items-center gap-2"><RadioGroup.Item id="bond-currency-native" value="native" /><Label for="bond-currency-native">Исходная валюта</Label></div>
						<div class="flex items-center gap-2"><RadioGroup.Item id="bond-currency-rub" value="rub" /><Label for="bond-currency-rub">В рублях</Label></div>
					</RadioGroup.Root>
				</div>
			</div>
			<p role="status" class="mb-4 text-sm text-primary">{notice}</p>
		{/snippet}
	</BondGroups>
{/if}
{#if weightsOpen}
	<GroupWeightsDialog groups={portfolio.groups} {saving} onsave={commitWeights} onclose={() => weightsOpen = false} />
{/if}
