package app

import (
	"context"
	"math"
	"math/big"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"strings"
)

func (a *app) GetSnapshot(ctx context.Context, user int) (model.Snapshot, error) {
	return a.repo.GetSnapshot(ctx, user)
}

func (a *app) InstallSnapshot(ctx context.Context, user int, snapshot model.Snapshot) error {
	if snapshot.Version != 1 || snapshot.Profiles == nil {
		return servererrors.ErrInvalidSnapshot
	}
	defaults := 0
	for _, p := range snapshot.Profiles {
		if p.Default {
			defaults++
		}
		if p.Name == "" || p.StockGroups == nil || p.BondGroups == nil || p.CryptoGroups == nil || p.Deposits == nil || p.Insurances == nil || p.CreditCards == nil || p.Transactions == nil || p.CryptoTransactions == nil || p.DepositTransactions == nil || p.InsuranceTransactions == nil {
			return servererrors.ErrInvalidSnapshot
		}
		if !validSnapshotProfile(p) {
			return servererrors.ErrInvalidSnapshot
		}
	}
	if defaults > 1 {
		return servererrors.ErrInvalidSnapshot
	}
	return a.repo.InstallSnapshot(ctx, user, snapshot)
}

func snapshotText(s string) bool    { return strings.TrimSpace(s) != "" }
func snapshotNumber(n float64) bool { return !math.IsNaN(n) && !math.IsInf(n, 0) && n >= 0 }
func snapshotAmount(s string) bool  { v, ok := new(big.Rat).SetString(s); return ok && v.Sign() > 0 }

func validSnapshotProfile(p model.SnapshotProfile) bool {
	if !snapshotText(p.Name) {
		return false
	}
	for _, g := range p.StockGroups {
		if g.Stocks == nil || g.Weight < 0 || g.Weight > 100 {
			return false
		}
		for _, a := range g.Stocks {
			if a.LotAmount <= 0 {
				return false
			}
		}
	}
	for _, g := range p.BondGroups {
		if g.Bonds == nil || g.Weight < 0 || g.Weight > 100 {
			return false
		}
		for _, a := range g.Bonds {
			if a.LotAmount <= 0 {
				return false
			}
		}
	}
	for _, g := range p.CryptoGroups {
		if !snapshotText(g.Name) || g.Assets == nil || g.Weight < 0 || g.Weight > 100 {
			return false
		}
		for _, a := range g.Assets {
			if a.ID <= 0 || !snapshotText(a.Name) || !snapshotText(a.Platform) || !snapshotNumber(a.AnnualInterest) || !snapshotAmount(a.Amount) {
				return false
			}
		}
	}
	for _, t := range p.Transactions {
		if t.ID <= 0 || (t.InstrumentType != model.InstrumentStock && t.InstrumentType != model.InstrumentBond) || t.LotAmount <= 0 || math.IsNaN(t.Price) || math.IsInf(t.Price, 0) || t.Time.IsZero() {
			return false
		}
	}
	for _, t := range p.CryptoTransactions {
		if t.ID <= 0 || t.AssetID <= 0 || !snapshotText(t.AssetName) || !snapshotText(t.Platform) || !snapshotAmount(t.Amount) || !snapshotNumber(t.Price) || t.Price == 0 || !snapshotText(t.Currency) || t.CreatedAt.IsZero() {
			return false
		}
	}
	deposits := map[int64]bool{}
	for _, t := range p.DepositTransactions {
		_, exists := deposits[t.ID]
		if t.ID <= 0 || exists || !snapshotText(t.Name) || !snapshotText(t.BankName) || !validDepositAmount(t.AmountRUB) || t.CreatedAt.IsZero() {
			return false
		}
		deposits[t.ID] = t.Buying
	}
	insurances := map[int64]bool{}
	for _, t := range p.InsuranceTransactions {
		_, exists := insurances[t.ID]
		if t.ID <= 0 || exists || !snapshotText(t.Name) || !snapshotText(t.CompanyName) || !validInsuranceAmount(t.AmountRUB) || t.CreatedAt.IsZero() {
			return false
		}
		insurances[t.ID] = t.Buying
	}
	for _, d := range p.Deposits {
		if !deposits[d.AcquisitionTransactionID] || !validDepositMetadata(d.Name, d.BankName, d.AnnualInterest, d.EndDate) || !validDepositAmount(d.AmountRUB) || !validContractDates(d.StartDate, d.EndDate) {
			return false
		}
	}
	for _, d := range p.Insurances {
		if !insurances[d.AcquisitionTransactionID] || !validInsuranceMetadata(d.Name, d.CompanyName, d.AnnualInterest, d.EndDate) || !validInsuranceAmount(d.AmountRUB) || !validContractDates(d.StartDate, d.EndDate) {
			return false
		}
	}
	for _, c := range p.CreditCards {
		if !validCreditCard(c) {
			return false
		}
	}
	return true
}
