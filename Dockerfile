FROM golang:1.27.4-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w" -o /out/api ./cmd/api

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 -h /app shop

WORKDIR /app
COPY --from=build /out/api /usr/local/bin/

ENV ADDR=:9901

USER shop
EXPOSE 9901

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:9901/api/v1/health >/dev/null || exit 1

ENTRYPOINT ["api"]
