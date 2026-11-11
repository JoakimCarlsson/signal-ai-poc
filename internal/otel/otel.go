package otel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"strings"

	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config is what Setup needs to identify the process and find a collector.
type Config struct {
	// ServiceName and ServiceVersion name this process in telemetry.
	ServiceName    string
	ServiceVersion string
	// OTLPExporterEndpoint is the collector or gateway base URL. Empty
	// disables export.
	OTLPExporterEndpoint string
	// OTLPExporterHeaders is "k=v,k=v" — auth for the endpoint above.
	OTLPExporterHeaders string
}

// otlpTarget is the resolved OTLP destination: a full gateway URL (e.g. Grafana
// Cloud, or a collector at http://host:port) plus any auth headers. The scheme
// of the URL decides TLS, so http:// is automatically insecure.
type otlpTarget struct {
	endpoint string
	headers  map[string]string
}

// newInstanceID identifies this running process in telemetry. It uses the
// container hostname (each instance gets a distinct one) so concurrent
// instances show up as distinct service.instance.id values, and falls back to
// a random id when no hostname is available.
func newInstanceID() string {
	if host, err := os.Hostname(); err == nil && host != "" {
		return host
	}
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// resolveTarget reads the OTLP destination from cfg. The bool is false when
// no endpoint is set, which means export is off.
func resolveTarget(cfg Config) (otlpTarget, bool) {
	if cfg.OTLPExporterEndpoint == "" {
		return otlpTarget{}, false
	}
	return otlpTarget{
		endpoint: strings.TrimRight(cfg.OTLPExporterEndpoint, "/"),
		headers:  parseHeaders(cfg.OTLPExporterHeaders),
	}, true
}

// parseHeaders reads the "k=v,k=v" header form. A pair without an "=" is
// skipped rather than failing the boot: a malformed header is not worth
// refusing to start over.
func parseHeaders(raw string) map[string]string {
	headers := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if k != "" {
			headers[k] = v
		}
	}
	return headers
}

// Setup installs the trace, metric and log providers as the global ones and
// returns the func that flushes and closes them. Call it once from main, and
// defer the returned shutdown: batched telemetry is lost if the process exits
// without it.
func Setup(
	ctx context.Context,
	cfg Config,
) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.ServiceInstanceID(newInstanceID()),
		),
	)
	if err != nil {
		handleErr(err)
		return
	}

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	target, useOTLP := resolveTarget(cfg)
	if useOTLP {
		slog.InfoContext(ctx, "otel export enabled",
			"endpoint", target.endpoint,
			"auth", len(target.headers) > 0,
		)
	} else {
		slog.WarnContext(
			ctx,
			"otel export disabled: OTEL_EXPORTER_OTLP_ENDPOINT not set",
		)
	}

	tracerProvider, err := newTracerProvider(ctx, res, target, useOTLP)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, res, target, useOTLP)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	if useOTLP {
		if startErr := otelruntime.Start(
			otelruntime.WithMeterProvider(meterProvider),
		); startErr != nil {
			handleErr(startErr)
			return
		}
	}

	loggerProvider, err := newLoggerProvider(ctx, res, target, useOTLP)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	otel.SetLoggerProvider(loggerProvider)

	return shutdown, nil
}

// newTracerProvider builds the trace provider, exporting over OTLP when
// useOTLP is set and otherwise only recording against res.
func newTracerProvider(
	ctx context.Context,
	res *resource.Resource,
	target otlpTarget,
	useOTLP bool,
) (*trace.TracerProvider, error) {
	if !useOTLP {
		return trace.NewTracerProvider(trace.WithResource(res)), nil
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(target.endpoint + "/v1/traces"),
	}
	if len(target.headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(target.headers))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	), nil
}

// newMeterProvider builds the metric provider, exporting over OTLP when
// useOTLP is set and otherwise only recording against res.
func newMeterProvider(
	ctx context.Context,
	res *resource.Resource,
	target otlpTarget,
	useOTLP bool,
) (*metric.MeterProvider, error) {
	if !useOTLP {
		return metric.NewMeterProvider(metric.WithResource(res)), nil
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(target.endpoint + "/v1/metrics"),
	}
	if len(target.headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(target.headers))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	), nil
}

// newLoggerProvider builds the log provider, exporting over OTLP when
// useOTLP is set and otherwise only recording against res.
func newLoggerProvider(
	ctx context.Context,
	res *resource.Resource,
	target otlpTarget,
	useOTLP bool,
) (*log.LoggerProvider, error) {
	if !useOTLP {
		return log.NewLoggerProvider(log.WithResource(res)), nil
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpointURL(target.endpoint + "/v1/logs"),
	}
	if len(target.headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(target.headers))
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(res),
	), nil
}
