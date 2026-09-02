GO ?= go
BINARY ?= bin/ingest
RESET := \033[0m
BOLD := \033[1m
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m

.DEFAULT_GOAL := help

.PHONY: help build run test fmt vet tidy check clean

help:
	@printf '%b\n' \
		'$(BOLD)$(CYAN)Usage: make <target>$(RESET)' \
		'' \
		'$(BOLD)Available targets:$(RESET)' \
		'  $(GREEN)help$(RESET)   Show this help message' \
		'  $(GREEN)run$(RESET)    Run the ingestion job' \
		'  $(GREEN)build$(RESET)  Build the ingestion binary' \
		'  $(GREEN)test$(RESET)   Run tests' \
		'  $(GREEN)fmt$(RESET)    Format Go files' \
		'  $(GREEN)vet$(RESET)    Run static analysis' \
		'  $(GREEN)tidy$(RESET)   Tidy Go dependencies' \
		'  $(YELLOW)check$(RESET)  Format, test, and vet the project' \
		'  $(GREEN)clean$(RESET)  Remove build artifacts'

build:
	$(GO) build -o $(BINARY) ./cmd/ingest

run:
	$(GO) run ./cmd/ingest

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

check: fmt test vet

clean:
	rm -f $(BINARY)
