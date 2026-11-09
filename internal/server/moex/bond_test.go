package moex

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// All MOEX requests in these tests must be served locally.
func bondTestClient(t *testing.T, handler http.HandlerFunc) *moex {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &moex{client: server.Client(), maxConns: 4, baseUrl: server.URL}
}

func bondTestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path == "/securities.json" {
		switch r.URL.Query().Get("q") {
		case "mixed & bonds":
			fmt.Fprint(w, `{"securities":{"columns":["secid","shortname","type"],"data":[["SINGLE","Single bond","corporate_bond"],["SHARE","A share","common_share"],["MULTI","Amortizing bond","corporate_bond"]]}}`)
		case "bad search":
			fmt.Fprint(w, `{`)
		default:
			http.Error(w, "unexpected search", http.StatusNotFound)
		}
		return
	}

	path := r.URL.Path
	code := strings.TrimSuffix(path[strings.LastIndex(path, "/")+1:], ".json")
	if code == "HTTPERR" {
		http.Error(w, "MOEX unavailable", http.StatusServiceUnavailable)
		return
	}
	if strings.HasPrefix(path, "/securities/") {
		if code == "BADBOARD" {
			fmt.Fprint(w, `{"boards":{"columns":["boardid","is_primary"],"data":[["TQOB"]]}}`)
			return
		}
		if code == "USD" {
			fmt.Fprint(w, `{"boards":{"columns":["boardid","engine","market","is_traded","is_primary","unit","currencyid"],"data":[["TQOB","stock","bonds",1,1,"%","USD"]]}}`)
			return
		}
		fmt.Fprint(w, `{"boards":{"columns":["boardid","engine","market","is_traded","is_primary","unit","currencyid"],"data":[["TQXX","stock","bonds",1,0,"%","RUB"],["TQOB","stock","bonds",1,1,"%","RUB"]]}}`)
		return
	}
	if strings.HasPrefix(path, "/engines/stock/markets/bonds/boards/TQOB/securities/") {
		if code == "BADSNAP" {
			fmt.Fprint(w, `{"securities":{"columns":["LOTSIZE","FACEVALUEONSETTLEDATE"],"data":[[10]]},"marketdata":{"columns":["LAST"],"data":[[98]]}}`)
			return
		}
		last, market, prev := "98.5", "97", "96"
		switch code {
		case "MARKET":
			last = "null"
		case "PREVIOUS":
			last, market = "null", "null"
		case "NOPRICE":
			last, market, prev = "null", "null", "null"
		}
		fmt.Fprintf(w, `{"securities":{"columns":["FACEVALUEONSETTLEDATE","ACCRUEDINT","LOTSIZE","MATDATE","FACEUNIT","CURRENCYID","PREVPRICE"],"data":[[1000,12.5,10,"2030-01-01","SUR","RUB",%s]]},"marketdata":{"columns":["LAST","MARKETPRICE"],"data":[[%s,%s]]}}`, prev, last, market)
		return
	}
	if strings.HasPrefix(path, "/statistics/engines/stock/markets/bonds/bondization/") {
		if code == "PAGED" {
			if r.URL.Query().Get("start") == "0" {
				fmt.Fprint(w, `{"amortizations":{"columns":["amortdate","value","faceunit"],"data":[["2029-01-01",500,"RUB"]]},"amortizations.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,2,1]]}}`)
			} else {
				fmt.Fprint(w, `{"amortizations":{"columns":["amortdate","value","faceunit"],"data":[["2030-01-01",500,"RUB"]]},"amortizations.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[1,2,1]]}}`)
			}
			return
		}
		rows := `["2030-01-01",1000,100,1000,"RUB"]`
		if code == "MULTI" {
			rows = `["2029-01-01",500,50,1000,"RUB"],["2030-01-01",500,50,500,"RUB"]`
		}
		fmt.Fprintf(w, `{"amortizations":{"columns":["amortdate","value","valueprc","facevalue","faceunit"],"data":[%s]},"amortizations.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,%d,100]]}}`, rows, strings.Count(rows, "["))
		return
	}
	http.Error(w, "unexpected endpoint", http.StatusNotFound)
}

func TestGetBondsInfoPaginatedSchedule(t *testing.T) {
	client := bondTestClient(t, bondTestHandler)
	result := client.GetBondsInfo(context.Background(), []string{"PAGED"})[0]
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if result.MaturityPayment != nil {
		t.Fatalf("expected null maturity payment for paginated schedule, got %v", *result.MaturityPayment)
	}
}

func TestGetBondsInfoConversionAndBatchOrder(t *testing.T) {
	client := bondTestClient(t, bondTestHandler)
	codes := []string{"SINGLE", "MARKET", "PREVIOUS", "MULTI", "HTTPERR"}
	results := client.GetBondsInfo(context.Background(), codes)
	if len(results) != len(codes) {
		t.Fatalf("got %d results, want %d", len(results), len(codes))
	}
	for i, wantPrice := range []float64{985, 970, 960, 985} {
		got := results[i]
		if got.Err != nil {
			t.Errorf("%s: unexpected error: %v", codes[i], got.Err)
			continue
		}
		if got.CleanPrice != wantPrice || got.AccruedInterest != 12.5 || got.LotSize != 10 || got.MaturityDate != "2030-01-01" {
			t.Errorf("%s: unexpected bond: %+v", codes[i], got.Bond)
		}
		if i == 3 {
			if got.MaturityPayment != nil {
				t.Errorf("%s: multi-payment bond has maturity payment %v", codes[i], *got.MaturityPayment)
			}
		} else if got.MaturityPayment == nil || *got.MaturityPayment != 1000 {
			t.Errorf("%s: maturity payment = %v, want 1000", codes[i], got.MaturityPayment)
		}
	}
	if results[4].Err == nil {
		t.Error("HTTP error must be reported for its batch item")
	}
}

func TestGetBondsInfoErrors(t *testing.T) {
	client := bondTestClient(t, bondTestHandler)
	codes := []string{"BADBOARD", "BADSNAP", "NOPRICE", "USD"}
	results := client.GetBondsInfo(context.Background(), codes)
	if len(results) != len(codes) {
		t.Fatalf("got %d results, want %d", len(results), len(codes))
	}
	for i, result := range results {
		if result.Err == nil {
			t.Errorf("%s: expected error", codes[i])
		}
	}
}

func TestFindBondsSearchAndBatchErrors(t *testing.T) {
	client := bondTestClient(t, bondTestHandler)
	results := client.FindBonds(context.Background(), []string{"mixed & bonds", "bad search"})
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("search: %v", results[0].Err)
	}
	found := results[0].SearchResults
	if len(found) != 2 {
		t.Fatalf("got %d bond search results, want 2 (shares must be filtered)", len(found))
	}
	for i, want := range []struct{ code, name string }{{"SINGLE", "Single bond"}, {"MULTI", "Amortizing bond"}} {
		if found[i].Err != nil || found[i].Code != want.code || found[i].Name != want.name || found[i].CleanPrice != 985 {
			t.Errorf("result %d: got %+v, want %s (%s) with price 985", i, found[i], want.code, want.name)
		}
	}
	if results[1].Err == nil {
		t.Error("malformed search JSON must produce an error")
	}
}
