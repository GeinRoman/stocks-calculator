package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/lib/pq"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
)

// cryptoWrite serializes crypto mutations within a profile, including weight updates.
func (r *repo) cryptoWrite(ctx context.Context, user int, fn func(*sql.Tx, int) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int
	err = tx.QueryRowContext(ctx, `SELECT id FROM profiles WHERE user_id=$1 AND is_default=true FOR UPDATE`, user).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return servererrors.ErrNoDefaultProfile
	}
	if err != nil {
		return err
	}
	if err = fn(tx, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *repo) GetCryptoGroups(ctx context.Context, user int) ([]model.CryptoGroup, error) {
	id, err := r.getDefaultProfileId(ctx, user)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT name, weight FROM crypto_groups WHERE profile_id=$1 ORDER BY name`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := []model.CryptoGroup{}
	for rows.Next() {
		var g model.CryptoGroup
		if err := rows.Scan(&g.Name, &g.Weight); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *repo) InsertCryptoGroups(ctx context.Context, user int, groups []model.CryptoGroup) error {
	return r.cryptoWrite(ctx, user, func(tx *sql.Tx, id int) error {
		for _, g := range groups {
			res, err := tx.ExecContext(ctx, `INSERT INTO crypto_groups (profile_id,name) VALUES ($1,$2) ON CONFLICT (profile_id,name) DO NOTHING`, id, g.Name)
			if err != nil {
				return err
			}
			n, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if n == 0 {
				return fmt.Errorf("%w: %q", servererrors.ErrGroupExists, g.Name)
			}
		}
		return nil
	})
}

func (r *repo) RemoveCryptoGroups(ctx context.Context, user int, groups []model.CryptoGroup) error {
	return r.cryptoWrite(ctx, user, func(tx *sql.Tx, id int) error {
		for _, g := range groups {
			var groupID int
			err := tx.QueryRowContext(ctx, `SELECT id FROM crypto_groups WHERE profile_id=$1 AND name=$2 FOR UPDATE`, id, g.Name).Scan(&groupID)
			if errors.Is(err, sql.ErrNoRows) {
				return servererrors.ErrGroupNotFound
			}
			if err != nil {
				return err
			}
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM crypto_assets WHERE group_id=$1)`, groupID).Scan(&exists); err != nil {
				return err
			}
			if exists {
				return servererrors.ErrCryptoGroupNotEmpty
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM crypto_groups WHERE id=$1`, groupID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repo) UpdateCryptoGroupName(ctx context.Context, user int, old, name string) error {
	return r.cryptoWrite(ctx, user, func(tx *sql.Tx, id int) error {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM crypto_groups WHERE profile_id=$1 AND name=$2 AND name<>$3)`, id, name, old).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return servererrors.ErrGroupExists
		}
		res, err := tx.ExecContext(ctx, `UPDATE crypto_groups SET name=$1 WHERE profile_id=$2 AND name=$3`, name, id, old)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return servererrors.ErrGroupNotFound
		}
		return nil
	})
}

func (r *repo) UpdateCryptoWeights(ctx context.Context, user int, groups []model.CryptoGroup) error {
	return r.cryptoWrite(ctx, user, func(tx *sql.Tx, id int) error {
		for _, g := range groups {
			res, err := tx.ExecContext(ctx, `UPDATE crypto_groups SET weight=$1 WHERE profile_id=$2 AND name=$3`, g.Weight, id, g.Name)
			if err != nil {
				return err
			}
			n, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if n == 0 {
				return servererrors.ErrGroupNotFound
			}
		}
		var sum int
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(weight),0) FROM crypto_groups WHERE profile_id=$1`, id).Scan(&sum); err != nil {
			return err
		}
		if sum > 100 {
			return servererrors.ErrInvalidWeights
		}
		return nil
	})
}

func (r *repo) GetCryptoAssets(ctx context.Context, user int) ([]model.CryptoAsset, error) {
	id, err := r.getDefaultProfileId(ctx, user)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT a.id,a.name,g.name,a.platform,a.url,a.annual_interest,a.amount FROM crypto_assets a JOIN crypto_groups g ON g.id=a.group_id AND g.profile_id=a.profile_id WHERE a.profile_id=$1 ORDER BY a.name`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := []model.CryptoAsset{}
	for rows.Next() {
		var a model.CryptoAsset
		if err := rows.Scan(&a.ID, &a.Name, &a.GroupName, &a.Platform, &a.URL, &a.AnnualInterest, &a.Amount); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (r *repo) UpdateCryptoAsset(ctx context.Context, user int, asset model.CryptoAssetUpdate) error {
	return r.cryptoWrite(ctx, user, func(tx *sql.Tx, profileID int) error {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM crypto_assets WHERE profile_id=$1 AND id=$2)`, profileID, asset.ID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return servererrors.ErrCryptoNotFound
		}
		_, err := tx.ExecContext(ctx, `UPDATE crypto_assets SET name=$1,platform=$2,url=$3,annual_interest=$4 WHERE profile_id=$5 AND id=$6`, asset.Name, asset.Platform, asset.URL, asset.AnnualInterest, profileID, asset.ID)
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return servererrors.ErrCryptoExists
		}
		return err
	})
}

// The API exposes float64, so reject balances that cannot round-trip as decimals.
func cryptoBalanceAfter(balance, quantity string, buy bool) (string, bool) {
	b, ok := new(big.Rat).SetString(balance)
	if !ok || b.Sign() <= 0 {
		return "", false
	}
	q, ok := new(big.Rat).SetString(quantity)
	if !ok || q.Sign() <= 0 {
		return "", false
	}
	if buy {
		b.Add(b, q)
	} else {
		b.Sub(b, q)
	}
	if b.Sign() <= 0 {
		return "", false
	}
	f, _ := b.Float64()
	if math.IsInf(f, 0) || f == 0 {
		return "", false
	}
	next := strconv.FormatFloat(f, 'f', -1, 64)
	roundTrip, ok := new(big.Rat).SetString(next)
	return next, ok && roundTrip.Cmp(b) == 0
}

func (r *repo) TransactCrypto(ctx context.Context, user int, ops []model.CryptoOperation, buy bool) error {
	return r.cryptoWrite(ctx, user, func(tx *sql.Tx, id int) error {
		for _, op := range ops {
			var name, platform string
			quantity := strconv.FormatFloat(op.Amount, 'f', -1, 64)
			assetID := op.ID
			if assetID == 0 && buy {
				var groupID int64
				err := tx.QueryRowContext(ctx, `SELECT id FROM crypto_groups WHERE profile_id=$1 AND name=$2`, id, op.GroupName).Scan(&groupID)
				if errors.Is(err, sql.ErrNoRows) {
					return servererrors.ErrGroupNotFound
				}
				if err != nil {
					return err
				}
				err = tx.QueryRowContext(ctx, `INSERT INTO crypto_assets (profile_id,group_id,name,platform,url,annual_interest,amount) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (profile_id,name) DO NOTHING RETURNING id`, id, groupID, op.Name, op.Platform, op.URL, op.AnnualInterest, quantity).Scan(&assetID)
				if errors.Is(err, sql.ErrNoRows) {
					return servererrors.ErrCryptoExists
				}
				if err != nil {
					return err
				}
				name, platform = op.Name, op.Platform
			} else {
				var amount string
				err := tx.QueryRowContext(ctx, `SELECT name,platform,amount FROM crypto_assets WHERE profile_id=$1 AND id=$2 FOR UPDATE`, id, assetID).Scan(&name, &platform, &amount)
				if errors.Is(err, sql.ErrNoRows) {
					return servererrors.ErrCryptoNotFound
				}
				if err != nil {
					return err
				}
				balance, ok := new(big.Rat).SetString(amount)
				requested, valid := new(big.Rat).SetString(quantity)
				if !ok || !valid {
					return servererrors.ErrInvalidCrypto
				}
				if !buy && requested.Cmp(balance) >= 0 {
					quantity = amount
					_, err = tx.ExecContext(ctx, `DELETE FROM crypto_assets WHERE profile_id=$1 AND id=$2`, id, assetID)
				} else {
					next, exact := cryptoBalanceAfter(amount, quantity, buy)
					if !exact {
						return fmt.Errorf("%w: quantity cannot be represented exactly in the resulting balance", servererrors.ErrInvalidCrypto)
					}
					_, err = tx.ExecContext(ctx, `UPDATE crypto_assets SET amount=$1 WHERE profile_id=$2 AND id=$3`, next, id, assetID)
				}
				if err != nil {
					return err
				}
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO crypto_transactions (profile_id,asset_id,asset_name,platform,amount,buying,price,currency) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, assetID, name, platform, quantity, buy, op.Price.Price, op.Price.Currency)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
