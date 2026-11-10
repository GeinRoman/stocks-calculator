package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBondMaturityPaymentJSON(t *testing.T) {
	payment := Money{Currency: "RUB", Price: 1000}
	for _, tc := range []struct {
		name    string
		payment *Money
		want    string
	}{
		{name: "single payment", payment: &payment, want: `"maturity_payment":{"currency":"RUB","price":1000}`},
		{name: "multiple payments", want: `"maturity_payment":null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(Bond{MaturityPayment: tc.payment})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), tc.want) {
				t.Errorf("got %s, want %s", data, tc.want)
			}
		})
	}
}

func TestBondMissingISINJSON(t *testing.T) {
	data, err := json.Marshal(Bond{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"isin":""`) {
		t.Fatalf("missing ISIN must be an empty JSON string: %s", data)
	}
}
