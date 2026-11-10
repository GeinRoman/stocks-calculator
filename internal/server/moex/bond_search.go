package moex

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"stocks_calculator/internal/model"
	"strings"
	"sync"
)

const maxBondSearchResults = 10

func (m *moex) FindBonds(ctx context.Context, names []string) []model.MoexBondFindResult {
	results := make([]model.MoexBondFindResult, len(names))
	var wg sync.WaitGroup
	limit := max(m.maxConns, 1)
	sem := make(chan struct{}, limit)
	for i, name := range names {
		wg.Go(func() { results[i] = m.findBond(ctx, name, sem) })
	}
	wg.Wait()
	return results
}

func (m *moex) findBond(ctx context.Context, query string, sem chan struct{}) model.MoexBondFindResult {
	var response struct {
		Securities bondTable `json:"securities"`
	}
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return model.MoexBondFindResult{Err: ctx.Err()}
	}
	err := m.bondRequest(ctx, "/securities.json", url.Values{"q": {query}}, &response)
	<-sem
	if err != nil {
		return model.MoexBondFindResult{Err: err}
	}
	rows, err := response.Securities.rows("secid", "shortname", "type")
	if err != nil {
		return model.MoexBondFindResult{Err: err}
	}
	words := strings.Fields(normalizeBondSearch(query))
	type candidate struct {
		result model.MoexBondSearchResult
		score  int
	}
	var candidates []candidate
	for _, row := range rows {
		typ, ok := row["type"].(string)
		if !ok {
			return model.MoexBondFindResult{Err: fmt.Errorf("invalid security type")}
		}
		switch typ {
		case "ofz_bond", "cb_bond", "state_bond", "subfederal_bond", "municipal_bond", "corporate_bond", "exchange_bond", "ifi_bond", "euro_bond", "non_exchange_bond":
		default:
			continue
		}
		if traded, exists := row["is_traded"]; exists {
			if traded == float64(0) {
				continue
			}
			if traded != float64(1) {
				return model.MoexBondFindResult{Err: fmt.Errorf("invalid bond trading status")}
			}
		}
		code, ok := row["secid"].(string)
		if !ok || code == "" {
			return model.MoexBondFindResult{Err: fmt.Errorf("invalid bond code")}
		}
		name, ok := row["shortname"].(string)
		if !ok || name == "" {
			return model.MoexBondFindResult{Err: fmt.Errorf("invalid bond name")}
		}
		fullName, _ := row["name"].(string)
		isin, _ := row["isin"].(string)
		fields := []string{normalizeBondSearch(code), normalizeBondSearch(name), normalizeBondSearch(fullName), normalizeBondSearch(isin)}
		score := 0
		for _, word := range words {
			for _, field := range fields {
				if strings.Contains(field, word) {
					score++
					break
				}
			}
		}
		candidates = append(candidates, candidate{result: model.MoexBondSearchResult{Bond: model.Bond{Code: code, Name: name}}, score: score})
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	if len(candidates) > maxBondSearchResults {
		candidates = candidates[:maxBondSearchResults]
	}
	var wg sync.WaitGroup
	for i := range candidates {
		wg.Go(func() {
			item := &candidates[i].result
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				item.Err = ctx.Err()
				return
			}
			defer func() { <-sem }()
			bond, err := m.getBondInfo(ctx, item.Code)
			bond.Code, bond.Name = item.Code, item.Name
			item.Bond, item.Err = bond, err
		})
	}
	wg.Wait()
	var result model.MoexBondFindResult
	for _, candidate := range candidates {
		if !errors.Is(candidate.result.Err, errUnsupportedBondCurrency) {
			result.SearchResults = append(result.SearchResults, candidate.result)
		}
	}
	return result
}

func normalizeBondSearch(value string) string {
	return strings.ReplaceAll(strings.ToLower(value), "р", "p")
}
