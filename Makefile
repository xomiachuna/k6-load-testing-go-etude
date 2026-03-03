.PHONY: default
default: test-load

.PHONY: test-load
test-load:
	docker run --rm -it \
		--name k6-load-test-etude \
		-v ./tests/k6/:/tests:ro \
		-w /tests \
		-e K6_WEB_DASHBOARD=true \
		-e K6_WEB_DASHBOARD_PERIOD=2s \
		--network host \
		grafana/k6 run test.js
