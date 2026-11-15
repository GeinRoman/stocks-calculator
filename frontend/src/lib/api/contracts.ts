// Wire types mirror internal/model. They are separate from frontend view models.
export interface Profile { name: string; default: boolean }
export interface Group { name: string; weight: number }
export interface RenameGroup { name_old: string; name_new: string }
export interface Money { currency: string; price: number }
export interface Stock { name: string; code: string; group_name: string; price: number; lot_amount: number; lot_size: number }
export interface Bond { name: string; code: string; isin: string; group_name: string; clean_price: Money; accrued_interest: Money; maturity_date: string; maturity_payment: Money | null; lot_amount: number; lot_size: number }
export interface BondOperation { name: string; code: string; group_name: string; lot_amount: number; lot_size: number; price: Money }
export interface CryptoAsset { id: number; name: string; group_name: string; platform: string; url: string; annual_interest: number; amount: number }
export type CryptoAssetUpdate = Pick<CryptoAsset, 'id' | 'name' | 'platform' | 'url' | 'annual_interest'>;
export type CryptoOperation = CryptoAsset & { price: Money };
export interface Deposit { id: number; name: string; bank_name: string; amount_rub: number; annual_interest: number; start_date: string; end_date: string }
export interface Insurance { id: number; name: string; company_name: string; amount_rub: number; annual_interest: number; start_date: string; end_date: string }
export interface CreditCard { id: number; name: string; bank_name: string; mandatory_payment_date: string; mandatory_payment_amount_rub: number; overall_limit_rub: number }
export type DepositUpdate = Pick<Deposit, 'id'> & Partial<Omit<Deposit, 'id' | 'amount_rub'>>;
export type InsuranceUpdate = Pick<Insurance, 'id'> & Partial<Omit<Insurance, 'id' | 'amount_rub'>>;
export type CreditCardUpdate = Pick<CreditCard, 'id'> & Partial<Omit<CreditCard, 'id'>>;
export interface PersonalAssets { deposits: Deposit[]; credit_cards: CreditCard[]; insurances: Insurance[] }
export interface AllInfo { bonds: Bond[]; personal_assets: PersonalAssets; crypto_assets: CryptoAsset[]; stocks: Stock[]; profiles: Profile[] }
export interface SearchResult<T> { search_results: T[]; err_msg: string }
export interface BondCoupons { code: string; coupons: { payment_date: string; amount: Money | null }[] }
export interface StockRebalance { stocks: Stock[]; total_prev_cost: number; cost_diff: number }
export interface BondRebalance { bonds: Bond[]; total_prev_cost: number; cost_diff: number }
export interface TransactionCursor { before: string; before_id?: number }
export interface StockTransaction { id: number; instrument_code: string; lot_amount: number; buying: boolean; lot_size: number; price: number; currency: string; time: string }
export interface CryptoTransaction { id: number; asset_id: number; asset_name: string; platform: string; amount: number; buying: boolean; price: number; currency: string; created_at: string }
export interface DepositTransaction { id: number; name: string; bank_name: string; amount_rub: number; buying: boolean; created_at: string }
export interface InsuranceTransaction { id: number; name: string; company_name: string; amount_rub: number; buying: boolean; created_at: string }
