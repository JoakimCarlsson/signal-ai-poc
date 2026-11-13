package shop

import (
	"errors"
	"testing"
)

// TestOrderGuest verifies that the seeded guest order can be read without a customer.
func TestOrderGuest(t *testing.T) {
	got, err := NewStore().Order(3)
	if err != nil {
		t.Fatal(err)
	}
	want := OrderSummary{ID: 3, ProductID: 4, Quantity: 1, Total: 129900}
	if got != want {
		t.Fatalf("Order(3) = %+v, want %+v", got, want)
	}
}

// TestOrderCustomer verifies that order summaries retain the customer's name.
func TestOrderCustomer(t *testing.T) {
	got, err := NewStore().Order(1)
	if err != nil {
		t.Fatal(err)
	}
	want := OrderSummary{ID: 1, CustomerName: "Alice", ProductID: 1, Quantity: 1, Total: 79900}
	if got != want {
		t.Fatalf("Order(1) = %+v, want %+v", got, want)
	}
}

// TestOrderNotFound verifies the missing-order error contract.
func TestOrderNotFound(t *testing.T) {
	if _, err := NewStore().Order(999999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Order(999999) error = %v, want ErrNotFound", err)
	}
}

// TestOrderPlacedGuest verifies that a newly placed guest order can be read back.
func TestOrderPlacedGuest(t *testing.T) {
	store := NewStore()
	want, err := store.PlaceOrder(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.Order(want.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != want || got.CustomerName != "" {
		t.Fatalf("Order(%d) = %+v, want %+v with no customer name", want.ID, got, want)
	}
}
