import type { Profile, PortfolioSnapshot } from './portfolio.js';
import type { BondPortfolioSnapshot } from './bonds.js';
import type { CryptoSnapshot } from './crypto.js';
import type { PersonalAssetsSnapshot } from './personal-assets.js';
import type { CreditCardsSnapshot } from './credit-cards.js';

// TEMP: This is a frontend-only demo format, not a backend snapshot.
export const DEMO_BACKUP_MAX_BYTES = 32 * 1024 * 1024;
export type DemoBackup = {
	format: 'stocks-calculator-demo-backup'; version: 1; selectedId: string;
	profiles: { profile: Profile; stocks: PortfolioSnapshot; bonds: BondPortfolioSnapshot; crypto: CryptoSnapshot; personal: PersonalAssetsSnapshot; creditCards: CreditCardsSnapshot }[];
};

type Obj = Record<string, unknown>;
const obj = (v: unknown): v is Obj => v !== null && typeof v === 'object' && !Array.isArray(v);
const str = (v: unknown): v is string => typeof v === 'string';
const text = (v: unknown): v is string => str(v) && !!v.trim();
const num = (v: unknown): v is number => typeof v === 'number' && Number.isFinite(v) && v >= 0;
const positive = (v: unknown): v is number => num(v) && v > 0;
const integer = (v: unknown): v is number => num(v) && Number.isSafeInteger(v);
const id = (v: unknown): v is string => text(v) && !['__proto__', 'prototype', 'constructor'].includes(v);
const nullableNum = (v: unknown) => v === null || num(v);
const optionalNum = (v: unknown) => v === undefined || nullableNum(v);
const list = (v: unknown, check: (item: unknown) => boolean): boolean => Array.isArray(v) && v.every(check);
const unique = (values: unknown[]) => new Set(values).size === values.length;
const money = (v: unknown) => obj(v) && num(v.price) && text(v.currency);
const stock = (v: unknown) => obj(v) && text(v.code) && text(v.name) && positive(v.price) && integer(v.lotCount) && v.lotCount > 0 && integer(v.lotSize) && v.lotSize > 0 && num(v.totalValue) && optionalNum(v.targetPercent) && optionalNum(v.actualPercent);
const stockGroup = (v: unknown) => obj(v) && id(v.id) && text(v.name) && num(v.totalValue) && optionalNum(v.targetPercent) && optionalNum(v.actualPercent) && list(v.stocks, stock) && unique((v.stocks as Obj[]).map(s => s.code));
const stocks = (v: unknown) => obj(v) && num(v.totalStocksValue) && obj(v.assetValues) && num(v.assetValues.stocks) && num(v.assetValues.bonds) && num(v.assetValues.deposits) && list(v.groups, stockGroup) && unique((v.groups as Obj[]).map(g => g.id)) && unique((v.groups as Obj[]).flatMap(g => (g.stocks as Obj[]).map(s => s.code)));
const bond = (v: unknown) => obj(v) && text(v.code) && text(v.name) && str(v.isin) && money(v.cleanPrice) && money(v.accruedInterest) && (v.maturityDate === null || str(v.maturityDate)) && (v.maturityPayment === null || money(v.maturityPayment)) && integer(v.lotSize) && v.lotSize > 0 && integer(v.lotCount) && v.lotCount > 0 && nullableNum(v.totalValueRub) && nullableNum(v.targetPercent) && nullableNum(v.actualPercent);
const bondGroup = (v: unknown) => obj(v) && id(v.id) && text(v.name) && nullableNum(v.totalValueRub) && nullableNum(v.targetPercent) && nullableNum(v.actualPercent) && list(v.bonds, bond) && unique((v.bonds as Obj[]).map(b => b.code));
const bonds = (v: unknown) => obj(v) && nullableNum(v.totalBondsValueRub) && str(v.asOf) && Number.isFinite(Date.parse(v.asOf)) && list(v.groups, bondGroup) && unique((v.groups as Obj[]).map(g => g.id)) && unique((v.groups as Obj[]).flatMap(g => (g.bonds as Obj[]).map(b => b.code)));
const safeUrl = (v: unknown) => { if (!str(v)) return false; if (!v) return true; try { return ['http:', 'https:'].includes(new URL(v).protocol); } catch { return false; } };
const cryptoAsset = (v: unknown) => obj(v) && id(v.id) && text(v.name) && text(v.platform) && safeUrl(v.url) && num(v.annualInterest) && positive(v.amount) && nullableNum(v.price);
const cryptoGroup = (v: unknown) => obj(v) && id(v.id) && text(v.name) && integer(v.weight) && v.weight <= 100 && list(v.assets, cryptoAsset) && unique((v.assets as Obj[]).map(a => a.id));
const crypto = (v: unknown) => obj(v) && (v.priceCurrency === undefined || ['USDT', 'ETH', 'BTC'].includes(String(v.priceCurrency))) && list(v.groups, cryptoGroup) && unique((v.groups as Obj[]).map(g => g.id)) && unique((v.groups as Obj[]).flatMap(g => (g.assets as Obj[]).map(a => a.id))) && unique((v.groups as Obj[]).flatMap(g => (g.assets as Obj[]).map(a => a.name)));
const date = (v: unknown) => { if (!str(v) || !/^\d{4}-\d{2}-\d{2}$/.test(v)) return false; const d = new Date(`${v}T00:00:00Z`); return Number.isFinite(d.getTime()) && d.toISOString().slice(0, 10) === v && !v.startsWith('0000'); };
const personalAsset = (v: unknown, type: string) => obj(v) && integer(v.id) && v.id > 0 && v.type === type && text(v.name) && text(v.provider) && positive(v.amountRub) && num(v.annualInterest) && date(v.startDate) && date(v.endDate) && (v.startDate as string) <= (v.endDate as string);
const personal = (v: unknown) => obj(v) && list(v.deposits, a => personalAsset(a, 'deposit')) && list(v.insurances, a => personalAsset(a, 'insurance')) && unique((v.deposits as Obj[]).map(a => a.id)) && unique((v.insurances as Obj[]).map(a => a.id));
const card = (v: unknown) => obj(v) && integer(v.id) && v.id > 0 && text(v.name) && text(v.bankName) && date(v.paymentDate) && num(v.paymentAmountRub) && num(v.overallLimitRub);
const cards = (v: unknown) => obj(v) && list(v.cards, card) && unique((v.cards as Obj[]).map(c => c.id));
const entry = (v: unknown) => obj(v) && obj(v.profile) && id(v.profile.id) && text(v.profile.name) && stocks(v.stocks) && bonds(v.bonds) && crypto(v.crypto) && personal(v.personal) && cards(v.creditCards);

export function parseDemoBackup(source: string): DemoBackup {
	let data: unknown;
	try { data = JSON.parse(source); } catch { throw new Error('Файл не является корректным JSON.'); }
	if (!obj(data) || data.format !== 'stocks-calculator-demo-backup' || data.version !== 1 || !id(data.selectedId) || !Array.isArray(data.profiles) || !data.profiles.length || !data.profiles.every(entry)) throw new Error('Неверный или неподдерживаемый формат демонстрационной копии.');
	const entries = data.profiles as DemoBackup['profiles'];
	if (!unique(entries.map(e => e.profile.id)) || !unique(entries.map(e => e.profile.name.toLocaleLowerCase('ru-RU'))) || !entries.some(e => e.profile.id === data.selectedId)) throw new Error('Некорректный список профилей в копии.');
	return data as DemoBackup;
}
