export function draftLotChange(original: number | undefined, current: number | undefined): string | null {
	if (current === undefined || !Number.isSafeInteger(current) || current < 0) return null;
	if (original === undefined) return `Добавлено: ${current} лот. (черновик)`;
	if (original === current) return null;
	return `Лоты: ${original} → ${current} (черновик)`;
}
