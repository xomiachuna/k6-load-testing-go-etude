.PHONY: default
default: test-load

.PHONY: test-load
test-load:
	docker run --rm -it \
		-u $$(id -u) \
		--name k6-load-test-etude \
		-v ./tests/k6/:/tests \
		-w /tests \
		-e K6_WEB_DASHBOARD=true \
		-e K6_WEB_DASHBOARD_PERIOD=5s \
		-e K6_WEB_DASHBOARD_EXPORT=/tests/report.html \
		--network host \
		grafana/k6 run test.js
