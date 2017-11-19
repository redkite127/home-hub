package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
var (
	nodeCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_status",
			Help: "Soa manager service status.",
		},
		[]string{"venture", "service", "satus", "resource"},
	)
)

func main() {
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.MustRegister(nodeCounter)

	http.Handle("/metrics", prometheus.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
