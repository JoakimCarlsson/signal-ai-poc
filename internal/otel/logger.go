package otel

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/trace"
)

// SetupLogger installs the process-wide slog logger: every record goes both to
// stdout (JSON, or text when format is "text") and, through the otelslog
// bridge, to the logger provider Setup installed. Call it from main, after
// Setup, so log lines made before it are not lost to a provider that is not
// there yet.
//
// scope names the instrumentation scope on exported records; the service name
// is the obvious value.
func SetupLogger(scope, level, format string) {
	slog.SetDefault(slog.New(newHandler(os.Stdout, scope, level, format)))
}

// newHandler builds the handler SetupLogger installs. It is separate so tests
// can hold one without touching the default logger.
func newHandler(w io.Writer, scope, level, format string) slog.Handler {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var base slog.Handler
	if strings.EqualFold(format, "text") {
		base = slog.NewTextHandler(w, opts)
	} else {
		base = slog.NewJSONHandler(w, opts)
	}

	return &multiHandler{handlers: []slog.Handler{
		&traceContextHandler{handler: base},
		otelslog.NewHandler(scope),
	}}
}

// parseLevel reads the configured level, falling back to info rather than
// failing: an unreadable level should not keep the process from starting.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// traceContextHandler stamps the trace and span ids of the record's context
// onto the stdout line, so a log line and the span it happened inside can be
// lined up. The OTLP side needs no such help — it carries the context itself.
type traceContextHandler struct {
	handler slog.Handler
}

// Enabled reports whether the wrapped handler handles records at level.
func (h *traceContextHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle adds the trace and span ids of ctx to the record, when ctx carries
// a valid span, and passes it to the wrapped handler.
func (h *traceContextHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs returns a handler whose wrapped handler carries attrs.
func (h *traceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceContextHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup returns a handler whose wrapped handler opens the named group.
func (h *traceContextHandler) WithGroup(name string) slog.Handler {
	return &traceContextHandler{handler: h.handler.WithGroup(name)}
}

// multiHandler fans one record out to several handlers.
type multiHandler struct {
	handlers []slog.Handler
}

// Enabled reports whether any wrapped handler handles records at level.
func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle dispatches each record to every wrapped handler, so a single slog
// call can write to stdout and export over OTLP at once. The record is cloned
// per handler so one handler's mutations can't leak into the next.
func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs returns a handler that applies attrs to every wrapped handler.
func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

// WithGroup returns a handler that opens the named group on every wrapped
// handler.
func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
