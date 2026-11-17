import type { portfolioApi } from '#lib/api/portfolio-api.js';
import type { CreditCardInput } from '#lib/logic/tabs/credit-cards.js';

type PersonalApi = ReturnType<typeof portfolioApi>['personal'];
export type CreditCardOperation =
	| { kind: 'create'; input: CreditCardInput }
	| { kind: 'update'; id: number; input: CreditCardInput }
	| { kind: 'remove'; id: number };

export async function saveLiveCreditCard(api: PersonalApi, operation: CreditCardOperation): Promise<void> {
	if (operation.kind === 'remove') {
		await api.removeCreditCard(operation.id);
		return;
	}
	const input = operation.input;
	const card = {
		name: input.name.trim(),
		bank_name: input.bankName.trim(),
		mandatory_payment_date: input.paymentDate,
		mandatory_payment_amount_rub: input.paymentAmountRub,
		overall_limit_rub: input.overallLimitRub
	};
	if (operation.kind === 'create') await api.addCreditCard({ ...card, id: 0 });
	else await api.updateCreditCard({ ...card, id: operation.id });
}
