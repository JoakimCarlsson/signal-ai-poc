package httpx

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/router"
)

// TestProductMissing checks the existing JSON 404 response for absent product IDs.
func TestProductMissing(t *testing.T) {
	server := New(Config{}, Deps{Shop: shop.NewStore()})
	for _, id := range []int{42, -1, 0, 6} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, apiPrefix+"/products/"+strconv.Itoa(id), nil)
			server.router.ServeHTTP(response, request)
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404; body = %s", response.Code, response.Body.String())
			}
			mediaType, _, err := mime.ParseMediaType(response.Header().Get("Content-Type"))
			if err != nil || mediaType != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", response.Header().Get("Content-Type"))
			}
			var problem router.ProblemDetails
			if err := json.NewDecoder(response.Body).Decode(&problem); err != nil {
				t.Fatal(err)
			}
			if problem.Status != http.StatusNotFound || problem.Detail != "no such product" {
				t.Fatalf("problem = %+v, want status 404 and detail no such product", problem)
			}
		})
	}
}

// TestProductSeeded checks that individual responses match every field in the product list.
func TestProductSeeded(t *testing.T) {
	server := New(Config{}, Deps{Shop: shop.NewStore()})
	response := httptest.NewRecorder()
	server.router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, apiPrefix+"/products", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", response.Code)
	}
	var products []productResponse
	if err := json.NewDecoder(response.Body).Decode(&products); err != nil {
		t.Fatal(err)
	}
	if len(products) != 5 {
		t.Fatalf("product count = %d, want 5", len(products))
	}
	for _, want := range products {
		t.Run(strconv.Itoa(want.ID), func(t *testing.T) {
			response := httptest.NewRecorder()
			server.router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, apiPrefix+"/products/"+strconv.Itoa(want.ID), nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
			}
			var got productResponse
			if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("product = %+v, want %+v", got, want)
			}
		})
	}
}
