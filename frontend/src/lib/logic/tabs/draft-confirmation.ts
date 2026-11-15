export function hasDraftChanges(initial: string, current: unknown, pendingOperations = 0): boolean {
	return pendingOperations > 0 || JSON.stringify(current) !== initial;
}

export function hasRebalanceChanges(trades: { quantity: number | undefined }[]): boolean {
	return trades.some(trade => trade.quantity !== 0);
}
