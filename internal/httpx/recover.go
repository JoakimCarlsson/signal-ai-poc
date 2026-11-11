package httpx

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/joakimcarlsson/minmux/router"
)

// recoverPanics turns a panicking handler into a 500 and logs the panic
// with its stack at error level, in the request's context so the line
// carries its trace.
func recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			slog.ErrorContext(r.Context(), "panic serving request",
				"method", r.Method,
				"path", r.URL.Path,
				"panic", fmt.Sprint(rec),
				"stack", string(debug.Stack()),
			)
			writeProblem(w, router.InternalServerError("something went wrong"))
		}()
		next.ServeHTTP(w, r)
	})
}
