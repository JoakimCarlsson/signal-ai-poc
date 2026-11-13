package shop

import (
	"errors"
	"strconv"
	"testing"
)

// TestProductSeeded checks that every seeded ID returns the complete matching product.
func TestProductSeeded(t *testing.T) {
	store := NewStore()
	for _, want := range store.Products() {
		t.Run(strconv.Itoa(want.ID), func(t *testing.T) {
			got, err := store.Product(want.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("Product(%d) = %+v, want %+v", want.ID, got, want)
			}
		})
	}
}

// TestProductMissing checks that absent IDs return ErrNotFound without panicking.
func TestProductMissing(t *testing.T) {
	store := NewStore()
	for _, id := range []int{-1, 0, 6, 42} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			if _, err := store.Product(id); !errors.Is(err, ErrNotFound) {
				t.Fatalf("Product(%d) error = %v, want ErrNotFound", id, err)
			}
		})
	}
}

// TestProductReordered checks that identity is independent of slice position.
func TestProductReordered(t *testing.T) {
	store := NewStore()
	store.products[0], store.products[4] = store.products[4], store.products[0]
	for _, want := range store.Products() {
		t.Run(strconv.Itoa(want.ID), func(t *testing.T) {
			got, err := store.Product(want.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("Product(%d) = %+v, want %+v", want.ID, got, want)
			}
		})
	}
}
