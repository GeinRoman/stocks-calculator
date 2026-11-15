import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

for (const component of ['stock-groups', 'bond-groups', 'crypto']) {
	test(`${component} card fields have two mobile columns and allow values to wrap`, async () => {
		const source = await readFile(new URL(`../src/lib/components/tabs/${component}.svelte`, import.meta.url), 'utf8');
		const lists = [...source.matchAll(/<dl class="([^"]+)"/g)].map(match => match[1]);
		assert.equal(lists.length, 2);
		for (const classes of lists) {
			assert.match(classes, /(?:^|\s)grid-cols-2(?:\s|$)/);
			assert.ok(classes.includes('[&>div]:min-w-0'));
			assert.ok(classes.includes('[&_dd]:wrap-anywhere'));
		}
		assert.match(lists[1], /(?:^|\s)sm:grid-cols-3(?:\s|$)/);
	});
}
