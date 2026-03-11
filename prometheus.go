package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func instrumentPrometheus(api *HttpAPI){
    reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
    api.Mux().Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
}

