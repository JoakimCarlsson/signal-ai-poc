package shop

import (
	"errors"
	"strconv"
	"testing"
)

// TestProductOutOfRange checks that invalid catalogue indices return ErrNotFound.
func TestProductOutOfRange(t *testing.T) {
	store := NewStore()
	for _, id := range []int{-1, len(store.products), 42} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			product, err := store.Product(id)
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("Product(%d) error = %v, want ErrNotFound", id, err)
			}
			if product != (Product{}) {
				t.Fatalf("Product(%d) = %+v, want zero product", id, product)
			}
		})
	}
}
