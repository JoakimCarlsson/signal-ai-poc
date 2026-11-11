package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/openapi"
	"github.com/joakimcarlsson/minmux/router"
)

// orderResponse is one order.
type orderResponse struct {
	ID           int    `json:"id"                      desc:"The order."`
	CustomerName string `json:"customer_name,omitempty" desc:"Who placed it; empty for guests."`
	ProductID    int    `json:"product_id"              desc:"What was bought."`
	Quantity     int    `json:"quantity"                desc:"How many units."`
	Total        int    `json:"total"                   desc:"What it cost in öre."`
}

// orderParams carries the order to read.
type orderParams struct {
	ID int `path:"id" desc:"The order."`
}

// placeOrderRequest is what a guest orders.
type placeOrderRequest struct {
	ProductID int `json:"product_id" desc:"The product to buy."`
	Quantity  int `json:"quantity"   desc:"How many units."`
}

// placeOrderParams carries the order to place.
type placeOrderParams struct {
	Body placeOrderRequest `body:""`
}

// toOrder maps an order to its response.
func toOrder(o shop.OrderSummary) orderResponse {
	return orderResponse{
		ID:           o.ID,
		CustomerName: o.CustomerName,
		ProductID:    o.ProductID,
		Quantity:     o.Quantity,
		Total:        o.Total,
	}
}

// registerOrders registers the order endpoints.
func (s *Server) registerOrders() {
	s.api.Get("/orders/{id}", func(c *router.Context, p orderParams) {
		order, err := s.shop.Order(p.ID)
		if errors.Is(err, shop.ErrNotFound) {
			problem(c, router.NotFound("no such order"))
			return
		}
		if err != nil {
			slog.ErrorContext(c.Ctx(), "reading order", "id", p.ID, "error", err)
			problem(c, router.InternalServerError("the order could not be read"))
			return
		}
		c.JSON(http.StatusOK, toOrder(order))
	},
		openapi.Tags("Orders"),
		openapi.Summary("Read an order"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[orderResponse](http.StatusOK, "The order."),
		openapi.ReturnsBody[router.ProblemDetails](http.StatusNotFound, "No such order."),
		openapi.ReturnsBody[router.ProblemDetails](
			http.StatusInternalServerError, "The order could not be read."),
	)

	s.api.Post("/orders", func(c *router.Context, p placeOrderParams) {
		order, err := s.shop.PlaceOrder(p.Body.ProductID, p.Body.Quantity)
		switch {
		case errors.Is(err, shop.ErrNotFound):
			problem(c, router.NotFound("no such product"))
		case errors.Is(err, shop.ErrOutOfStock):
			slog.WarnContext(c.Ctx(), "order refused", "error", err)
			problem(c, router.Conflict("the product is out of stock"))
		case err != nil:
			slog.ErrorContext(c.Ctx(), "placing order", "error", err)
			problem(c, router.InternalServerError("the order could not be placed"))
		default:
			slog.InfoContext(c.Ctx(), "order placed",
				"order_id", order.ID, "total", order.Total)
			c.JSON(http.StatusCreated, toOrder(order))
		}
	},
		openapi.Tags("Orders"),
		openapi.Summary("Place an order as a guest"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[orderResponse](http.StatusCreated, "The order placed."),
		openapi.ReturnsBody[router.ProblemDetails](http.StatusNotFound, "No such product."),
		openapi.ReturnsBody[router.ProblemDetails](
			http.StatusConflict, "The product is out of stock."),
		openapi.ReturnsBody[router.ProblemDetails](
			http.StatusInternalServerError, "The order could not be placed."),
	)
}
