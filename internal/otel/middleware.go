package otel

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Tracing wraps the handler in an OpenTelemetry HTTP server span. It goes
// outermost, so everything below it — Logging included — runs inside the span.
func Tracing(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http-server")
}

// responseWriter remembers what the handler answered, which net/http does not
// expose once the response is written.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// WriteHeader records the status before passing it on.
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write counts the bytes written, and records a 200 status when the handler
// never set one.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Unwrap exposes the wrapped writer so http.ResponseController can reach the
// optional interfaces (Flusher, Hijacker) this type does not implement.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Logging logs one line per completed request at info level (4xx at warn,
// 5xx at error). It runs inside the tracing middleware, so each line carries
// the request's trace_id/span_id via the slog handler installed in SetupLogger.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		level := slog.LevelInfo
		switch {
		case wrapped.status >= 500:
			level = slog.LevelError
		case wrapped.status >= 400:
			level = slog.LevelWarn
		}

		slog.LogAttrs(r.Context(), level, "http request",
			slog.String("subsystem", "http"),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.status),
			slog.Int("bytes", wrapped.bytes),
			slog.Duration("duration", time.Since(start)),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("referer", r.Referer()),
		)
	})
}
