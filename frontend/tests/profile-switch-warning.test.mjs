import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('switching profiles requires confirmation and warns about other devices and tabs', async () => {
	const source = await readFile(new URL('../src/lib/components/tabs/profile.svelte', import.meta.url), 'utf8');
	assert.match(source, /profile\.id === selectedId\) open = false; else \{ pendingProfile = profile; mode = 'confirm-switch'; \}/);
	assert.match(source, /обновите страницу на других активных устройствах и во вкладках/);
	const confirmation = source.split("{:else if mode === 'confirm-switch'}").at(-1).split("{:else if mode === 'delete'}")[0];
	assert.match(confirmation, /show\('switch'\)/);
	assert.match(confirmation, /onselect\(pendingProfile\.id\)/);
});
