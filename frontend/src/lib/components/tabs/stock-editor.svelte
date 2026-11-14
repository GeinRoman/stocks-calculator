<script lang="ts">
	import XIcon from '@lucide/svelte/icons/x';
	import { untrack } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Table from '#lib/components/ui/table/index.js';
	import GroupEditor from './group-editor.svelte';
	import AddStockDialog from './add-stock-dialog.svelte';
	import { addStockToGroup, validateDraft } from '#lib/logic/tabs/stock-editor.js';
	import type { DraftGroup } from '#lib/logic/tabs/portfolio.js';
	import { draftLotChange } from '#lib/logic/tabs/draft-lot-change.js';

	let { groups = $bindable(), profileName, saving, onsave, oncancel }: {
		groups: DraftGroup[]; profileName: string; saving: boolean;
		onsave: (groups: DraftGroup[]) => Promise<void>; oncancel: () => void;
	} = $props();
	let addingTo = $state<string | null>(null);
	const originalLots = untrack(() => new Map(groups.flatMap(group => group.stocks.map(stock => [`${group.id}\u0000${stock.code}`, stock.lotCount] as const))));
</script>

<GroupEditor bind:groups {profileName} {saving} {onsave} {oncancel} instrumentPlural="акции" validate={validateDraft} instruments={(group) => group.stocks} prepare={(items, removed) => items.filter(group => !removed.includes(group.id)).map(group => ({ ...group, stocks: group.stocks.filter(stock => stock.lotCount !== 0) }))} createGroup={(name) => ({ id: crypto.randomUUID(), name, targetPercent: null, stocks: [] })}>
	{#snippet children(group)}
		<Table.Root class="min-w-max [&_td]:text-center [&_th]:text-center">
			<Table.Caption class="sr-only">Акции группы {group.name}. Таблицу можно прокручивать по горизонтали.</Table.Caption>
			<Table.Header><Table.Row><Table.Head>Название</Table.Head><Table.Head>Код</Table.Head><Table.Head>Размер лота</Table.Head><Table.Head>Лотов</Table.Head><Table.Head><span class="sr-only">Действия</span></Table.Head></Table.Row></Table.Header>
			<Table.Body>
				{#each group.stocks as stock (stock.code)}
					<Table.Row class="odd:bg-card even:bg-muted">
						<Table.Cell class="max-w-64 whitespace-normal font-medium">{stock.name}</Table.Cell>
						<Table.Cell>{stock.code}</Table.Cell><Table.Cell>{stock.lotSize}</Table.Cell>
						<Table.Cell><Input aria-label={`Лотов ${stock.code}`} type="number" min={0} max={Number.MAX_SAFE_INTEGER} step={1} bind:value={stock.lotCount} aria-invalid={!Number.isSafeInteger(stock.lotCount) || stock.lotCount === undefined || stock.lotCount < 0} class="mx-auto w-28 text-center [&::-webkit-inner-spin-button]:opacity-100" /></Table.Cell>
						<Table.Cell><Button variant="destructive" size="icon-sm" class="rounded-full" disabled={stock.lotCount === 0} aria-label={`Удалить акцию ${stock.code}`} title={`Удалить акцию ${stock.code}`} onclick={() => stock.lotCount = 0}><XIcon aria-hidden="true" /></Button></Table.Cell>
					</Table.Row>
					{@const change = draftLotChange(originalLots.get(`${group.id}\u0000${stock.code}`), stock.lotCount)}
					{#if change}<Table.Row><Table.Cell colspan={5} class="whitespace-normal bg-muted text-left text-xs text-muted-foreground">{change}</Table.Cell></Table.Row>{/if}
				{:else}<Table.Row><Table.Cell colspan={5} class="text-muted-foreground">В группе пока нет акций.</Table.Cell></Table.Row>{/each}
			</Table.Body>
		</Table.Root>
		<div class="border-t border-border p-4"><Button variant="outline" size="sm" class="h-auto min-h-8 max-w-full whitespace-normal py-2" aria-label={`Добавить акцию в группу ${group.name}`} onclick={() => addingTo = group.id}>Добавить акцию</Button></div>
	{/snippet}
</GroupEditor>

{#if addingTo}
	<AddStockDialog groupId={addingTo} {groups} onclose={() => addingTo = null} onadd={(stock, lots) => groups = addStockToGroup(groups, addingTo!, stock, lots, true)} />
{/if}
