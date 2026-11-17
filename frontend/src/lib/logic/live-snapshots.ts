import type { AllInfo } from '#lib/api/contracts.js';
import type { CryptoSnapshot } from './tabs/crypto.js';
import type { PersonalAssetsSnapshot } from './tabs/personal-assets.js';
import type { CreditCardsSnapshot } from './tabs/credit-cards.js';

/** Map backend records into the existing read-only tab models. */
export function liveSnapshots(data: AllInfo): { crypto: CryptoSnapshot; personal: PersonalAssetsSnapshot; cards: CreditCardsSnapshot } {
	return {
		crypto: { groups: data.crypto_groups.map(group => ({ id: group.name, name: group.name, weight: group.weight, assets: data.crypto_assets.filter(asset => asset.group_name === group.name).map(asset => ({ id: String(asset.id), name: asset.name, platform: asset.platform, url: asset.url, annualInterest: asset.annual_interest, amount: asset.amount, price: null })) })) },
		personal: {
			deposits: data.personal_assets.deposits.map(item => ({ id: item.id, type: 'deposit', name: item.name, provider: item.bank_name, amountRub: item.amount_rub, annualInterest: item.annual_interest, startDate: item.start_date, endDate: item.end_date })),
			insurances: data.personal_assets.insurances.map(item => ({ id: item.id, type: 'insurance', name: item.name, provider: item.company_name, amountRub: item.amount_rub, annualInterest: item.annual_interest, startDate: item.start_date, endDate: item.end_date }))
		},
		cards: { cards: data.personal_assets.credit_cards.map(card => ({ id: card.id, name: card.name, bankName: card.bank_name, paymentDate: card.mandatory_payment_date, paymentAmountRub: card.mandatory_payment_amount_rub, overallLimitRub: card.overall_limit_rub })) }
	};
}
