// TEMP: Crypto snapshots and operations simulate the backend in memory.
export type CryptoAsset = { id: string; name: string; platform: string; url: string; annualInterest: number; amount: number; price: number | null };
export type CryptoGroup = { id: string; name: string; weight: number; assets: CryptoAsset[] };
export type CryptoSnapshot = { groups: CryptoGroup[]; priceCurrency?: 'USDT' | 'ETH' | 'BTC' };
export type CryptoMetadata = Pick<CryptoAsset, 'name' | 'platform' | 'url' | 'annualInterest'>;
export type CryptoCommand =
	| { kind: 'add-group'; name: string }
	| { kind: 'delete-group'; name: string }
	| { kind: 'rename-group'; oldName: string; name: string }
	| { kind: 'edit-asset'; asset: CryptoAsset }
	| { kind: 'buy' | 'sell'; asset: CryptoAsset; groupName: string; amount: number; price: number; currency: string };

export function priceInput(value: string): string {
	let dot = false;
	return Array.from(value).filter((char) => {
		if (char >= '0' && char <= '9') return true;
		if (char === '.' && !dot) { dot = true; return true; }
		return false;
	}).join('');
}

export function priceInputValue(value: number): string {
	const text = String(value);
	if (!/[eE]/.test(text)) return text;
	const [coefficient, exponentText] = text.toLowerCase().split('e');
	const [whole, fraction = ''] = coefficient.split('.');
	const digits = whole + fraction, point = whole.length + Number(exponentText);
	return point <= 0 ? `0.${'0'.repeat(-point)}${digits}` : point >= digits.length ? digits + '0'.repeat(point - digits.length) : `${digits.slice(0, point)}.${digits.slice(point)}`;
}

export function decimal(value: string): number | null {
	const text = value.trim().replace(',', '.');
	if (!/^(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?$/i.test(text)) return null;
	const result = Number(text);
	return Number.isFinite(result) ? result : null;
}

// Use decimal arithmetic for ordinary fractional holdings rather than accumulating binary rounding errors.
function addAmount(a: number, b: number, subtract: boolean): number {
	const parts = (value: number) => {
		const [mantissa, exponent = '0'] = value.toString().toLowerCase().split('e');
		const [whole, fraction = ''] = mantissa.split('.');
		return { digits: BigInt(whole + fraction), scale: fraction.length - Number(exponent) };
	};
	const x = parts(a), y = parts(b), scale = Math.max(x.scale, y.scale, 0);
	const units = x.digits * 10n ** BigInt(scale - x.scale) + (subtract ? -1n : 1n) * y.digits * 10n ** BigInt(scale - y.scale);
	return Number(`${units}e-${scale}`);
}

export function cryptoValue(assets: CryptoAsset[]): { value: number | null; partial: boolean } {
	const known = assets.filter((asset) => asset.price !== null);
	const partial = known.length < assets.length;
	if (assets.length && !known.length) return { value: null, partial };
	const total = known.reduce((sum, asset) => sum + asset.amount * (asset.price ?? 0), 0);
	return { value: Number.isFinite(total) ? total : null, partial };
}
export function validateMetadata(input: CryptoMetadata): string | null {
	if (!input.name.trim() || !input.platform.trim()) return 'Укажите название и платформу.';
	if (!Number.isFinite(input.annualInterest) || input.annualInterest < 0) return 'Укажите неотрицательную годовую ставку.';
	if (input.url) { try { if (!['https:', 'http:'].includes(new URL(input.url).protocol)) return 'Укажите ссылку HTTP или HTTPS.'; } catch { return 'Укажите корректную ссылку.'; } }
	return null;
}
export function transactAmount(balance: number, amount: number, sell: boolean): number {
	if (!Number.isFinite(amount) || amount <= 0 || (sell && amount > balance)) throw new Error('Укажите положительное количество, не превышающее доступный остаток.');
	const result = addAmount(balance, amount, sell);
	if (!Number.isFinite(result) || (result === balance && amount !== 0)) throw new Error('Количество слишком мало или слишком велико для отображения.');
	return result;
}

export function demoCrypto(): Record<string, CryptoSnapshot> {
	return { main: { groups: [
		{ id: 'coins', name: 'Основные монеты', weight: 70, assets: [
			{ id: 'btc', name: 'Bitcoin', platform: 'Биржа', url: 'https://bitcoin.org/', annualInterest: 0, amount: 0.125, price: null },
			{ id: 'eth', name: 'Ethereum', platform: 'Кошелёк', url: '', annualInterest: 3.5, amount: 1.75, price: null }
		] },
		{ id: 'other', name: 'Другие активы', weight: 30, assets: [{ id: 'sol', name: 'Solana', platform: 'Биржа', url: '', annualInterest: 5, amount: 12.5, price: null }] }
	] } };
}
