package shop

import (
	"errors"
	"regexp/syntax"
	"testing"
)

// TestSearchInvalidExpression verifies malformed patterns return compilation errors.
func TestSearchInvalidExpression(t *testing.T) {
	store := NewStore()
	products, err := store.Search("[abc")
	var syntaxErr *syntax.Error
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("Search returned error %v, want a regular-expression syntax error", err)
	}
	if products != nil {
		t.Errorf("Search returned products with an invalid expression: %v", products)
	}
	products, err = store.Search("^MON")
	if err != nil {
		t.Fatalf("valid search after invalid expression: %v", err)
	}
	if len(products) != 1 || products[0].Name != "Monitor" {
		t.Errorf("Search(^MON) = %v, want only Monitor", products)
	}
}
