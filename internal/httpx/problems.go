package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/joakimcarlsson/minmux/router"
)

// problem writes pd as the response, with its status.
func problem(c *router.Context, pd *router.ProblemDetails) {
	c.JSON(pd.Status, pd)
}

// unavailable builds a 503, which minmux has no constructor for.
func unavailable(detail string) *router.ProblemDetails {
	return &router.ProblemDetails{
		Status: http.StatusServiceUnavailable,
		Title:  http.StatusText(http.StatusServiceUnavailable),
		Detail: detail,
	}
}

// writeProblem writes pd as the response outside a handler, such as from a
// middleware.
func writeProblem(w http.ResponseWriter, pd *router.ProblemDetails) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(pd.Status)
	if err := json.NewEncoder(w).Encode(pd); err != nil {
		slog.Error("writing problem", "error", err)
	}
}
