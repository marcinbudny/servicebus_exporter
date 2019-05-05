package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	log = logrus.New()

	timeout time.Duration
	port    uint
	verbose bool

	connectionString string
)

func serveLandingPage() {
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
}

func serveMetrics() {
	prometheus.MustRegister(newExporter())

	http.Handle("/metrics", promhttp.Handler())
}

func readAndValidateConfig() {
	flag.StringVar(&connectionString, "connection-string", "", "Azure ServiceBus connection string")
	flag.UintVar(&port, "port", 9999, "Port to expose scraping endpoint on")
	flag.DurationVar(&timeout, "timeout", time.Second*30, "Timeout for scrape")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	flag.Parse()

	if connectionString == "" {
		log.Fatal("Azure ServiceBus connection string not provided")
	}

	log.WithFields(logrus.Fields{
		"port":    port,
		"timeout": timeout,
		"verbose": verbose,
	}).Infof("Azure ServiceBus exporter configured")
}

func setupLogger() {
	if verbose {
		log.Level = logrus.DebugLevel
	}
}

func startHTTPServer() {
	listenAddr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func main() {

	readAndValidateConfig()
	setupLogger()

	serveLandingPage()
	serveMetrics()

	startHTTPServer()
}
