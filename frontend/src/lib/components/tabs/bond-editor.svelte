<script lang="ts">
	import XIcon from '@lucide/svelte/icons/x';
	import { untrack } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Table from '#lib/components/ui/table/index.js';
	import GroupEditor from './group-editor.svelte';
	import AddBondDialog from './add-bond-dialog.svelte';
	import { addBondToGroup, validateBondDraft } from '#lib/logic/tabs/bond-editor.js';
	import type { DraftBondGroup, BondSearchResult } from '#lib/logic/tabs/bonds.js';
	import { draftLotChange } from '#lib/logic/tabs/draft-lot-change.js';

	let { groups = $bindable(), profileName, saving, onsave, oncancel, includeAccrued, inRub, search, rates, blocked = false }: {
		groups: DraftBondGroup[]; profileName: string; saving: boolean;
		includeAccrued: boolean; inRub: boolean;
		onsave: (groups: DraftBondGroup[]) => Promise<void>; oncancel: () => void;
		search?: (query: string) => Promise<BondSearchResult[]>; rates?: Readonly<Record<string, number>>; blocked?: boolean;
	} = $props();
	let addingTo = $state<string | null>(null);
	const originalLots = untrack(() => new Map(groups.flatMap(group => group.bonds.map(bond => [`${group.id}\u0000${bond.code}`, bond.lotCount] as const))));
</script>

{#if search}<p class="mx-auto mb-3 max-w-6xl text-xs text-muted-foreground">Для записей покупки и продажи используется загруженная цена одной облигации с НКД; для новой позиции — котировка из поиска. Это может отличаться от фактической цены сделки.</p>{/if}
<GroupEditor bind:groups {profileName} {saving} {onsave} {oncancel} {blocked} live={!!search} instrumentPlural="облигации" validate={validateBondDraft} instruments={(group) => group.bonds} prepare={(items, removed) => items.filter(group => !removed.includes(group.id)).map(group => ({ ...group, bonds: group.bonds.filter(bond => bond.lotCount !== 0) }))} createGroup={(name) => ({ id: crypto.randomUUID(), name, targetPercent: search ? 0 : null, bonds: [] })}>
	{#snippet children(group)}
		<Table.Root class="min-w-[25rem] table-fixed [&_td]:px-1 [&_td]:text-center [&_th]:px-1 [&_th]:text-center">
			<Table.Caption class="sr-only">Облигации группы {group.name}. Таблицу можно прокручивать по горизонтали.</Table.Caption>
			<Table.Header><Table.Row><Table.Head class="w-[30%]">Название</Table.Head><Table.Head class="w-[25%]">Код</Table.Head><Table.Head class="w-[15%] whitespace-normal">Размер лота</Table.Head><Table.Head class="w-[19%]">Лотов</Table.Head><Table.Head class="w-[11%]"><span class="sr-only">Действия</span></Table.Head></Table.Row></Table.Header>
			<Table.Body>
				{#each group.bonds as bond (bond.code)}
					<Table.Row class="odd:bg-card even:bg-muted">
						<Table.Cell class="whitespace-normal break-words font-medium">{bond.name}</Table.Cell>
						<Table.Cell class="whitespace-normal break-all">{bond.code}</Table.Cell><Table.Cell>{bond.lotSize}</Table.Cell>
						<Table.Cell><Input aria-label={`Лотов ${bond.code}`} type="number" min={0} max={Number.MAX_SAFE_INTEGER} step={1} bind:value={bond.lotCount} aria-invalid={!Number.isSafeInteger(bond.lotCount) || bond.lotCount === undefined || bond.lotCount < 0} class="mx-auto w-full min-w-0 px-1 text-center [&::-webkit-inner-spin-button]:opacity-100" /></Table.Cell>
						<Table.Cell><Button variant="destructive" size="icon-sm" class="rounded-full" disabled={bond.lotCount === 0} aria-label={`Удалить облигацию ${bond.code}`} title={`Удалить облигацию ${bond.code}`} onclick={() => bond.lotCount = 0}><XIcon aria-hidden="true" /></Button></Table.Cell>
					</Table.Row>
					{@const change = draftLotChange(originalLots.get(`${group.id}\u0000${bond.code}`), bond.lotCount)}
					{#if change}<Table.Row><Table.Cell colspan={5} class="whitespace-normal bg-muted text-left text-xs text-muted-foreground">{change}</Table.Cell></Table.Row>{/if}
				{:else}<Table.Row><Table.Cell colspan={5} class="text-muted-foreground">В группе пока нет облигаций.</Table.Cell></Table.Row>{/each}
			</Table.Body>
		</Table.Root>
		<div class="border-t border-border p-4"><Button variant="outline" size="sm" class="h-auto min-h-8 max-w-full whitespace-normal py-2" aria-label={`Добавить облигацию в группу ${group.name}`} onclick={() => addingTo = group.id}>Добавить облигацию</Button></div>
	{/snippet}
</GroupEditor>

{#if addingTo}
	<AddBondDialog groupId={addingTo} {groups} {includeAccrued} {inRub} {search} {rates} onclose={() => addingTo = null} onadd={(bond, quantity) => groups = addBondToGroup(groups, addingTo!, bond, quantity, true)} />
{/if}
