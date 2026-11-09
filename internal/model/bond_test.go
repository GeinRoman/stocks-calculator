package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBondMaturityPaymentJSON(t *testing.T) {
	payment := 1000.0
	for _, tc := range []struct {
		name    string
		payment *float64
		want    string
	}{
		{name: "single payment", payment: &payment, want: `"maturity_payment":1000`},
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
