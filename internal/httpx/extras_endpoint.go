package httpx

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/openapi"
	"github.com/joakimcarlsson/minmux/router"
)

// searchParams carries what to search for.
type searchParams struct {
	Q string `query:"q" desc:"A regular expression matched against product names."`
}

// averageResponse is a customer's average order.
type averageResponse struct {
	Average int `json:"average" desc:"The average order total in öre."`
}

// shippingParams carries where to ship.
type shippingParams struct {
	PostalCode string `query:"postal_code" desc:"A Swedish postal code, such as 123 45."`
}

// shippingResponse is how long delivery takes.
type shippingResponse struct {
	Days int `json:"days" desc:"Working days until delivery."`
}

// reviewResponse is one review.
type reviewResponse struct {
	Author  string `json:"author"  desc:"Who wrote it."`
	Rating  int    `json:"rating"  desc:"From 1 to 5."`
	Written string `json:"written" desc:"When, YYYY-MM-DD."`
}

// reviewsParams carries the product whose reviews to read.
type reviewsParams struct {
	ID int `path:"id" desc:"The product."`
}

// registerExtras registers search, average orders, shipping estimates and
// reviews.
func (s *Server) registerExtras() {
	s.api.Get("/search", func(c *router.Context, p searchParams) {
		found := s.shop.Search(p.Q)
		out := make([]productResponse, 0, len(found))
		for _, pr := range found {
			out = append(out, toProduct(pr))
		}
		c.JSON(http.StatusOK, out)
	},
		openapi.Tags("Products"),
		openapi.Summary("Search products by name"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[[]productResponse](http.StatusOK, "The matches."),
	)

	s.api.Get("/customers/{id}/average-order", func(c *router.Context, p customerParams) {
		avg, err := s.shop.AverageOrder(p.ID)
		if errors.Is(err, shop.ErrNotFound) {
			problem(c, router.NotFound("no such customer"))
			return
		}
		c.JSON(http.StatusOK, averageResponse{Average: avg})
	},
		openapi.Tags("Customers"),
		openapi.Summary("Read a customer's average order"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[averageResponse](http.StatusOK, "The average."),
		openapi.ReturnsBody[router.ProblemDetails](http.StatusNotFound, "No such customer."),
	)

	s.api.Get("/shipping/estimate", func(c *router.Context, p shippingParams) {
		c.JSON(http.StatusOK, shippingResponse{Days: shop.ShippingEstimate(p.PostalCode)})
	},
		openapi.Tags("Shipping"),
		openapi.Summary("Estimate delivery time to a postal code"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[shippingResponse](http.StatusOK, "The estimate."),
	)

	s.api.Get("/products/{id}/reviews", func(c *router.Context, p reviewsParams) {
		reviews, err := s.shop.Reviews(p.ID)
		if err != nil {
			slog.ErrorContext(c.Ctx(), "reading reviews", "product_id", p.ID, "error", err)
			problem(c, router.InternalServerError("the reviews could not be read"))
			return
		}
		out := make([]reviewResponse, 0, len(reviews))
		for _, r := range reviews {
			out = append(out, reviewResponse{
				Author:  r.Author,
				Rating:  r.Rating,
				Written: r.Written.Format(time.DateOnly),
			})
		}
		c.JSON(http.StatusOK, out)
	},
		openapi.Tags("Products"),
		openapi.Summary("Read a product's reviews"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[[]reviewResponse](http.StatusOK, "The reviews, newest first."),
		openapi.ReturnsBody[router.ProblemDetails](
			http.StatusInternalServerError, "The reviews could not be read."),
	)
}
