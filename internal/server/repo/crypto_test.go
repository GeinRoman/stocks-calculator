package repo

import "testing"

func TestCryptoBalanceAfter(t *testing.T) {
	tests := []struct {
		name, balance, quantity string
		buy                     bool
		want                    string
		valid                   bool
	}{
		{"decimal buy", "0.1", "0.2", true, "0.3", true},
		{"decimal sell", "0.3", "0.1", false, "0.2", true},
		{"fractional sell", "1", "0.1", false, "0.9", true},
		{"large unrepresentable buy", "10000000000000000", "3", true, "", false},
		{"large unrepresentable sell", "10000000000000000", "3", false, "", false},
		{"large representable buy", "10000000000000000", "4", true, "10000000000000004", true},
		{"tiny decimal", "0.00000001", "0.00000002", true, "0.00000003", true},
		{"invalid", "1", "NaN", true, "", false},
		{"full sale uses delete", "1", "1", false, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := cryptoBalanceAfter(tt.balance, tt.quantity, tt.buy)
			if ok != tt.valid || (ok && got != tt.want) {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, ok, tt.want, tt.valid)
			}
		})
	}
}
