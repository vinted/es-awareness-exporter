package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/vinted/es-awareness-exporter/collector"
)

var (
	bindAddr      = flag.String("telemetry.addr", ":9709", "host:port Address to listen on for web interface")
	queryInterval = flag.Int("query.interval", 15, "How often should daemon query metrics")
	logLevel      = flag.String("log.level", "info", "Logging level")
	esAddress     = flag.String("es.address", "http://localhost:9200", "ElasticSearch address http://host:port")
)

func main() {

	flag.Parse()

	switch *logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	go collector.CollectTimer(*queryInterval, *esAddress)

	pf := collector.NewShardCollector()
	prometheus.MustRegister(pf)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
    <head><title>ES Shard Allocation Awareness Exporter</title></head>
    <body>
    <h1>ElasticSearch shard allocation awareness exporter</h1>
    <p><a href='metrics'>Metrics</a></p>
    </body>
    </html>`))
		if err != nil {
			log.Error("HTTP write failed: ", err)
		}
	})
	log.Info("Listening on: ", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
