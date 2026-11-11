# AGENTS.md

Dummy shop API used as the target repository for signal-ai: it serves traffic, logs and traces through OpenTelemetry, and signal-ai files issues here from what it sees in Grafana. Go 1.27.4 on `minmux`, in-memory data, no database.

## Commands

The toolchain is installed as `go1.27.4`, not `go`. The Makefile uses it; pass `GO=go` to override.

```sh
cp .env.example .env
make run       # api on :9901, Scalar docs at /api/v1/docs
make traffic   # loops requests against every endpoint until Ctrl-C
make lint      # go vet ./...
```

Telemetry goes to the LGTM stack of the signal-ai repository (`make up` there): OTLP on :9918, Grafana on :9930 with service name `shop`.

The manual "Deploy" workflow runs `docker compose up` on the VPS's self-hosted runner (label `signal-ai-poc`), serving on :9901 and sending telemetry to the signal-ai stack on the same host.

## Rules

- Domain logic lives in `internal/shop/` with no `net/http` or `minmux` imports; `internal/httpx/` only adapts it.
- Log with `slog` and its `Context` variants, so lines carry the trace ID.
- Every function and type gets a doc comment. Write no inline comments.
