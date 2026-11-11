// Package otel is the observability adapter: structured logging, traces and
// metrics.
//
// Setup installs the OpenTelemetry providers and SetupLogger installs the
// slog handler that writes to stdout and exports over OTLP at once. Both are
// called once, from a binary's main, before anything else runs. Everything
// else in the process then logs through the slog default logger, and a log
// line made with a request's context carries that request's trace and span
// ids.
//
// Export is off unless OTEL_EXPORTER_OTLP_ENDPOINT is set. Without it the
// providers are still installed — instrumented code keeps working, spans and
// metrics are simply dropped — and logs still go to stdout.
package otel
