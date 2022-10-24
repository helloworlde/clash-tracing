package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	clashMetricsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "clash",
		Name:      "metrics_counter",
		Help:      "Clash metrics counter for WebSocket data",
	}, []string{"metric_type"})
)

func init() {
	// 流量
	prometheus.MustRegister(clashMetricsCounter)
}

func UpdateMetricsCounter(typeName string) {
	clashMetricsCounter.WithLabelValues(typeName).Inc()
}
