package main

import (
	"fmt"
	"net/http"
	"time"

	"flag"

	klog "k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	sb "github.com/marcinbudny/servicebus_exporter/client"
	"github.com/marcinbudny/servicebus_exporter/collector"
)

type config struct {
	timeout          time.Duration
	port             uint
	connectionString string
}

func readAndValidateConfig() config {
	var result config

	flag.StringVar(&result.connectionString, "connection-string", "", "Azure ServiceBus connection string")
	flag.UintVar(&result.port, "port", 9580, "Port to expose scraping endpoint on")
	flag.DurationVar(&result.timeout, "timeout", time.Second*30, "Timeout for scrape")

	flag.Parse()

	if result.connectionString == "" {
		klog.Fatal("Azure ServiceBus connection string not provided")
	}

	return result
}

func configureRoutes() {
	var landingPage = []byte(`<html>
		<head><title>Azure ServiceBus exporter for Prometheus</title></head>
		<body>
		<h1>Azure ServiceBus exporter for Prometheus</h1>
		<p><a href='/metrics'>Metrics</a></p>
		</body>
		</html>
		`)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage) // nolint: errcheck
	})

	http.Handle("/metrics", promhttp.Handler())
}

func startHTTPServer(config config) {
	listenAddr := fmt.Sprintf(":%d", config.port)
	klog.Fatal(http.ListenAndServe(listenAddr, nil))
}

func main() {
	klog.InitFlags(nil)
	config := readAndValidateConfig()

	configureRoutes()

	client := sb.New(config.connectionString, config.timeout)
	coll := collector.New(client)
	prometheus.MustRegister(coll)

	startHTTPServer(config)
}
