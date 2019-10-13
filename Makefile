.PHONY: build
build:
	go build main.go

.PHONY: test
test:
	go test -v

.PHONY: bq/setup
bq/setup: ## create BigQuery Dataset And Tables
	./scripts/bq_clean
	./scripts/bq_setup