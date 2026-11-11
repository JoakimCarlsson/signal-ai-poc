GO ?= go1.27.4
BASE ?= http://localhost:9901/api/v1

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: run build lint traffic up down

run:
	$(GO) run ./cmd/api

build:
	$(GO) build -o bin/api ./cmd/api

lint:
	$(GO) vet ./...

up:
	docker compose up --build -d

down:
	docker compose down

traffic:
	@while true; do \
	  curl -s -o /dev/null $(BASE)/products; \
	  curl -s -o /dev/null $(BASE)/products/$$((RANDOM % 6 + 1)); \
	  curl -s -o /dev/null $(BASE)/orders/$$((RANDOM % 4 + 1)); \
	  curl -s -o /dev/null $(BASE)/customers/$$((RANDOM % 3 + 1))/discount; \
	  curl -s -o /dev/null "$(BASE)/search?q=mon"; \
	  curl -s -o /dev/null "$(BASE)/search?q=%28mon"; \
	  curl -s -o /dev/null $(BASE)/customers/$$((RANDOM % 3 + 1))/average-order; \
	  curl -s -o /dev/null "$(BASE)/shipping/estimate?postal_code=123%2045"; \
	  curl -s -o /dev/null "$(BASE)/shipping/estimate?postal_code=42"; \
	  curl -s -o /dev/null $(BASE)/products/$$((RANDOM % 4 + 1))/reviews; \
	  curl -s -o /dev/null -X POST -H 'Content-Type: application/json' \
	    -d "{\"product_id\":$$((RANDOM % 5 + 1)),\"quantity\":$$((RANDOM % 4))}" $(BASE)/orders; \
	  sleep 1; \
	done
