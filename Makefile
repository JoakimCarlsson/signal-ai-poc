GO ?= go1.27.4
BASE ?= http://localhost:9901/api/v1

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: run build lint traffic

run:
	$(GO) run ./cmd/api

build:
	$(GO) build -o bin/api ./cmd/api

lint:
	$(GO) vet ./...

traffic:
	@while true; do \
	  curl -s -o /dev/null $(BASE)/products; \
	  curl -s -o /dev/null $(BASE)/products/$$((RANDOM % 6 + 1)); \
	  curl -s -o /dev/null $(BASE)/orders/$$((RANDOM % 4 + 1)); \
	  curl -s -o /dev/null $(BASE)/customers/$$((RANDOM % 3 + 1))/discount; \
	  curl -s -o /dev/null -X POST -H 'Content-Type: application/json' \
	    -d "{\"product_id\":$$((RANDOM % 5 + 1)),\"quantity\":$$((RANDOM % 4))}" $(BASE)/orders; \
	  sleep 1; \
	done
