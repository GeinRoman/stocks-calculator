<script lang="ts">
	import InstrumentSearchDialog from './instrument-search-dialog.svelte';
	import { searchBonds } from '#lib/logic/tabs/bond-service.js';
	import type { BondSearchResult, DraftBondGroup } from '#lib/logic/tabs/bonds.js';
	import { convertMoney, formatMoney, quoteBond } from '#lib/logic/tabs/currency-service.js';
	let { groupId, groups, onadd, onclose, includeAccrued, inRub }: {
		groupId: string; groups: DraftBondGroup[];
		includeAccrued: boolean; inRub: boolean;
		onadd: (bond: BondSearchResult, quantity: number) => void; onclose: () => void;
	} = $props();
	const group = $derived(groups.find(({ id }) => id === groupId));
</script>

<InstrumentSearchDialog groupName={group?.name ?? ''} instrument="облигацию" queryLabel="Название или код облигации" placeholder="Например, ОФЗ" quantityLabel="Количество лотов" search={searchBonds} displayPrice={(bond) => formatMoney(quoteBond(bond, includeAccrued, inRub))} owner={(code) => groups.find((item) => item.id !== groupId && item.bonds.some((bond) => bond.code === code))?.name} alreadyHeld={(code) => !!group?.bonds.some((bond) => bond.code === code)} {onadd} {onclose}>
	{#snippet details(bond)}<p class="text-xs text-muted-foreground">Выплата при погашении: {formatMoney(inRub ? convertMoney(bond.maturityPayment, 'RUB') : bond.maturityPayment)}. Размер лота: {bond.lotSize} облигаций.</p>{/snippet}
</InstrumentSearchDialog>
