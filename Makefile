SHELL := /bin/bash

GO ?= go
NPM ?= npm

.PHONY: all build server web tidy test run clean dev

all: build

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

web:
	cd web && $(NPM) install && $(NPM) run build

server:
	$(GO) build -o bin/siphongear ./cmd/server

build: web server

run:
	./bin/siphongear --config config.yaml

dev:
	@echo "Backend: SIPHON_CONFIG=config.yaml go run ./cmd/server"
	@echo "Frontend: cd web && npm run dev"

clean:
	rm -rf bin web/dist data
