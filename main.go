package main

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func envOrDefault(env string, def string) string {
	value, exist := os.LookupEnv(env)
	if !exist {
		return def
	}
	return value
}

func main() {
	log.Info("开始启动 Clash Reporter")

	clashHost := envOrDefault("CLASH_HOST", "localhost:9090")
	clashToken := envOrDefault("CLASH_TOKEN", "")
	lokiAddr := envOrDefault("LOKI_ADDR", "http://localhost:3100/loki/api/v1/push")
	metricPort := envOrDefault("METRIC_PORT", "9001")

	log.Info("CLASH_HOST: ", clashHost)
	log.Info("CLASH_TOKEN: ", clashToken)
	log.Info("LOKI_ADDR: ", lokiAddr)
	log.Info("METRIC_PORT: ", metricPort)

	go handleReport(lokiAddr, clashHost, clashToken)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>Clash Reporter</title></head>
			<body>
			<h1>Clash Reporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Println("运行 Reporter 错误", err)
			os.Exit(1)
		}
	})
	log.Printf("监控 Metrics： http://localhost:%s/metrics", metricPort)
	log.Fatal(http.ListenAndServe(":"+metricPort, nil))
}
