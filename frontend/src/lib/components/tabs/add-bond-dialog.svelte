<script lang="ts">
	import InstrumentSearchDialog from './instrument-search-dialog.svelte';
	import { searchBonds } from '#lib/logic/tabs/bond-service.js';
	import type { BondSearchResult, DraftBondGroup } from '#lib/logic/tabs/bonds.js';
	let { groupId, groups, onadd, onclose }: {
		groupId: string; groups: DraftBondGroup[];
		onadd: (bond: BondSearchResult, quantity: number) => void; onclose: () => void;
	} = $props();
	const group = $derived(groups.find(({ id }) => id === groupId));
	const currency = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 2 });
</script>

<InstrumentSearchDialog groupName={group?.name ?? ''} instrument="облигацию" queryLabel="Название или код облигации" placeholder="Например, ОФЗ" quantityLabel="Количество облигаций" search={searchBonds} owner={(code) => groups.find((item) => item.id !== groupId && item.bonds.some((bond) => bond.code === code))?.name} alreadyHeld={(code) => !!group?.bonds.some((bond) => bond.code === code)} {onadd} {onclose}>
	{#snippet details(bond)}<p class="text-xs text-muted-foreground">Цена погашения: {bond.redemptionPrice === null ? 'Нет данных' : currency.format(bond.redemptionPrice)}. Количество указано в отдельных облигациях.</p>{/snippet}
</InstrumentSearchDialog>
