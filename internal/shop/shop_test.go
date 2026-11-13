package shop

import (
	"errors"
	"testing"
)

// TestPlaceOrder_ZeroQuantity reproduces the integer divide by zero panic when
// PlaceOrder is called with quantity=0.
func TestPlaceOrder_ZeroQuantity(t *testing.T) {
	store := NewStore()
	_, err := store.PlaceOrder(3, 0)
	if err == nil {
		t.Error("PlaceOrder with quantity=0 should fail, got success")
	}
}

// TestPlaceOrder_ValidOrder verifies that normal order placement still works.
func TestPlaceOrder_ValidOrder(t *testing.T) {
	store := NewStore()
	order, err := store.PlaceOrder(1, 2)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}
	if order.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", order.Quantity)
	}
	if order.Total <= 0 {
		t.Errorf("expected positive total, got %d", order.Total)
	}
}

// TestPlaceOrder_NonExistentProduct verifies product not found error.
func TestPlaceOrder_NonExistentProduct(t *testing.T) {
	store := NewStore()
	_, err := store.PlaceOrder(999, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent product, got %v", err)
	}
}

// TestPlaceOrder_OutOfStock verifies out of stock error.
func TestPlaceOrder_OutOfStock(t *testing.T) {
	store := NewStore()
	// Product 5 (Webcam) has 0 stock
	_, err := store.PlaceOrder(5, 1)
	if !errors.Is(err, ErrOutOfStock) {
		t.Errorf("expected ErrOutOfStock, got %v", err)
	}
}
