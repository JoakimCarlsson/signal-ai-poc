package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/openapi"
	"github.com/joakimcarlsson/minmux/router"
)

// productResponse is one product.
type productResponse struct {
	ID    int    `json:"id"    desc:"The product."`
	Name  string `json:"name"  desc:"What it is called."`
	Price int    `json:"price" desc:"The price in öre."`
	Stock int    `json:"stock" desc:"Units left."`
}

// productParams carries the product to read.
type productParams struct {
	ID int `path:"id" desc:"The product."`
}

// toProduct maps a product to its response.
func toProduct(p shop.Product) productResponse {
	return productResponse{ID: p.ID, Name: p.Name, Price: p.Price, Stock: p.Stock}
}

// registerProducts registers the product endpoints.
func (s *Server) registerProducts() {
	s.api.Get("/products", func(c *router.Context) {
		products := s.shop.Products()
		out := make([]productResponse, 0, len(products))
		for _, p := range products {
			out = append(out, toProduct(p))
		}
		c.JSON(http.StatusOK, out)
	},
		openapi.Tags("Products"),
		openapi.Summary("List the products"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[[]productResponse](http.StatusOK, "Every product."),
	)

	s.api.Get("/products/{id}", func(c *router.Context, p productParams) {
		product, err := s.shop.Product(p.ID)
		if errors.Is(err, shop.ErrNotFound) {
			problem(c, router.NotFound("no such product"))
			return
		}
		if err != nil {
			slog.ErrorContext(c.Ctx(), "reading product", "id", p.ID, "error", err)
			problem(c, router.InternalServerError("the product could not be read"))
			return
		}
		c.JSON(http.StatusOK, toProduct(product))
	},
		openapi.Tags("Products"),
		openapi.Summary("Read a product"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[productResponse](http.StatusOK, "The product."),
		openapi.ReturnsBody[router.ProblemDetails](http.StatusNotFound, "No such product."),
		openapi.ReturnsBody[router.ProblemDetails](
			http.StatusInternalServerError, "The product could not be read."),
	)
}
