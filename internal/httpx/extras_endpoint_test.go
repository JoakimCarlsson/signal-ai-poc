package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/router"
)

// TestSearchEndpoint_InvalidPattern is a regression test for the panic on
// invalid regexp in GET /api/v1/search. It verifies that malformed patterns
// return HTTP 400 instead of panicking.
func TestSearchEndpoint_InvalidPattern(t *testing.T) {
	s := New(Config{Addr: ":0"}, Deps{Shop: shop.NewStore()})

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
			req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q="+tt.pattern, nil)
			w := httptest.NewRecorder()

			s.router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("GET /api/v1/search?q=%s status = %d, want %d", tt.pattern, w.Code, http.StatusBadRequest)
			}

			var problem router.ProblemDetails
			if err := json.NewDecoder(w.Body).Decode(&problem); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if problem.Status != http.StatusBadRequest {
				t.Errorf("problem status = %d, want %d", problem.Status, http.StatusBadRequest)
			}

			if problem.Detail != "invalid search pattern" {
				t.Errorf("problem detail = %q, want %q", problem.Detail, "invalid search pattern")
			}
		})
	}
}

// TestSearchEndpoint_ValidPattern verifies that valid search requests still
// work after the fix.
func TestSearchEndpoint_ValidPattern(t *testing.T) {
	s := New(Config{Addr: ":0"}, Deps{Shop: shop.NewStore()})

	tests := []struct {
		name    string
		pattern string
		wantLen int
	}{
		{"exact match", "Mouse", 1},
		{"case insensitive", "mouse", 1},
		{"no match", "Laptop", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q="+tt.pattern, nil)
			w := httptest.NewRecorder()

			s.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("GET /api/v1/search?q=%s status = %d, want %d", tt.pattern, w.Code, http.StatusOK)
			}

			var products []productResponse
			if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if len(products) != tt.wantLen {
				t.Errorf("GET /api/v1/search?q=%s returned %d products, want %d", tt.pattern, len(products), tt.wantLen)
			}
		})
	}
}
