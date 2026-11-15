<script lang="ts">
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	let { action, onconfirm, onclose }: { action: 'apply' | 'discard'; onconfirm: () => void; onclose: () => void } = $props();
</script>

<AlertDialog.Root open={true} onOpenChange={(open) => { if (!open) onclose(); }}>
	<AlertDialog.Content><AlertDialog.Header><AlertDialog.Title>{action === 'apply' ? 'Применить изменения?' : 'Отменить изменения?'}</AlertDialog.Title><AlertDialog.Description>{action === 'apply' ? 'Изменения будут применены к демонстрационному портфелю. Брокерские операции не выполняются.' : 'Несохранённые изменения будут потеряны.'}</AlertDialog.Description></AlertDialog.Header>
		<AlertDialog.Footer><AlertDialog.Cancel>Вернуться</AlertDialog.Cancel><AlertDialog.Action variant={action === 'discard' ? 'destructive' : 'default'} onclick={onconfirm}>{action === 'apply' ? 'Применить' : 'Отменить изменения'}</AlertDialog.Action></AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
