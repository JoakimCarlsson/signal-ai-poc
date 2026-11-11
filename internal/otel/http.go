package otel

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Transport wraps the default HTTP transport with OpenTelemetry client
// instrumentation. Every outgoing request gets a client span and feeds the
// http.client.* metrics, tagged with:
//
//   - peer.service=<name> — the logical upstream (e.g. "azure-openai"), so
//     traffic to each can be grouped even when it spans several hostnames.
//   - http.route — the request path with high-cardinality segments (numeric
//     ids, UUIDs) collapsed to {id}, so per-route metrics stay bounded.
//
// Pass the result as the Transport of an http.Client:
//
//	&http.Client{Transport: otel.Transport("azure-openai")}
func Transport(name string) http.RoundTripper {
	peer := attribute.String("peer.service", name)

	instrumented := otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithSpanOptions(trace.WithAttributes(peer)),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + normalizeRoute(r.URL.Path)
		}),
	)

	return &peerTransport{base: instrumented, peer: name}
}

// peerTransport seeds the request context with a Labeler so the wrapped
// otelhttp transport attaches the peer.service and http.route attributes to
// its metrics.
type peerTransport struct {
	base http.RoundTripper
	peer string
}

// RoundTrip seeds a fresh Labeler scoped to this outgoing request. Reusing
// the one already on the context (e.g. the inbound server span's Labeler)
// would leak peer.service and http.route onto the server metric and across
// sibling upstream calls within the same request.
func (t *peerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	labeler := &otelhttp.Labeler{}
	labeler.Add(
		attribute.String("peer.service", t.peer),
		attribute.String("http.route", normalizeRoute(req.URL.Path)),
	)
	req = req.WithContext(otelhttp.ContextWithLabeler(req.Context(), labeler))
	return t.base.RoundTrip(req)
}

// normalizeRoute collapses variable path segments (numeric ids, UUIDs, other
// long opaque tokens) to {id} so the http.route metric label has bounded
// cardinality — "/customers/123/keys" and "/customers/456/keys" both become
// "/customers/{id}/keys".
func normalizeRoute(path string) string {
	if path == "" {
		return "/"
	}
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if isVariableSegment(s) {
			segs[i] = "{id}"
		}
	}
	return strings.Join(segs, "/")
}

// isVariableSegment reports whether a path segment looks like an id: all
// digits, or a hex-and-dash token of at least 16 characters with a digit in
// it, such as a UUID.
func isVariableSegment(s string) bool {
	if len(s) == 0 {
		return false
	}

	digits, hexish := 0, 0
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits++
			hexish++
		case (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == '-':
			hexish++
		}
	}

	if digits == len(s) {
		return true
	}
	if hexish == len(s) && digits > 0 && len(s) >= 16 {
		return true
	}
	return false
}
