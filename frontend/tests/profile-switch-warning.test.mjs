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

test('live profile picker keeps the original card grid and confirms switching', async () => {
	const source = await readFile(new URL('../src/lib/components/tabs/live-profile.svelte', import.meta.url), 'utf8');
	const list = source.split("{#if mode === 'list'}").at(-1).split("{:else if mode === 'create'}")[0];
	assert.match(list, /grid gap-3 sm:grid-cols-2/);
	assert.match(list, /aria-pressed=\{profile\.default\}/);
	assert.match(list, /if \(profile\.default\) open = false; else show\('switch', profile\)/);
	assert.match(list, /<CheckIcon/);
	assert.match(list, /<PlusIcon[^>]*\/>Добавить профиль/);
	assert.match(source, /обновите страницу на других активных устройствах и во вкладках/);
	assert.match(source, /await onselect\(profile\)/);
});
