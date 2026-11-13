package httpx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/router"
)

// TestSearch checks invalid expressions and subsequent searches on a fresh store.
func TestSearch(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	s := New(Config{}, Deps{Shop: shop.NewStore()})

	invalid := httptest.NewRecorder()
	s.router.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, "/api/v1/search?q=%5Babc", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Errorf("invalid expression: status = %d, body = %s", invalid.Code, invalid.Body.String())
	}
	if !strings.HasPrefix(invalid.Header().Get("Content-Type"), "application/json") {
		t.Errorf("invalid expression: Content-Type = %q", invalid.Header().Get("Content-Type"))
	}
	var pd router.ProblemDetails
	if err := json.Unmarshal(invalid.Body.Bytes(), &pd); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if pd.Status != http.StatusBadRequest || !strings.Contains(pd.Detail, "invalid search regular expression") {
		t.Errorf("invalid expression: problem = %+v", pd)
	}
	if strings.Contains(logs.String(), "panic serving request") {
		t.Errorf("invalid expression logged a panic: %s", logs.String())
	}

	for _, tc := range []struct {
		query string
		names []string
	}{
		{query: "?q=%5EMON", names: []string{"Monitor"}},
		{query: "?q=zzzznomatch", names: []string{}},
		{query: "", names: []string{"Keyboard", "Mouse", "Monitor", "Headset", "Webcam"}},
		{query: "?q=", names: []string{"Keyboard", "Mouse", "Monitor", "Headset", "Webcam"}},
	} {
		t.Run(tc.query, func(t *testing.T) {
			response := httptest.NewRecorder()
			s.router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/search"+tc.query, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var products []productResponse
			if err := json.Unmarshal(response.Body.Bytes(), &products); err != nil {
				t.Fatalf("decode products: %v", err)
			}
			if products == nil {
				t.Fatal("response must be a JSON array, not null")
			}
			names := make([]string, 0, len(products))
			for _, product := range products {
				names = append(names, product.Name)
			}
			if !reflect.DeepEqual(names, tc.names) {
				t.Errorf("product names = %v, want %v", names, tc.names)
			}
		})
	}
}
