<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Label } from '#lib/components/ui/label/index.js';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { DEMO_BACKUP_MAX_BYTES, parseDemoBackup, type DemoBackup } from '#lib/logic/tabs/demo-backup.js';
	let { getBackup, onrestore, onbusychange }: { getBackup: () => DemoBackup; onrestore: (backup: DemoBackup) => void; onbusychange: (busy: boolean) => void } = $props();
	let pending = $state<DemoBackup | null>(null), filename = $state(''), loading = $state(false), error = $state(''), notice = $state('');
	$effect(() => onbusychange(loading || pending !== null));
	function download() {
		error = ''; notice = '';
		try {
			const json = JSON.stringify(getBackup(), null, 2);
			parseDemoBackup(json);
			const blob = new Blob([json], { type: 'application/json' });
			if (blob.size > DEMO_BACKUP_MAX_BYTES) throw new Error('Демонстрационная копия превышает 32 МиБ.');
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a'); anchor.href = url; anchor.download = 'demo-portfolio-backup.json'; anchor.click();
			window.setTimeout(() => URL.revokeObjectURL(url), 1000);
			notice = 'Демонстрационная копия подготовлена к скачиванию.';
		} catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось создать копию.'; }
	}
	async function selectFile(event: Event) {
		const element = event.currentTarget as HTMLInputElement;
		const file = element.files?.[0]; element.value = '';
		if (!file) return;
		error = ''; notice = ''; pending = null;
		if (file.size > DEMO_BACKUP_MAX_BYTES) { error = 'Файл превышает 32 МиБ.'; return; }
		loading = true;
		try { pending = parseDemoBackup(await file.text()); filename = file.name; }
		catch (cause) { error = cause instanceof Error ? cause.message : 'Не удалось прочитать файл.'; }
		finally { loading = false; }
	}
	function restore() {
		if (!pending) return;
		try { onrestore(pending); pending = null; notice = 'Демонстрационные данные всех профилей восстановлены в памяти.'; }
		catch { error = 'Не удалось восстановить копию.'; pending = null; }
	}
</script>

<!-- TEMP: Backup and restore operate only on the demo state. -->
<section aria-labelledby="backup-heading" class="mt-6 rounded-xl border border-border bg-card p-4 text-card-foreground sm:p-8">
	<h2 id="backup-heading" class="text-base font-medium">Резервная копия</h2>
	<p class="mt-3 text-sm leading-relaxed text-muted-foreground">Скачайте JSON-копию демонстрационных данных всех профилей: состав портфелей, вклады, страхование и кредитные карты. Храните файл в безопасном месте: он содержит финансовые данные.</p>
	<p class="mt-3 text-sm leading-relaxed text-muted-foreground">Восстановление заменит все текущие демонстрационные профили данными из файла, а не объединит их. Сначала сохраните текущую копию, если она вам нужна. После перезагрузки страницы восстановленные данные исчезнут.</p>
	<p class="mt-3 text-xs leading-relaxed text-muted-foreground">Демонстрационный формат не совместим с резервными копиями сервера; история операций и данные аккаунта в нём отсутствуют. Максимальный размер файла — 32 МиБ.</p>
	<div class="mt-5 flex flex-wrap items-end gap-3">
		<Button type="button" variant="outline" disabled={loading || pending !== null} onclick={download}>Скачать резервную копию</Button>
		<div class="min-w-0 space-y-2"><Label for="demo-backup-file">Восстановить из JSON-файла</Label><input id="demo-backup-file" type="file" accept=".json,application/json" disabled={loading || pending !== null} onchange={selectFile} class="block h-9 w-full min-w-0 max-w-full rounded-3xl border border-input bg-input/50 px-3 py-1 text-sm file:mr-2 file:border-0 file:bg-transparent file:text-foreground focus-visible:outline-2 focus-visible:outline-ring disabled:opacity-50" /></div>
	</div>
	{#if loading}<p role="status" class="mt-3 text-sm text-muted-foreground">Проверка файла…</p>{/if}
	{#if notice}<p role="status" class="mt-3 text-sm text-primary">{notice}</p>{/if}
	{#if error}<p role="alert" class="mt-3 text-sm text-destructive">{error}</p>{/if}
</section>

<AlertDialog.Root open={pending !== null} onOpenChange={(open) => { if (!open) pending = null; }}>
	{#if pending}<AlertDialog.Content><AlertDialog.Header><AlertDialog.Title>Заменить все профили?</AlertDialog.Title><AlertDialog.Description>Файл «{filename}». Профилей в файле: {pending.profiles.length}. Все текущие демонстрационные профили и данные будут заменены. Это действие нельзя отменить.</AlertDialog.Description></AlertDialog.Header><AlertDialog.Footer><AlertDialog.Cancel>Отмена</AlertDialog.Cancel><AlertDialog.Action variant="destructive" onclick={restore}>Заменить все профили</AlertDialog.Action></AlertDialog.Footer></AlertDialog.Content>{/if}
</AlertDialog.Root>
