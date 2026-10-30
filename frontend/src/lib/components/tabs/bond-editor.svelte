<script lang="ts">
	import XIcon from '@lucide/svelte/icons/x';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Table from '#lib/components/ui/table/index.js';
	import GroupEditor from './group-editor.svelte';
	import AddBondDialog from './add-bond-dialog.svelte';
	import { addBondToGroup, validateBondDraft } from '#lib/logic/tabs/bond-editor.js';
	import type { DraftBondGroup } from '#lib/logic/tabs/bonds.js';

	let { groups = $bindable(), profileName, saving, onsave, oncancel }: {
		groups: DraftBondGroup[]; profileName: string; saving: boolean;
		onsave: (groups: DraftBondGroup[]) => Promise<void>; oncancel: () => void;
	} = $props();
	let addingTo = $state<string | null>(null);
</script>

<GroupEditor bind:groups {profileName} {saving} {onsave} {oncancel} instrumentPlural="облигации" validate={validateBondDraft} createGroup={(name) => ({ id: crypto.randomUUID(), name, targetPercent: null, bonds: [] })}>
	{#snippet children(group)}
		<Table.Root class="min-w-max [&_td]:text-center [&_th]:text-center">
			<Table.Caption class="sr-only">Облигации группы {group.name}. Таблицу можно прокручивать по горизонтали.</Table.Caption>
			<Table.Header><Table.Row><Table.Head>Название</Table.Head><Table.Head>Код</Table.Head><Table.Head>Количество</Table.Head><Table.Head><span class="sr-only">Действия</span></Table.Head></Table.Row></Table.Header>
			<Table.Body>
				{#each group.bonds as bond (bond.code)}
					<Table.Row class="odd:bg-card even:bg-muted">
						<Table.Cell class="max-w-48 whitespace-normal font-medium">{bond.name}</Table.Cell>
						<Table.Cell>{bond.code}</Table.Cell>
						<Table.Cell><Input aria-label={`Количество ${bond.code}`} type="number" min={1} max={Number.MAX_SAFE_INTEGER} step={1} bind:value={bond.quantity} aria-invalid={!Number.isSafeInteger(bond.quantity) || !bond.quantity || bond.quantity < 1} class="mx-auto w-28 text-center [&::-webkit-inner-spin-button]:opacity-100" /></Table.Cell>
						<Table.Cell><Button variant="destructive" size="icon-sm" class="rounded-full" aria-label={`Удалить облигацию ${bond.code}`} title={`Удалить облигацию ${bond.code}`} onclick={() => group.bonds = group.bonds.filter(({ code }) => code !== bond.code)}><XIcon aria-hidden="true" /></Button></Table.Cell>
					</Table.Row>
				{:else}<Table.Row><Table.Cell colspan={4} class="text-muted-foreground">В группе пока нет облигаций.</Table.Cell></Table.Row>{/each}
			</Table.Body>
		</Table.Root>
		<div class="border-t border-border p-4"><Button variant="outline" size="sm" class="h-auto min-h-8 max-w-full whitespace-normal py-2" aria-label={`Добавить облигацию в группу ${group.name}`} onclick={() => addingTo = group.id}>Добавить облигацию</Button></div>
	{/snippet}
</GroupEditor>

{#if addingTo}
	<AddBondDialog groupId={addingTo} {groups} onclose={() => addingTo = null} onadd={(bond, quantity) => groups = addBondToGroup(groups, addingTo!, bond, quantity)} />
{/if}
