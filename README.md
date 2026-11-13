# signal-ai-poc

Guest orders, including seeded order 3 and orders created with `POST /api/v1/orders`,
can be read with `GET /api/v1/orders/{id}`. Their responses omit `customer_name`;
named orders retain it, and nonexistent order IDs return HTTP 404.

Run regression tests with `go1.27.4 test ./...` and vet with `make lint`.
If the Go 1.27.4 binary is installed as `go`, use `go test ./...` and
`make lint GO=go` instead.