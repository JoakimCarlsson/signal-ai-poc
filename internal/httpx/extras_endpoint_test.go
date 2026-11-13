package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
)

// TestAverageOrderEndpoint checks the average-order response for each customer case.
func TestAverageOrderEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name   string
		id     string
		status int
		body   string
	}{
		{name: "no orders", id: "3", status: http.StatusOK, body: `{"average":0}`},
		{name: "Alice", id: "1", status: http.StatusOK, body: `{"average":79900}`},
		{name: "Bob", id: "2", status: http.StatusOK, body: `{"average":69800}`},
		{name: "unknown customer", id: "999", status: http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := New(Config{}, Deps{Shop: shop.NewStore()})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/"+tc.id+"/average-order", nil)
			res := httptest.NewRecorder()
			s.router.ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", res.Code, tc.status, res.Body.String())
			}
			if tc.status == http.StatusOK {
				if got := strings.TrimSpace(res.Body.String()); got != tc.body {
					t.Fatalf("body = %s, want %s", got, tc.body)
				}
			} else {
				var body struct {
					Detail string `json:"detail"`
				}
				if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Detail != "no such customer" {
					t.Fatalf("detail = %q, want %q", body.Detail, "no such customer")
				}
			}
		})
	}
}
