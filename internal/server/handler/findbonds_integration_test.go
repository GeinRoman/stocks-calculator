package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"
	"stocks_calculator/internal/server/moex"
)

type findBondsIntegrationRepo struct{ app.Repository }

type findBondsIntegrationApp struct {
	App
	application interface {
		FindBonds(context.Context, int, []string) ([]model.FoundBond, error)
	}
}

func (a findBondsIntegrationApp) FindBonds(ctx context.Context, user int, names []string) ([]model.FoundBond, error) {
	return a.application.FindBonds(ctx, user, names)
}

func (findBondsIntegrationRepo) GetBonds(context.Context, int) ([]model.Bond, error) {
	return []model.Bond{}, nil
}

// Set MOEX_INTEGRATION=1 to run against the live MOEX ISS API.
func TestFindBondsMOEXIntegration(t *testing.T) {
	if os.Getenv("MOEX_INTEGRATION") != "1" {
		t.Skip("set MOEX_INTEGRATION=1 to query the live MOEX API")
	}
	application := app.New(findBondsIntegrationRepo{}, nil, moex.New(10))
	server := httptest.NewServer(NewHttpServer(findBondsIntegrationApp{application: application}, bondTestTokens{}, 0).Handler)
	defer server.Close()
	client := &http.Client{Timeout: 45 * time.Second}

	for _, tc := range []struct{ query, code string }{
		{"ОФЗ 26249", "SU26249RMFS1"},
		{"ОФЗ 26241", "SU26241RMFS8"},
		{"ОФЗ 26221", "SU26221RMFS0"},
		{"ОФЗ 26252", "SU26252RMFS5"},
		{"ОФЗ 26225", "SU26225RMFS1"},
		{"Газпром нефть 3P2R", "RU000A1017J5"},
		{"RU000A1017J5", "RU000A1017J5"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			body, err := json.Marshal([]string{tc.query})
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, server.URL+"/findbonds", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", "Bearer alice")
			req.Header.Set("Content-Type", "application/json")
			response, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				message, _ := io.ReadAll(response.Body)
				t.Fatalf("HTTP %d: %s", response.StatusCode, message)
			}
			var results []model.FoundBond
			if err := json.NewDecoder(response.Body).Decode(&results); err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 {
				t.Fatalf("got %d response entries, want 1", len(results))
			}
			for _, bond := range results[0].SearchResults {
				if bond.Code == tc.code {
					return
				}
			}
			t.Errorf("bond %q not found in response: %+v", tc.code, results[0])
		})
	}
}
