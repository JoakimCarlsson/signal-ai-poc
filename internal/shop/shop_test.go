package shop

import (
	"errors"
	"testing"
)

// TestOrder checks seeded guest and named orders and the missing-order error.
func TestOrder(t *testing.T) {
	for _, tt := range []struct {
		name string
		want OrderSummary
	}{
		{name: "guest", want: OrderSummary{ID: 3, ProductID: 4, Quantity: 1, Total: 129900}},
		{name: "Alice", want: OrderSummary{ID: 1, CustomerName: "Alice", ProductID: 1, Quantity: 1, Total: 79900}},
		{name: "Bob", want: OrderSummary{ID: 2, CustomerName: "Bob", ProductID: 2, Quantity: 2, Total: 69800}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStore().Order(tt.want.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("Order(%d) = %+v, want %+v", tt.want.ID, got, tt.want)
			}
		})
	}

	if _, err := NewStore().Order(999999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Order(999999) error = %v, want ErrNotFound", err)
	}
}

// TestOrderAfterPlaceOrder checks that a newly placed guest order reads unchanged.
func TestOrderAfterPlaceOrder(t *testing.T) {
	store := NewStore()
	placed, err := store.PlaceOrder(4, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := OrderSummary{ID: placed.ID, ProductID: 4, Quantity: 1, Total: 129900}
	got, err := store.Order(placed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != want || got != placed {
		t.Fatalf("Order(%d) = %+v, want %+v (placed %+v)", placed.ID, got, want, placed)
	}
}
