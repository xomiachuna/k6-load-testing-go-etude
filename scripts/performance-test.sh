#!/bin/bash

TEST_KIND=${1:?"test_kind required (smoke|average|stress|breakpoint)"}
docker run --rm -it \
    -u $(id -u) \
    --name k6-load-test-etude \
    -v ./tests/k6/:/tests \
    -w /tests \
    -e K6_WEB_DASHBOARD=true \
    -e K6_WEB_DASHBOARD_PERIOD=10s \
    -e K6_WEB_DASHBOARD_EXPORT=/tests/report-$TEST_KIND.html \
    -e TEST_KIND=$TEST_KIND \
    --network host \
    grafana/k6 run test.js \
        --summary-mode=full
