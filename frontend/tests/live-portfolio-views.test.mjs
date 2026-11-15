import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import { compile } from 'svelte/compiler';
import { render } from 'svelte/server';

const dataUrl = code => `data:text/javascript;base64,${Buffer.from(code).toString('base64')}`;
const source = await readFile(new URL('../src/lib/components/tabs/live-portfolio-views.svelte', import.meta.url), 'utf8');
const tabs = dataUrl('export function Root(r,p){p.children?.(r)} export const List=Root, Trigger=Root, Content=Root;');
const stocks = dataUrl('export default function StockGroups() {}');
const code = compile(source, { generate: 'server', filename: 'live-portfolio-views.svelte' }).js.code
	.replace(/(['"])svelte\/internal\/server\1/g, JSON.stringify(import.meta.resolve('svelte/internal/server')))
	.replace(/(['"])#lib\/components\/ui\/tabs\/index\.js\1/g, JSON.stringify(tabs))
	.replace(/(['"])\.\/stock-groups\.svelte\1/g, JSON.stringify(stocks));
const { default: Views } = await import(dataUrl(code));

test('live views display bond, crypto, personal asset and card data without demo rates', () => {
	const portfolio = {
		stocks: { total: 0, groups: [] },
		bonds: { totalBondsValueRub: null, groups: [{ id: 'Debt', name: 'Debt', targetPercent: 100, actualPercent: null, totalValueRub: null, bonds: [{ code: 'B', name: 'Bond name', isin: '', cleanPrice: { currency: 'USD', price: 10 }, accruedInterest: { currency: 'USD', price: 1 }, lotCount: 2, lotSize: 5, totalValueRub: null, targetPercent: 100, actualPercent: null, maturityDate: null, maturityPayment: null }] }] },
		unavailableRates: ['USD'], loadedAt: '2026-01-01T00:00:00Z',
		data: { profiles: [{ name: 'Main', default: true }], crypto_groups: [{ name: 'Crypto', weight: 100 }], crypto_assets: [{ id: 1, name: 'Coin', group_name: 'Crypto', platform: 'Wallet', amount: 2, annual_interest: 0, url: '' }], personal_assets: { deposits: [{ id: 1, name: 'Deposit', bank_name: 'Bank', amount_rub: 10, annual_interest: 0, start_date: '2026-01-01', end_date: '2027-01-01' }], insurances: [{ id: 2, name: 'Insurance', company_name: 'Company', amount_rub: 20, annual_interest: 0, start_date: '2026-01-01', end_date: '2027-01-01' }], credit_cards: [{ id: 3, name: 'Card', bank_name: 'Bank', overall_limit_rub: 100, mandatory_payment_amount_rub: 10, mandatory_payment_date: '2026-12-01' }] } }
	};
	const { body } = render(Views, { props: { portfolio } });
	for (const text of ['Bond name', 'Coin', 'Deposit', 'Insurance', 'Card', 'Нет курса к рублю', 'Рыночные цены не предоставлены']) assert.ok(body.includes(text), text);
	assert.doesNotMatch(body, /Демонстрационные данные|Демонстрационные курсы/);
	assert.match(body, /Количество облигаций<\/dt><dd[^>]*>10/);
});
