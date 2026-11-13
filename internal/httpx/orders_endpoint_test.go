package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
)

// TestReadOrders verifies guest, customer, and missing-order HTTP responses.
func TestReadOrders(t *testing.T) {
	server := New(Config{}, Deps{Shop: shop.NewStore()})
	for _, tt := range []struct {
		name   string
		id     int
		status int
		want   map[string]any
	}{
		{"guest", 3, http.StatusOK, map[string]any{"id": float64(3), "product_id": float64(4), "quantity": float64(1), "total": float64(129900)}},
		{"customer", 1, http.StatusOK, map[string]any{"id": float64(1), "customer_name": "Alice", "product_id": float64(1), "quantity": float64(1), "total": float64(79900)}},
		{"missing", 999999, http.StatusNotFound, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/orders/%d", tt.id), nil)
			response := httptest.NewRecorder()
			server.router.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d; body: %s", response.Code, tt.status, response.Body)
			}
			var got map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if tt.status == http.StatusNotFound {
				if got["detail"] != "no such order" {
					t.Fatalf("detail = %v, want no such order", got["detail"])
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("response = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestReadPlacedGuestOrder verifies that POST creates a guest order readable through GET.
func TestReadPlacedGuestOrder(t *testing.T) {
	server := New(Config{}, Deps{Shop: shop.NewStore()})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(`{"product_id":1,"quantity":1}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201; body: %s", response.Code, response.Body)
	}
	var created map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"id": float64(4), "product_id": float64(1), "quantity": float64(1), "total": float64(79900)}
	if !reflect.DeepEqual(created, want) {
		t.Fatalf("POST response = %v, want %v", created, want)
	}
	request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/orders/%.0f", created["id"]), nil)
	response = httptest.NewRecorder()
	server.router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body: %s", response.Code, response.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, created) {
		t.Fatalf("GET response = %v, want %v", got, created)
	}
}
