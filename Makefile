.PHONY: default
default: run

.PHONY: run-signoz
run-sinoz:
	docker compose \
		-f signoz.docker-compose.yaml \
		--project-directory signoz/deploy/docker/ \
		up \
		-d --force-recreate

.PHONY: run
run:
	docker compose up api --force-recreate --build --watch

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
