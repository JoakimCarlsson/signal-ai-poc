package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
)

// TestReadOrder checks order responses, including omitted guest names and missing IDs.
func TestReadOrder(t *testing.T) {
	for _, tt := range []struct {
		name   string
		path   string
		status int
		want   map[string]any
	}{
		{
			name: "guest", path: "/api/v1/orders/3", status: http.StatusOK,
			want: map[string]any{"id": float64(3), "product_id": float64(4), "quantity": float64(1), "total": float64(129900)},
		},
		{
			name: "Alice", path: "/api/v1/orders/1", status: http.StatusOK,
			want: map[string]any{"id": float64(1), "customer_name": "Alice", "product_id": float64(1), "quantity": float64(1), "total": float64(79900)},
		},
		{
			name: "Bob", path: "/api/v1/orders/2", status: http.StatusOK,
			want: map[string]any{"id": float64(2), "customer_name": "Bob", "product_id": float64(2), "quantity": float64(2), "total": float64(69800)},
		},
		{name: "missing", path: "/api/v1/orders/999999", status: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := New(Config{}, Deps{Shop: shop.NewStore()})
			response := httptest.NewRecorder()
			server.router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if response.Code != tt.status {
				t.Fatalf("GET %s status = %d, want %d; body: %s", tt.path, response.Code, tt.status, response.Body.String())
			}
			if tt.want != nil {
				var got map[string]any
				if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("GET %s body = %v, want %v", tt.path, got, tt.want)
				}
			}
		})
	}
}
