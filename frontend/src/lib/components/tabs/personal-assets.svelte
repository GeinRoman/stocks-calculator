<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import PersonalAssetDialog from './personal-asset-dialog.svelte';
	import PersonalAssetCloseDialog from './personal-asset-close-dialog.svelte';
	import { estimatedMaturityAmount, getPersonalAssetStatus, localCalendarDate, type PersonalAsset, type PersonalAssetInput, type PersonalAssetsSnapshot, type PersonalAssetType } from '#lib/logic/tabs/personal-assets.js';
	import { createPersonalAsset, updatePersonalAsset, deletePersonalAsset, closePersonalAsset } from '#lib/logic/tabs/personal-assets-service.js';

	let { snapshot, profileName, oncommit, onbusychange }: { snapshot: PersonalAssetsSnapshot; profileName: string; oncommit: (updated: PersonalAssetsSnapshot) => void; onbusychange: (busy: boolean) => void } = $props();
	let dialog = $state<{ kind: 'edit'; type: PersonalAssetType; asset: PersonalAsset | null } | { kind: 'close' | 'delete'; type: PersonalAssetType; asset: PersonalAsset } | null>(null);
	let saving = $state(false);
	let error = $state('');
	let today = $state('');
	$effect(() => onbusychange(dialog !== null || saving));
	onMount(() => {
		const refresh = () => today = localCalendarDate(new Date());
		refresh();
		window.addEventListener('focus', refresh);
		const timer = window.setInterval(refresh, 60_000);
		return () => { window.removeEventListener('focus', refresh); window.clearInterval(timer); };
	});
	const money = (value: number) => new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value);
	const date = (value: string) => { const parts = value.split('-'); return parts.length === 3 ? `${parts[2]}.${parts[1]}.${parts[0]}` : value; };
	async function save(input: PersonalAssetInput) {
		if (!dialog || dialog.kind !== 'edit' || saving) return;
		saving = true;
		try {
			const { type, asset } = dialog;
			const { amountRub, startDate, ...update } = input;
			const updated = asset ? await updatePersonalAsset(snapshot, type, asset.id, update) : await createPersonalAsset(snapshot, type, input);
			oncommit(updated); dialog = null;
		} finally { saving = false; }
	}
	async function close(amount: number) {
		if (!dialog || dialog.kind !== 'close' || saving) return;
		saving = true;
		try { const updated = await closePersonalAsset(snapshot, dialog.type, dialog.asset.id, amount); oncommit(updated); dialog = null; }
		finally { saving = false; }
	}
	async function remove() {
		if (!dialog || dialog.kind !== 'delete' || saving) return;
		saving = true; error = '';
		try { const updated = await deletePersonalAsset(snapshot, dialog.type, dialog.asset.id); oncommit(updated); dialog = null; }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось удалить запись. Попробуйте ещё раз.'; }
		finally { saving = false; }
	}
</script>

<section class="mx-auto max-w-6xl space-y-6" aria-label="Вклады и страхование">
	<header><h1 class="text-xl font-semibold sm:text-2xl">Вклады и страхование</h1><p class="mt-2 break-words text-sm text-muted-foreground">{profileName}</p>
		<!-- TEMP -->
		<p class="mt-2 text-xs text-muted-foreground">Демонстрационные данные: изменения сохраняются только в памяти и пропадают при перезагрузке. Вывод средств пока не записывается.</p>
	</header>
	{#each [{ type: 'deposit' as const, title: 'Банковские вклады', items: snapshot.deposits }, { type: 'insurance' as const, title: 'Страхование', items: snapshot.insurances }] as group (group.type)}
		<section class="space-y-4 border-t border-border pt-6" aria-label={group.title}>
			<div class="flex flex-wrap items-center gap-4"><h2 class="text-lg font-semibold">{group.title}</h2><Button size="sm" onclick={() => { error = ''; dialog = { kind: 'edit', type: group.type, asset: null }; }}>Добавить</Button></div>
			<div class="space-y-3">
				{#each group.items as asset (`${group.type}-${asset.id}`)}
					{@const status = today ? getPersonalAssetStatus(asset.endDate, today) : 'normal'}
					{@const maturity = estimatedMaturityAmount(asset)}
					<article class="min-w-0 rounded-xl border border-border p-4 {status === 'ending' ? 'bg-destructive/10' : status === 'expired' ? 'bg-muted' : 'bg-card'}">
						<div class="flex flex-wrap items-start justify-between gap-4">
							<div class="min-w-0 space-y-2 break-words">
								<h3 class="font-semibold">{asset.name}</h3>
								<p class="text-sm">{group.type === 'deposit' ? 'Банк' : 'Страховая компания'}: {asset.provider}</p>
								<p class="text-sm">Первоначальная сумма: {money(asset.amountRub)} · Ставка: {new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 10 }).format(asset.annualInterest)}% годовых</p>
								<p class="text-sm">Начало: {date(asset.startDate)} · Окончание: {date(asset.endDate)}</p>
								<p class="text-sm">Оценка к окончанию: {maturity === null ? 'недоступна' : money(maturity)}</p>
								{#if status === 'ending'}<p class="text-sm font-medium text-destructive">Окончание в течение 14 дней</p>{:else if status === 'expired'}<p class="text-sm font-medium text-muted-foreground">Срок истёк</p>{/if}
							</div>
							<div class="flex flex-wrap gap-2"><Button size="sm" variant="outline" onclick={() => dialog = { kind: 'close', type: group.type, asset }}>Закрыть</Button><Button size="sm" variant="outline" onclick={() => dialog = { kind: 'edit', type: group.type, asset }}>Изменить</Button><Button size="sm" variant="destructive" onclick={() => { error = ''; dialog = { kind: 'delete', type: group.type, asset }; }}>Удалить</Button></div>
						</div>
					</article>
				{:else}<p class="text-sm text-muted-foreground">Записей пока нет.</p>{/each}
			</div>
		</section>
	{/each}
	<footer class="space-y-2 border-t border-border pt-4 text-xs text-muted-foreground">
		<p>Оценка к окончанию: простые проценты за фактическое число дней, 365 дней в году, без капитализации. Это не текущая стоимость и она не включается в график.</p>
		<div aria-label="Обозначения сроков" class="flex flex-wrap gap-x-6 gap-y-2">
			<p class="flex items-center gap-2"><span aria-hidden="true" class="size-3 shrink-0 rounded-sm border border-destructive/30 bg-destructive/10"></span>Окончание сегодня или в ближайшие 14 дней</p>
			<p class="flex items-center gap-2"><span aria-hidden="true" class="size-3 shrink-0 rounded-sm border border-border bg-muted"></span>Срок истёк</p>
		</div>
	</footer>
</section>

{#if dialog?.kind === 'edit'}
	<PersonalAssetDialog type={dialog.type} asset={dialog.asset} {saving} onsave={save} onclose={() => dialog = null} />
{:else if dialog?.kind === 'close'}
	<PersonalAssetCloseDialog asset={dialog.asset} {saving} onsave={close} onclose={() => dialog = null} />
{/if}
<AlertDialog.Root open={dialog?.kind === 'delete'} onOpenChange={(open) => { if (!open && !saving) dialog = null; }}>
	{#if dialog?.kind === 'delete'}
		<AlertDialog.Content>
			<AlertDialog.Header><AlertDialog.Title>Удалить ошибочную запись?</AlertDialog.Title><AlertDialog.Description>Запись «{dialog.asset.name}» и её приобретение будут стёрты. Для закрытия с записью вывода средств используйте действие «Закрыть».</AlertDialog.Description></AlertDialog.Header>
			{#if error}<p role="alert" class="text-sm text-destructive">{error}</p>{/if}
			<AlertDialog.Footer><Button variant="outline" disabled={saving} onclick={() => dialog = null}>Отмена</Button><Button variant="destructive" disabled={saving} onclick={remove}>{saving ? 'Удаление…' : 'Удалить запись и приобретение'}</Button></AlertDialog.Footer>
		</AlertDialog.Content>
	{/if}
</AlertDialog.Root>
