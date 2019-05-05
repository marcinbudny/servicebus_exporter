package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	sb "github.com/marcinbudny/servicebus_exporter/client"
	"github.com/marcinbudny/servicebus_exporter/collector"
)

var (
	log = logrus.New()
)

type config struct {
	timeout          time.Duration
	port             uint
	verbose          bool
	connectionString string
}

func readAndValidateConfig() config {
	var result config

	flag.StringVar(&result.connectionString, "connection-string", "", "Azure ServiceBus connection string")
	flag.UintVar(&result.port, "port", 9999, "Port to expose scraping endpoint on")
	flag.DurationVar(&result.timeout, "timeout", time.Second*30, "Timeout for scrape")
	flag.BoolVar(&result.verbose, "verbose", false, "Enable verbose logging")

	flag.Parse()

	if result.connectionString == "" {
		log.Fatal("Azure ServiceBus connection string not provided")
	}

	log.WithFields(logrus.Fields{
		"port":    result.port,
		"timeout": result.timeout,
		"verbose": result.verbose,
	}).Infof("Azure ServiceBus exporter configured")

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

func setupLogger(config config) {
	if config.verbose {
		log.Level = logrus.DebugLevel
	}
}

func startHTTPServer(config config) {
	listenAddr := fmt.Sprintf(":%d", config.port)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func main() {

	config := readAndValidateConfig()
	setupLogger(config)

	configureRoutes()

	client := sb.New(config.connectionString, config.timeout)
	coll := collector.New(client, log)
	prometheus.MustRegister(coll)

	startHTTPServer(config)
}
