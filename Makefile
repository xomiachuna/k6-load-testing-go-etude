.PHONY: default
default: run-gow

.PHONY: run-gow
run-gow:
	OTEL_EXPORTER_OTLP_ENDPOINT='' gow run .

.PHONY: test-load
test-load: test-smoke test-average test-stress

.PHONY: test-smoke
test-smoke:
	docker run --rm -it \
		-u $$(id -u) \
		--name k6-load-test-etude \
		-v ./tests/k6/:/tests \
		-w /tests \
		-e K6_WEB_DASHBOARD=true \
		-e K6_WEB_DASHBOARD_PERIOD=5s \
		-e K6_WEB_DASHBOARD_EXPORT=/tests/report-smoke.html \
		-e TEST_KIND=smoke \
		--network host \
		grafana/k6 run test.js \
			--summary-mode=full

.PHONY: test-average
test-average:
	docker run --rm -it \
		-u $$(id -u) \
		--name k6-load-test-etude \
		-v ./tests/k6/:/tests \
		-w /tests \
		-e K6_WEB_DASHBOARD=true \
		-e K6_WEB_DASHBOARD_PERIOD=5s \
		-e K6_WEB_DASHBOARD_EXPORT=/tests/report-average.html \
		-e TEST_KIND=average \
		--network host \
		grafana/k6 run test.js \
			--summary-mode=full

.PHONY: test-stress
test-stress:
	docker run --rm -it \
		-u $$(id -u) \
		--name k6-load-test-etude \
		-v ./tests/k6/:/tests \
		-w /tests \
		-e K6_WEB_DASHBOARD=true \
		-e K6_WEB_DASHBOARD_PERIOD=5s \
		-e K6_WEB_DASHBOARD_EXPORT=/tests/report-stress.html \
		-e TEST_KIND=stress \
		--network host \
		grafana/k6 run test.js \
			--summary-mode=full

.PHONY: test-breakpoint
test-breakpoint:
	docker run --rm -it \
		-u $$(id -u) \
		--name k6-load-test-etude \
		-v ./tests/k6/:/tests \
		-w /tests \
		-e K6_WEB_DASHBOARD=true \
		-e K6_WEB_DASHBOARD_PERIOD=10s \
		-e K6_WEB_DASHBOARD_EXPORT=/tests/report-breakpoint.html \
		-e TEST_KIND=breakpoint \
		--network host \
		grafana/k6 run test.js \
			--summary-mode=full
