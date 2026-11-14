package repo

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/lib/pq"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

func (r *repo) InstallSnapshot(ctx context.Context, user int, snapshot model.Snapshot) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM profiles WHERE user_id=$1`, user); err != nil {
		return err
	}
	for _, p := range snapshot.Profiles {
		if err = installSnapshotProfile(ctx, tx, user, p); err != nil {
			var pg *pq.Error
			if errors.As(err, &pg) && (pg.Code.Class() == "22" || pg.Code.Class() == "23") {
				return servererrors.ErrInvalidSnapshot
			}
			return err
		}
	}
	return tx.Commit()
}

func installSnapshotProfile(ctx context.Context, tx *sql.Tx, user int, p model.SnapshotProfile) error {
	var profile int
	if err := tx.QueryRowContext(ctx, `INSERT INTO profiles (user_id,name,is_default) VALUES ($1,$2,$3) RETURNING id`, user, p.Name, p.Default).Scan(&profile); err != nil {
		return err
	}
	for _, g := range p.StockGroups {
		var id int64
		if err := tx.QueryRowContext(ctx, `INSERT INTO groups(profile_id,name,weight) VALUES($1,$2,$3) RETURNING id`, profile, g.Name, g.Weight).Scan(&id); err != nil {
			return err
		}
		for _, s := range g.Stocks {
			if _, err := tx.ExecContext(ctx, `INSERT INTO stocks(group_id,name,code,lot_amount,lot_size) VALUES($1,$2,$3,$4,$5)`, id, s.Name, s.Code, s.LotAmount, s.LotSize); err != nil {
				return err
			}
		}
	}
	for _, g := range p.BondGroups {
		var id int64
		if err := tx.QueryRowContext(ctx, `INSERT INTO bond_groups(profile_id,name,weight) VALUES($1,$2,$3) RETURNING id`, profile, g.Name, g.Weight).Scan(&id); err != nil {
			return err
		}
		for _, s := range g.Bonds {
			if _, err := tx.ExecContext(ctx, `INSERT INTO bonds(group_id,name,code,lot_amount,lot_size) VALUES($1,$2,$3,$4,$5)`, id, s.Name, s.Code, s.LotAmount, s.LotSize); err != nil {
				return err
			}
		}
	}
	cryptoIDs := map[int64]int64{}
	cryptoID := func(old int64) (int64, error) {
		if old <= 0 {
			return 0, servererrors.ErrInvalidSnapshot
		}
		if id, ok := cryptoIDs[old]; ok {
			return id, nil
		}
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT nextval(pg_get_serial_sequence('crypto_assets','id'))`).Scan(&id); err != nil {
			return 0, err
		}
		cryptoIDs[old] = id
		return id, nil
	}
	seenAssets := map[int64]bool{}
	for _, g := range p.CryptoGroups {
		var group int64
		if err := tx.QueryRowContext(ctx, `INSERT INTO crypto_groups(profile_id,name,weight) VALUES($1,$2,$3) RETURNING id`, profile, g.Name, g.Weight).Scan(&group); err != nil {
			return err
		}
		for _, a := range g.Assets {
			if seenAssets[a.ID] {
				return servererrors.ErrInvalidSnapshot
			}
			seenAssets[a.ID] = true
			id, err := cryptoID(a.ID)
			if err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO crypto_assets(id,profile_id,group_id,name,platform,url,annual_interest,amount) VALUES($1,$2,$3,$4,$5,$6,$7,$8::numeric)`, id, profile, group, a.Name, a.Platform, a.URL, a.AnnualInterest, a.Amount); err != nil {
				return err
			}
		}
	}
	sort.Slice(p.Transactions, func(i, j int) bool { return p.Transactions[i].ID < p.Transactions[j].ID })
	var previous int64
	for _, t := range p.Transactions {
		if t.ID <= previous {
			return servererrors.ErrInvalidSnapshot
		}
		previous = t.ID
		if _, err := tx.ExecContext(ctx, `INSERT INTO transactions(profile_id,instrument_code,instrument_type,lot_amount,buying,lot_size,price,currency,datetime) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, profile, t.InstrumentCode, t.InstrumentType, t.LotAmount, t.Buying, t.LotSize, t.Price, t.Currency, t.Time); err != nil {
			return err
		}
	}
	sort.Slice(p.CryptoTransactions, func(i, j int) bool { return p.CryptoTransactions[i].ID < p.CryptoTransactions[j].ID })
	previous = 0
	for _, t := range p.CryptoTransactions {
		if t.ID <= previous {
			return servererrors.ErrInvalidSnapshot
		}
		previous = t.ID
		id, err := cryptoID(t.AssetID)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO crypto_transactions(profile_id,asset_id,asset_name,platform,amount,buying,price,currency,created_at) VALUES($1,$2,$3,$4,$5::numeric,$6,$7,$8,$9)`, profile, id, t.AssetName, t.Platform, t.Amount, t.Buying, t.Price, t.Currency, t.CreatedAt); err != nil {
			return err
		}
	}
	depositIDs := map[int64]int64{}
	sort.Slice(p.DepositTransactions, func(i, j int) bool { return p.DepositTransactions[i].ID < p.DepositTransactions[j].ID })
	for _, t := range p.DepositTransactions {
		if t.ID <= 0 || depositIDs[t.ID] != 0 {
			return servererrors.ErrInvalidSnapshot
		}
		var id int64
		if err := tx.QueryRowContext(ctx, `INSERT INTO deposit_transactions(profile_id,name,bank_name,amount_rub,buying,created_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, profile, t.Name, t.BankName, t.AmountRUB, t.Buying, t.CreatedAt).Scan(&id); err != nil {
			return err
		}
		depositIDs[t.ID] = id
	}
	insuranceIDs := map[int64]int64{}
	sort.Slice(p.InsuranceTransactions, func(i, j int) bool { return p.InsuranceTransactions[i].ID < p.InsuranceTransactions[j].ID })
	for _, t := range p.InsuranceTransactions {
		if t.ID <= 0 || insuranceIDs[t.ID] != 0 {
			return servererrors.ErrInvalidSnapshot
		}
		var id int64
		if err := tx.QueryRowContext(ctx, `INSERT INTO insurance_transactions(profile_id,name,company_name,amount_rub,buying,created_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, profile, t.Name, t.CompanyName, t.AmountRUB, t.Buying, t.CreatedAt).Scan(&id); err != nil {
			return err
		}
		insuranceIDs[t.ID] = id
	}
	for _, d := range p.Deposits {
		id := depositIDs[d.AcquisitionTransactionID]
		if id == 0 {
			return servererrors.ErrInvalidSnapshot
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO deposits(profile_id,name,bank_name,amount_rub,annual_interest,start_date,end_date,acquisition_transaction_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, profile, d.Name, d.BankName, d.AmountRUB, d.AnnualInterest, d.StartDate, d.EndDate, id); err != nil {
			return err
		}
	}
	for _, d := range p.Insurances {
		id := insuranceIDs[d.AcquisitionTransactionID]
		if id == 0 {
			return servererrors.ErrInvalidSnapshot
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO insurances(profile_id,name,company_name,amount_rub,annual_interest,start_date,end_date,acquisition_transaction_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, profile, d.Name, d.CompanyName, d.AmountRUB, d.AnnualInterest, d.StartDate, d.EndDate, id); err != nil {
			return err
		}
	}
	for _, c := range p.CreditCards {
		if _, err := tx.ExecContext(ctx, `INSERT INTO credit_cards(profile_id,name,bank_name,mandatory_payment_date,mandatory_payment_amount_rub,overall_limit_rub) VALUES($1,$2,$3,$4,$5,$6)`, profile, c.Name, c.BankName, c.MandatoryPaymentDate, c.MandatoryPaymentAmountRUB, c.OverallLimitRUB); err != nil {
			return err
		}
	}
	return nil
}
