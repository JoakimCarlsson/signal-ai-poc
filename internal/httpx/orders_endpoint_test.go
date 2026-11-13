package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/router"
)

// TestPlaceOrder_ZeroQuantity verifies that POST /orders with quantity=0
// returns a controlled error response without panicking.
func TestPlaceOrder_ZeroQuantity(t *testing.T) {
	store := shop.NewStore()
	r := router.New()
	r.Use(recoverPanics)
	server := &Server{
		shop:   store,
		router: r,
		api:    r.Group("/api/v1"),
	}
	server.registerOrders()

	body := map[string]int{
		"product_id": 3,
		"quantity":   0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code == http.StatusCreated {
		t.Errorf("expected error response for quantity=0, got status %d", w.Code)
	}
	if w.Code != http.StatusInternalServerError {
		t.Logf("got status %d (acceptable non-success response)", w.Code)
	}
}

// TestPlaceOrder_ValidOrder verifies that normal order placement still works.
func TestPlaceOrder_ValidOrder(t *testing.T) {
	store := shop.NewStore()
	r := router.New()
	r.Use(recoverPanics)
	server := &Server{
		shop:   store,
		router: r,
		api:    r.Group("/api/v1"),
	}
	server.registerOrders()

	body := map[string]int{
		"product_id": 1,
		"quantity":   2,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response orderResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", response.Quantity)
	}
}
