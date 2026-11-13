package shop

import "testing"

// TestAverageOrder checks empty, seeded, and unknown customers.
func TestAverageOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		customerID int
		want       int
		wantErr    error
	}{
		{name: "no orders", customerID: 3, want: 0},
		{name: "Alice", customerID: 1, want: 79900},
		{name: "Bob", customerID: 2, want: 69800},
		{name: "unknown customer", customerID: 999, wantErr: ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewStore().AverageOrder(tc.customerID)
			if got != tc.want || err != tc.wantErr {
				t.Fatalf("AverageOrder(%d) = (%d, %v), want (%d, %v)", tc.customerID, got, err, tc.want, tc.wantErr)
			}
		})
	}
}

// TestAverageOrderMultipleOrders checks whole-öre truncation and customer filtering.
func TestAverageOrderMultipleOrders(t *testing.T) {
	s := NewStore()
	s.orders[4] = &Order{ID: 4, Customer: s.customers[1], Total: 79901}
	got, err := s.AverageOrder(1)
	if got != 79900 || err != nil {
		t.Fatalf("AverageOrder(1) = (%d, %v), want (79900, nil)", got, err)
	}
}
