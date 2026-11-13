package shop

import (
	"errors"
	"testing"
)

// TestSearch_InvalidPattern verifies that Store.Search returns
// ErrInvalidPattern when given a malformed regular expression, instead of
// panicking.
func TestSearch_InvalidPattern(t *testing.T) {
	s := NewStore()

	tests := []struct {
		name    string
		pattern string
	}{
		{"unclosed bracket", "[abc"},
		{"unclosed paren", "(abc"},
		{"invalid repetition", "*abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products, err := s.Search(tt.pattern)
			if !errors.Is(err, ErrInvalidPattern) {
				t.Errorf("Search(%q) error = %v, want ErrInvalidPattern", tt.pattern, err)
			}
			if products != nil {
				t.Errorf("Search(%q) returned products %v, want nil", tt.pattern, products)
			}
		})
	}
}

// TestSearch_ValidPattern verifies that Store.Search still works correctly
// with valid regular expressions.
func TestSearch_ValidPattern(t *testing.T) {
	s := NewStore()

	tests := []struct {
		name     string
		pattern  string
		wantLen  int
		wantName string
	}{
		{"exact match", "Mouse", 1, "Mouse"},
		{"case insensitive", "mouse", 1, "Mouse"},
		{"prefix", "^Key", 1, "Keyboard"},
		{"contains", "ead", 1, "Headset"},
		{"no match", "Laptop", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products, err := s.Search(tt.pattern)
			if err != nil {
				t.Fatalf("Search(%q) unexpected error: %v", tt.pattern, err)
			}
			if len(products) != tt.wantLen {
				t.Errorf("Search(%q) returned %d products, want %d", tt.pattern, len(products), tt.wantLen)
			}
			if tt.wantLen > 0 && products[0].Name != tt.wantName {
				t.Errorf("Search(%q) first product = %q, want %q", tt.pattern, products[0].Name, tt.wantName)
			}
		})
	}
}
