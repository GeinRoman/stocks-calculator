package moex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestBondSearchRanksWordsBeforeLimit(t *testing.T) {
	var mu sync.Mutex
	lookups := map[string]bool{}
	client := bondTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/securities.json" {
			rows := make([][]any, 0, 13)
			for i := 0; i < 12; i++ {
				rows = append(rows, []any{fmt.Sprintf("OTHER%02d", i), "Other", "Газпром нефть выпуск", nil, "exchange_bond", 1})
			}
			rows = append(rows, []any{"TARGET", "Газпнф3Р2R", "Газпром нефть выпуск", "RU000A1017J5", "exchange_bond", 1})
			_ = json.NewEncoder(w).Encode(map[string]any{"securities": map[string]any{"columns": []string{"secid", "shortname", "name", "isin", "type", "is_traded"}, "data": rows}})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/securities/") {
			mu.Lock()
			lookups[strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/securities/"), ".json")] = true
			mu.Unlock()
		}
		bondTestHandler(w, r)
	})
	result := client.FindBonds(context.Background(), []string{"Газпром нефть 3P2R"})[0]
	if result.Err != nil || len(result.SearchResults) != 10 || result.SearchResults[0].Code != "TARGET" {
		t.Fatalf("unexpected search result: %+v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(lookups) != 10 || !lookups["TARGET"] || lookups["OTHER09"] {
		t.Fatalf("unexpected detail lookups: %v", lookups)
	}
}

func TestBondSearchRanksNameCodeAndISIN(t *testing.T) {
	client := bondTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/securities.json" {
			fmt.Fprint(w, `{"securities":{"columns":["secid","shortname","name","isin","type"],"data":[["FIRST","First","other",null,"ofz_bond"],["FULL","Second","Needle bond",null,"ofz_bond"],["CODE-NEEDLE","Third","other",null,"ofz_bond"],["FOURTH","Fourth","other","NEEDLE","ofz_bond"],["FIFTH","Needle","other",null,"ofz_bond"]]}}`)
			return
		}
		bondTestHandler(w, r)
	})
	result := client.FindBonds(context.Background(), []string{"nEeDlE"})[0]
	want := []string{"FULL", "CODE-NEEDLE", "FOURTH", "FIFTH", "FIRST"}
	if result.Err != nil || len(result.SearchResults) != len(want) {
		t.Fatalf("unexpected result: %+v", result)
	}
	for i, code := range want {
		if result.SearchResults[i].Err != nil || result.SearchResults[i].Code != code {
			t.Fatalf("result %d: %+v, want %s", i, result.SearchResults[i], code)
		}
	}
}
