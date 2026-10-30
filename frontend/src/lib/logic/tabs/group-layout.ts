export function masonryGroups(node: HTMLElement) {
	const observed = new Set<HTMLElement>();
	let frame: number | undefined;

	function measure() {
		if (!node.getBoundingClientRect().width) {
			delete node.dataset.masonryReady;
			return;
		}
		const gap = Number.parseFloat(getComputedStyle(node).columnGap) || 0;
		for (const child of observed) {
			// One-pixel grid rows reserve the card's natural height plus its column gap.
			child.style.setProperty('--group-rows', String(Math.ceil(child.getBoundingClientRect().height + gap)));
		}
		node.dataset.masonryReady = '';
	}

	function schedule() {
		if (frame !== undefined) return;
		frame = requestAnimationFrame(() => { frame = undefined; measure(); });
	}

	const resizeObserver = new ResizeObserver(schedule);
	function syncChildren() {
		const children = new Set(Array.from(node.children) as HTMLElement[]);
		for (const child of observed) {
			if (!children.has(child)) {
				resizeObserver.unobserve(child);
				child.style.removeProperty('--group-rows');
				observed.delete(child);
			}
		}
		for (const child of children) {
			if (!observed.has(child)) {
				observed.add(child);
				resizeObserver.observe(child);
			}
		}
		measure();
	}

	const mutationObserver = new MutationObserver(syncChildren);
	syncChildren();
	mutationObserver.observe(node, { childList: true });
	return {
		destroy() {
			resizeObserver.disconnect();
			mutationObserver.disconnect();
			if (frame !== undefined) cancelAnimationFrame(frame);
			for (const child of observed) child.style.removeProperty('--group-rows');
			delete node.dataset.masonryReady;
		}
	};
}
