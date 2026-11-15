import { getContext, setContext } from 'svelte';
import { goto } from '$app/navigation';
import { page } from '$app/state';
import { isBackPastLayer } from './portfolio-back-state.js';

type Layer = { id: string; kind: 'tab' | 'editor' | 'drawer' | 'rebalance'; onback: () => void; active: boolean };
export type PortfolioBack = { enter: (kind: Layer['kind'], onback: () => void) => () => void; observe: (ids: string[] | undefined) => void; afterSettled: (callback: () => void) => void };
const key = Symbol('portfolio-back');
export function usePortfolioBack(): PortfolioBack { return getContext<PortfolioBack>(key); }

export function providePortfolioBack(): PortfolioBack {
	let layers: Layer[] = [], counter = 0, syncing = false, pendingPop = false;
	let settled: (() => void)[] = [];
	function flushSettled() { if (syncing || pendingPop) return; const callbacks = settled; settled = []; callbacks.forEach(callback => callback()); }
	function afterSettled(callback: () => void) { settled.push(callback); flushSettled(); }
	const ids = () => layers.map(layer => layer.id);
	const same = (a: string[], b: string[]) => a.length === b.length && a.every((value, index) => value === b[index]);
	const navigate = async (replace = false) => goto(window.location.href, { shallow: true, replace, reset: false, state: { ...page.state, portfolioBack: ids() } });
	function enter(kind: Layer['kind'], onback: () => void) {
		const layer: Layer = { id: `${++counter}`, kind, onback, active: true };
		layers.push(layer); syncing = true;
		void navigate().finally(() => { syncing = false; flushSettled(); });
		return () => {
			if (!layer.active) return;
			layer.active = false;
			const index = layers.indexOf(layer);
			if (index < 0) return;
			const top = index === layers.length - 1;
			layers.splice(index, 1);
			if (top && (page.state.portfolioBack ?? []).includes(layer.id)) { pendingPop = true; history.back(); }
			else if (!top) { syncing = true; void navigate(true).finally(() => { syncing = false; flushSettled(); }); }
		};
	}
	function observe(stateIds: string[] | undefined) {
		if (syncing) return;
		const actual = stateIds ?? [];
		if (pendingPop) { if (same(actual, ids())) { pendingPop = false; flushSettled(); } return; }
		const layer = layers.at(-1);
		if (!layer || !isBackPastLayer(ids(), actual)) return;
		// Restore the shallow entry before asking for confirmation, so declining keeps Back usable.
		syncing = true;
		void navigate().then(() => {
			if (!layer.active) return;
			if (layer.kind !== 'drawer' && document.querySelector('[role="dialog"], [role="alertdialog"]')) {
				document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
			} else layer.onback();
		}).finally(() => { syncing = false; flushSettled(); });
	}
	const result = { enter, observe, afterSettled }; setContext(key, result); return result;
}
