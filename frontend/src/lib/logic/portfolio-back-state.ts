export function isBackPastLayer(expected: string[], actual: string[]): boolean {
	const top = expected.at(-1);
	if (!top) return false;
	return !actual.includes(top);
}
