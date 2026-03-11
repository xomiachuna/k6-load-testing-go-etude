.PHONY: default
default: run-gow

.PHONY: run-signoz
run-sinoz:
	docker compose \
		-f signoz.docker-compose.yaml \
		--project-directory signoz/deploy/docker/ \
		up \
		-d --force-recreate

.PHONY: run-gow
run-gow:
	OTEL_EXPORTER_OTLP_ENDPOINT='' gow run .

.PHONY: run
run:
	OTEL_EXPORTER_OTLP_ENDPOINT='' go run .

.PHONY: test-load
test-load: test-smoke test-average test-stress

.PHONY: test-smoke
test-smoke:
	./scripts/performance-test.sh smoke

.PHONY: test-average
test-average:
	./scripts/performance-test.sh average

.PHONY: test-stress
test-stress:
	./scripts/performance-test.sh stress

.PHONY: test-breakpoint
test-breakpoint:
	./scripts/performance-test.sh breakpoint
