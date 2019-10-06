.PHONY: bq/setup
bq/setup: ## create BigQuery Dataset And Tables
	@./scripts/bq_setup