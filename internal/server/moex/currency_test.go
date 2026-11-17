package moex

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestCurrencyReferenceRates(t *testing.T) {
	client := bondTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/statistics/engines/currency/markets/selt/rates.json" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"cbrf":{"columns":["CBRF_EUR_LAST","CBRF_USD_LAST"],"data":[[90,80]]},"wap_rates":{"columns":["price","nominal","secid"],"data":[[1200,100,"CNYRUB_TOM"]]}}`)
	})
	rates, err := client.GetCurrencyRates(context.Background())
	if err != nil || rates["RUB"] != 1 || rates["USD"] != 80 || rates["EUR"] != 90 || rates["CNY"] != 12 {
		t.Fatalf("rates %v, error %v", rates, err)
	}
}

func TestForeignBondSeparateCurrencies(t *testing.T) {
	client := bondTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/securities/"):
			fmt.Fprint(w, `{"boards":{"columns":["boardid","engine","market","is_traded","is_primary","unit","currencyid"],"data":[["TQCB","stock","bonds",1,1,"%","RUB"],["TQOD","stock","bonds",1,0,"%","USD"]]}}`)
		case strings.HasPrefix(r.URL.Path, "/engines/"):
			fmt.Fprint(w, `{"securities":{"columns":["LOTSIZE","FACEVALUEONSETTLEDATE","ACCRUEDINT","MATDATE","FACEUNIT","CURRENCYID"],"data":[[1,1000,1424.18,"2029-01-27","USD","SUR"]]},"marketdata":{"columns":["LAST"],"data":[[86.151]]}}`)
		case strings.HasPrefix(r.URL.Path, "/statistics/"):
			fmt.Fprint(w, `{"amortizations":{"columns":["amortdate","value","faceunit"],"data":[["2029-01-27",1000,"USD"]]},"amortizations.cursor":{"columns":["INDEX","TOTAL","PAGESIZE"],"data":[[0,1,20]]}}`)
		default:
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	})
	got := client.GetBondsInfo(context.Background(), []string{"RU000A105KU0"})[0]
	if got.Err != nil || got.CleanPrice.Currency != "USD" || got.CleanPrice.Price != 861.51 || got.AccruedInterest.Currency != "RUB" || got.AccruedInterest.Price != 1424.18 || got.MaturityPayment == nil || got.MaturityPayment.Currency != "USD" || got.MaturityPayment.Price != 1000 {
		t.Fatalf("unexpected foreign bond: %+v", got)
	}
}
