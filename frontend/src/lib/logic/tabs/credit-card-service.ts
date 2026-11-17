import { emptyCreditCards, validateCreditCard } from './credit-cards.js';
import type { CreditCard, CreditCardInput, CreditCardsSnapshot } from './credit-cards.js';
import { createDemoCreditCards } from './credit-card-demo.js';

// TEMP
function checkedSnapshot(snapshot: CreditCardsSnapshot): CreditCardsSnapshot {
	const ids = new Set<number>();
	for (const card of snapshot.cards) {
		if (!Number.isSafeInteger(card.id) || card.id <= 0 || ids.has(card.id)) throw new Error('Некорректные или повторяющиеся идентификаторы карт.');
		ids.add(card.id);
	}
	return { cards: snapshot.cards.map((card) => ({ ...card })) };
}

// TEMP
function cardFromInput(id: number, input: CreditCardInput): CreditCard {
	const error = validateCreditCard(input);
	if (error) throw new Error(error);
	return { id, name: input.name.trim(), bankName: input.bankName.trim(), paymentDate: input.paymentDate, paymentAmountRub: input.paymentAmountRub, overallLimitRub: input.overallLimitRub };
}

// TEMP
export async function getCreditCards(profileId: string): Promise<CreditCardsSnapshot> {
	return profileId === 'main' ? checkedSnapshot(createDemoCreditCards().main) : emptyCreditCards();
}

// TEMP
export async function createCreditCard(snapshot: CreditCardsSnapshot, input: CreditCardInput): Promise<CreditCardsSnapshot> {
	const result = checkedSnapshot(snapshot);
	const id = result.cards.reduce((max, card) => Math.max(max, card.id), 0) + 1;
	if (!Number.isSafeInteger(id)) throw new Error('Невозможно создать идентификатор карты.');
	result.cards.push(cardFromInput(id, input));
	return result;
}

// TEMP
export async function updateCreditCard(snapshot: CreditCardsSnapshot, id: number, input: CreditCardInput): Promise<CreditCardsSnapshot> {
	const result = checkedSnapshot(snapshot);
	const index = result.cards.findIndex((card) => card.id === id);
	if (index < 0) throw new Error('Карта не найдена.');
	result.cards[index] = cardFromInput(id, input);
	return result;
}

// TEMP
export async function deleteCreditCard(snapshot: CreditCardsSnapshot, id: number): Promise<CreditCardsSnapshot> {
	const result = checkedSnapshot(snapshot);
	const index = result.cards.findIndex((card) => card.id === id);
	if (index < 0) throw new Error('Карта не найдена.');
	result.cards.splice(index, 1);
	return result;
}
