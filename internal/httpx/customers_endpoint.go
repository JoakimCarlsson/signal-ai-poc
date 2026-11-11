package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/openapi"
	"github.com/joakimcarlsson/minmux/router"
)

// discountResponse is what a customer's discount code gives.
type discountResponse struct {
	Percent int `json:"percent" desc:"The percentage off; zero without a code."`
}

// customerParams carries the customer to read.
type customerParams struct {
	ID int `path:"id" desc:"The customer."`
}

// registerCustomers registers the customer endpoints.
func (s *Server) registerCustomers() {
	s.api.Get("/customers/{id}/discount", func(c *router.Context, p customerParams) {
		percent, err := s.shop.Discount(p.ID)
		if errors.Is(err, shop.ErrNotFound) {
			problem(c, router.NotFound("no such customer"))
			return
		}
		if err != nil {
			slog.ErrorContext(c.Ctx(), "reading discount",
				"customer_id", p.ID, "error", err)
			problem(c, router.InternalServerError("the discount could not be read"))
			return
		}
		c.JSON(http.StatusOK, discountResponse{Percent: percent})
	},
		openapi.Tags("Customers"),
		openapi.Summary("Read a customer's discount"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[discountResponse](http.StatusOK, "The discount."),
		openapi.ReturnsBody[router.ProblemDetails](http.StatusNotFound, "No such customer."),
		openapi.ReturnsBody[router.ProblemDetails](
			http.StatusInternalServerError, "The discount could not be read."),
	)
}
