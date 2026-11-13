package httpx

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/router"
)

// TestShippingEstimateRejectsShortPostalCodes checks that invalid input produces a JSON problem instead of a recovered panic.
func TestShippingEstimateRejectsShortPostalCodes(t *testing.T) {
	s := New(Config{}, Deps{Shop: shop.NewStore()})
	for _, query := range []string{
		"?postal_code=42",
		"",
		"?postal_code=",
		"?postal_code=%20%20",
		"?postal_code=4",
		"?postal_code=4%202",
	} {
		t.Run(query, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, apiPrefix+"/shipping/estimate"+query, nil)
			s.router.ServeHTTP(w, r)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body = %s", w.Code, w.Body.String())
			}
			mediaType, _, err := mime.ParseMediaType(w.Header().Get("Content-Type"))
			if err != nil || mediaType != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", w.Header().Get("Content-Type"))
			}
			var pd router.ProblemDetails
			if err := json.Unmarshal(w.Body.Bytes(), &pd); err != nil {
				t.Fatal(err)
			}
			if pd.Status != http.StatusBadRequest || !strings.Contains(pd.Detail, "postal code") || !strings.Contains(pd.Detail, "too short") {
				t.Errorf("problem = %+v, want status 400 explaining that the postal code is too short", pd)
			}
		})
	}
}

// TestShippingEstimateDeliveryDays checks normalization and the existing regional delivery estimates.
func TestShippingEstimateDeliveryDays(t *testing.T) {
	s := New(Config{}, Deps{Shop: shop.NewStore()})
	for _, tc := range []struct {
		postalCode string
		days       int
	}{
		{postalCode: "12345", days: 1},
		{postalCode: "123%2045", days: 1},
		{postalCode: "20000", days: 2},
		{postalCode: "60000", days: 4},
	} {
		t.Run(tc.postalCode, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, apiPrefix+"/shipping/estimate?postal_code="+tc.postalCode, nil)
			s.router.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
			}
			var body shippingResponse
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Days != tc.days {
				t.Errorf("days = %d, want %d", body.Days, tc.days)
			}
		})
	}
}
