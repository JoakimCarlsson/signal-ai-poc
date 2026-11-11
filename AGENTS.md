# AGENTS.md

Dummy shop API used as the target repository for signal-ai: it serves traffic, logs and traces through OpenTelemetry, and signal-ai files issues here from what it sees in Grafana. Go 1.27.4 on `minmux`, in-memory data, no database.

## Commands

The toolchain is installed as `go1.27.4`, not `go`. The Makefile uses it; pass `GO=go` to override.

```sh
cp .env.example .env
make run       # api on :9901, Scalar docs at /api/v1/docs
make traffic   # loops requests against every endpoint until Ctrl-C
make lint      # go vet ./...
make up        # api :9901 with its own Grafana :9931 and OTLP :9919
make down
```

`make up` starts the API (:9901) with its own LGTM stack (`grafana/otel-lgtm`): Grafana on :9931 (admin, `GRAFANA_PASSWORD`; anonymous access is off), OTLP on :9919 for `make run`, service name `shop`. Loki is not published; read it through Grafana's data source proxy, `<grafana>/api/datasources/proxy/uid/loki`, with a Grafana service account token as bearer. That is the endpoint and credential signal-ai is given as this project's log source.

The manual "Deploy" workflow runs the same compose file on the VPS's self-hosted runner (label `signal-ai-poc`), with `GRAFANA_PASSWORD` from the repository's secrets and `GRAFANA_URL` from its variables.

## Rules

- Domain logic lives in `internal/shop/` with no `net/http` or `minmux` imports; `internal/httpx/` only adapts it.
- Log with `slog` and its `Context` variants, so lines carry the trace ID.
- Every function and type gets a doc comment. Write no inline comments.
