package moex

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/servererrors"
	"strconv"
	"strings"
)

func (m *moex) GetBondCoupons(ctx context.Context, code string) (model.BondCouponsResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return model.BondCouponsResponse{}, servererrors.ErrMoexBondNotFound
	}
	escaped := url.PathEscape(code)
	var security struct {
		Description bondTable `json:"description"`
	}
	if err := m.bondRequest(ctx, "/securities/"+escaped+".json", url.Values{"iss.only": {"description"}}, &security); err != nil {
		return model.BondCouponsResponse{}, err
	}
	description, err := security.Description.rows("name", "value")
	if err != nil {
		return model.BondCouponsResponse{}, err
	}
	if len(description) == 0 {
		return model.BondCouponsResponse{}, servererrors.ErrMoexBondNotFound
	}
	var group, currency string
	for _, row := range description {
		switch row["name"] {
		case "GROUP":
			group, _ = row["value"].(string)
		case "FACEUNIT":
			currency, _ = row["value"].(string)
		}
	}
	if group == "" {
		return model.BondCouponsResponse{}, fmt.Errorf("incomplete MOEX bond description")
	}
	if group != "stock_bonds" {
		return model.BondCouponsResponse{}, servererrors.ErrMoexBondNotFound
	}
	if currency == "" {
		return model.BondCouponsResponse{}, fmt.Errorf("incomplete MOEX bond description")
	}
	if _, err := currencyCode(currency); err != nil {
		return model.BondCouponsResponse{}, fmt.Errorf("invalid bond currency: %w", err)
	}

	response := model.BondCouponsResponse{Code: code, Coupons: []model.Coupon{}}
	start, total := 0, -1
	for {
		var page struct {
			Coupons bondTable `json:"coupons"`
			Cursor  bondTable `json:"coupons.cursor"`
		}
		query := url.Values{"iss.only": {"coupons,coupons.cursor"}, "start": {strconv.Itoa(start)}}
		if err := m.bondRequest(ctx, "/statistics/engines/stock/markets/bonds/bondization/"+escaped+".json", query, &page); err != nil {
			return model.BondCouponsResponse{}, err
		}
		rows, err := page.Coupons.rows("coupondate", "value", "faceunit")
		if err != nil {
			return model.BondCouponsResponse{}, err
		}
		cursor, err := page.Cursor.rows("INDEX", "TOTAL", "PAGESIZE")
		if err != nil || len(cursor) != 1 {
			return model.BondCouponsResponse{}, fmt.Errorf("invalid coupon cursor")
		}
		index, iok := cursor[0]["index"].(float64)
		count, tok := cursor[0]["total"].(float64)
		size, sok := cursor[0]["pagesize"].(float64)
		if !iok || !tok || !sok || index != float64(start) || count < 0 || count != math.Trunc(count) || count >= float64(math.MaxInt) || size < 1 || size != math.Trunc(size) || float64(len(rows)) > size || float64(start)+float64(len(rows)) > count || (len(rows) == 0 && float64(start) < count) {
			return model.BondCouponsResponse{}, fmt.Errorf("incomplete coupon schedule")
		}
		if total == -1 {
			total = int(count)
		} else if total != int(count) {
			return model.BondCouponsResponse{}, fmt.Errorf("coupon schedule changed during lookup")
		}
		for _, row := range rows {
			date, ok := row["coupondate"].(string)
			if !ok || !validBondDate(date) {
				return model.BondCouponsResponse{}, fmt.Errorf("invalid coupon date")
			}
			coupon := model.Coupon{PaymentDate: date}
			if row["value"] != nil {
				couponCurrency, err := currencyCode(row["faceunit"])
				if err != nil {
					return model.BondCouponsResponse{}, fmt.Errorf("invalid coupon currency")
				}
				amount, ok := row["value"].(float64)
				if !ok || amount < 0 || math.IsInf(amount, 0) || math.IsNaN(amount) {
					return model.BondCouponsResponse{}, fmt.Errorf("invalid coupon amount")
				}
				coupon.Amount = &model.Money{Currency: couponCurrency, Price: amount}
			}
			response.Coupons = append(response.Coupons, coupon)
		}
		start += len(rows)
		if start == total {
			break
		}
	}
	sort.SliceStable(response.Coupons, func(i, j int) bool {
		return response.Coupons[i].PaymentDate < response.Coupons[j].PaymentDate
	})
	return response, nil
}
