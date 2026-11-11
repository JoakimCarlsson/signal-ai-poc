package httpx

import (
	"net/http"

	"github.com/joakimcarlsson/minmux/openapi"
	"github.com/joakimcarlsson/minmux/router"
)

// healthResponse is the body of a passing health check.
type healthResponse struct {
	Status string `json:"status" desc:"Always ok."`
}

// registerHealth registers the health endpoint.
func (s *Server) registerHealth() {
	s.api.Get("/health", func(c *router.Context) {
		c.JSON(http.StatusOK, healthResponse{Status: "ok"})
	},
		openapi.Tags("Health"),
		openapi.Summary("Check that the API is up"),
		openapi.NoSecurity(),
		openapi.ReturnsBody[healthResponse](http.StatusOK, "The API is up."),
	)
}
