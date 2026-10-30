import assert from 'node:assert/strict';
import test from 'node:test';
import { groupLayoutModuleUrl } from './group-layout-module.mjs';

const { masonryGroups } = await import(groupLayoutModuleUrl);

test('masonry tracks natural card heights, dynamic children, hidden tabs and observer cleanup', () => {
	const names = ['ResizeObserver', 'MutationObserver', 'requestAnimationFrame', 'cancelAnimationFrame', 'getComputedStyle'];
	const originals = new Map(names.map((name) => [name, globalThis[name]]));
	const frames = new Map();
	const observed = new Set();
	let resize, mutate, frameId = 0, resizeDisconnected = false, mutationDisconnected = false;
	function card(height) {
		const properties = new Map();
		return {
			height, properties,
			style: { setProperty: (key, value) => properties.set(key, value), removeProperty: (key) => properties.delete(key) },
			getBoundingClientRect() { return { height: this.height }; }
		};
	}
	const first = card(100.2), second = card(200.4);
	const node = {
		children: [first, second], dataset: {}, width: 1000,
		getBoundingClientRect() { return { width: this.width }; }
	};
	const flush = () => { const callbacks = [...frames.values()]; frames.clear(); callbacks.forEach((callback) => callback()); };
	try {
		globalThis.ResizeObserver = class {
			constructor(callback) { resize = callback; }
			observe(child) { observed.add(child); }
			unobserve(child) { observed.delete(child); }
			disconnect() { resizeDisconnected = true; }
		};
		globalThis.MutationObserver = class {
			constructor(callback) { mutate = callback; }
			observe(target, options) { assert.equal(target, node); assert.deepEqual(options, { childList: true }); }
			disconnect() { mutationDisconnected = true; }
		};
		globalThis.requestAnimationFrame = (callback) => { frames.set(++frameId, callback); return frameId; };
		globalThis.cancelAnimationFrame = (id) => frames.delete(id);
		globalThis.getComputedStyle = () => ({ columnGap: '24px' });

		const action = masonryGroups(node);
		assert.equal(first.properties.get('--group-rows'), '125');
		assert.equal(second.properties.get('--group-rows'), '225');
		assert.equal(node.dataset.masonryReady, '');
		first.height = 180;
		resize(); resize();
		assert.equal(frames.size, 1, 'Resize notifications share one frame');
		flush();
		assert.equal(first.properties.get('--group-rows'), '204');
		assert.equal(second.properties.get('--group-rows'), '225', 'Other card height is not stretched');

		const third = card(60);
		node.children.push(third); mutate();
		assert.equal(third.properties.get('--group-rows'), '84');
		assert.ok(observed.has(third));
		node.children.shift(); mutate();
		assert.ok(!observed.has(first));
		assert.ok(!first.properties.has('--group-rows'));
		assert.deepEqual(node.children, [second, third], 'DOM order stays unchanged');
		node.width = 0; resize(); flush();
		assert.equal(node.dataset.masonryReady, undefined, 'Hidden tabs fall back to normal flow');
		node.width = 800; resize(); flush();
		assert.equal(node.dataset.masonryReady, '');
		resize();
		action.destroy();
		assert.equal(frames.size, 0);
		assert.ok(resizeDisconnected && mutationDisconnected);
		assert.equal(node.dataset.masonryReady, undefined);
		assert.ok(!second.properties.has('--group-rows') && !third.properties.has('--group-rows'));
	} finally {
		for (const [name, value] of originals) {
			if (value === undefined) delete globalThis[name];
			else globalThis[name] = value;
		}
	}
});
