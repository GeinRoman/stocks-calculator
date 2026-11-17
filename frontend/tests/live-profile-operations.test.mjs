import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const page = await readFile(new URL('../src/routes/+page.svelte', import.meta.url), 'utf8');
const profile = await readFile(new URL('../src/lib/components/tabs/live-profile.svelte', import.meta.url), 'utf8');

test('successful snapshot installation reloads the same portfolio state as startup', () => {
	const install = page.slice(page.indexOf('async function installSnapshot'), page.indexOf('async function loadPortfolio'));
	assert.match(install, /await portfolioApi\(client\)\.installSnapshot\(snapshot\);[\s\S]*await loadPortfolio\(\)/);
	assert.match(page, /loadPortfolioState\(portfolioApi\(client\)\)/);
	assert.match(install, /portfolioStatus !== 'error'/);
});

test('profile controls validate rename and require confirmation before installing a snapshot', () => {
	assert.match(profile, /trimmed === selected\.name/);
	assert.match(profile, /await onrename\(selected, trimmed\)/);
	assert.match(profile, /32 \* 1024 \* 1024/);
	assert.match(profile, /Все текущие профили, их данные и история операций будут заменены/);
	assert.match(profile, /onclick=\{restoreSnapshot\}/);
	assert.match(profile, /Копия применена, но данные не удалось обновить/);
});
