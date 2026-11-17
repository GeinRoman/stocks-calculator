import type { portfolioApi } from '#lib/api/portfolio-api.js';
import type { PersonalAssetInput, PersonalAssetType } from '#lib/logic/tabs/personal-assets.js';

type PersonalApi = ReturnType<typeof portfolioApi>['personal'];
export type PersonalAssetOperation =
	| { kind: 'create'; type: PersonalAssetType; input: PersonalAssetInput }
	| { kind: 'update'; type: PersonalAssetType; id: number; input: PersonalAssetInput }
	| { kind: 'remove'; type: PersonalAssetType; id: number }
	| { kind: 'close'; type: PersonalAssetType; id: number; amount: number };

export class PersonalAssetSaveError extends Error {
	constructor(message: string, public uncertain: boolean) { super(message); }
}

export async function saveLivePersonalAsset(api: PersonalApi, operation: PersonalAssetOperation): Promise<void> {
	if (operation.kind === 'remove') {
		if (operation.type === 'deposit') await api.removeDeposit(operation.id);
		else await api.removeInsurance(operation.id);
	} else if (operation.kind === 'close') {
		if (operation.type === 'deposit') await api.closeDeposit(operation.id, operation.amount);
		else await api.closeInsurance(operation.id, operation.amount);
	} else {
		const input = operation.input;
		const fields = { name: input.name.trim(), annual_interest: input.annualInterest, start_date: input.startDate, end_date: input.endDate };
		if (operation.type === 'deposit') {
			if (operation.kind === 'create') await api.addDeposit({ ...fields, id: 0, bank_name: input.provider.trim(), amount_rub: input.amountRub });
			else await api.updateDeposit({ ...fields, id: operation.id, bank_name: input.provider.trim() });
		} else {
			if (operation.kind === 'create') await api.addInsurance({ ...fields, id: 0, company_name: input.provider.trim(), amount_rub: input.amountRub });
			else await api.updateInsurance({ ...fields, id: operation.id, company_name: input.provider.trim() });
		}
	}
}
