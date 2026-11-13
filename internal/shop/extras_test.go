package shop

import "testing"

// TestShippingEstimate checks short normalized inputs, byte-length boundaries, and existing regional estimates.
func TestShippingEstimate(t *testing.T) {
	for _, tc := range []struct {
		postalCode string
		days       int
		wantError  bool
	}{
		{postalCode: "", wantError: true},
		{postalCode: "  ", wantError: true},
		{postalCode: "4", wantError: true},
		{postalCode: "42", wantError: true},
		{postalCode: "4 2", wantError: true},
		{postalCode: "é", wantError: true},
		{postalCode: "123", days: 1},
		{postalCode: "12345", days: 1},
		{postalCode: "123 45", days: 1},
		{postalCode: "19999", days: 1},
		{postalCode: "20000", days: 2},
		{postalCode: "59999", days: 2},
		{postalCode: "60000", days: 4},
		{postalCode: "abc", days: 4},
		{postalCode: "€", days: 4},
	} {
		t.Run(tc.postalCode, func(t *testing.T) {
			days, err := ShippingEstimate(tc.postalCode)
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, want error = %t", err, tc.wantError)
			}
			if days != tc.days {
				t.Errorf("days = %d, want %d", days, tc.days)
			}
		})
	}
}
