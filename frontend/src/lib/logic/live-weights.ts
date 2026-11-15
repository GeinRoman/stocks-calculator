import type { portfolioApi } from '#lib/api/portfolio-api.js';
import type { GroupWeight } from '#lib/logic/tabs/portfolio.js';
import { validateWeights } from '#lib/logic/tabs/stock-editor.js';
import { stockView } from '#lib/api/stock-view.js';
import { bondView } from '#lib/api/bond-view.js';
import type { PortfolioState } from './portfolio-state.js';

export async function saveLiveWeights(
	api: Pick<ReturnType<typeof portfolioApi>, 'stocks' | 'bonds' | 'crypto'>,
	state: PortfolioState,
	kind: 'stock' | 'bond' | 'crypto',
	weights: GroupWeight[]
): Promise<PortfolioState> {
	const groups = kind === 'stock' ? state.data.stock_groups : kind === 'bond' ? state.data.bond_groups : state.data.crypto_groups;
	const invalid = validateWeights(weights, true);
	if (invalid) throw new Error(invalid);
	const byName = new Map(weights.map(weight => [weight.id, weight.value!]));
	if (byName.size !== groups.length || groups.some(group => !byName.has(group.name))) {
		throw new Error('Список групп изменился. Повторите редактирование долей.');
	}
	const updated = groups.map(group => ({ ...group, weight: byName.get(group.name)! }));
	try { await (kind === 'stock' ? api.stocks : kind === 'bond' ? api.bonds : api.crypto).updateWeights(updated); }
	catch { throw new Error('Не удалось сохранить доли. Попробуйте ещё раз.'); }
	const data = kind === 'stock' ? { ...state.data, stock_groups: updated } : kind === 'bond' ? { ...state.data, bond_groups: updated } : { ...state.data, crypto_groups: updated };
	return { ...state, data, stocks: stockView(data), bonds: bondView(data, state.rates, state.loadedAt) };
}
