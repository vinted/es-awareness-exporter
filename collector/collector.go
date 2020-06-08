package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type shardCollector struct {
	shardsZoneAware    *prometheus.Desc
	shardsNotZoneAware *prometheus.Desc
	collectionTime     *prometheus.Desc
	scrapeTime         *prometheus.Desc
}

type shardAwarenessMetrics struct {
	shardsZoneAware    float64
	shardsNotZoneAware float64
	collectionTime     float64
}

var clusterName string
var metricsCtx shardAwarenessMetrics

func NewShardCollector() *shardCollector {
	return &shardCollector{
		shardsZoneAware: prometheus.NewDesc("es_awareness_zone_aware_shards_count",
			"Number of shards with primary and replica located in different zones",
			[]string{"cluster"}, nil,
		),
		shardsNotZoneAware: prometheus.NewDesc("es_awareness_zone_not_aware_shards_count",
			"Number of shards with primary and replica located in the same zone",
			[]string{"cluster"}, nil,
		),
		collectionTime: prometheus.NewDesc("es_awareness_metric_collection_time",
			"Time it took for a collection thread to collect metrics",
			nil, nil,
		),
		scrapeTime: prometheus.NewDesc("es_awareness_metric_scrape_time",
			"Time it took for prometheus to scrape metrics",
			nil, nil,
		),
	}
}

func (collector *shardCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.shardsZoneAware
	ch <- collector.shardsNotZoneAware
	ch <- collector.collectionTime
	ch <- collector.scrapeTime
}

func (collector *shardCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeTime := time.Now()
	ch <- prometheus.MustNewConstMetric(collector.shardsZoneAware, prometheus.GaugeValue, metricsCtx.shardsZoneAware, clusterName)
	ch <- prometheus.MustNewConstMetric(collector.shardsNotZoneAware, prometheus.GaugeValue, metricsCtx.shardsNotZoneAware, clusterName)
	ch <- prometheus.MustNewConstMetric(collector.collectionTime, prometheus.GaugeValue, metricsCtx.collectionTime)
	ch <- prometheus.MustNewConstMetric(collector.scrapeTime, prometheus.GaugeValue, time.Since(scrapeTime).Seconds())
}

func CollectTimer(queryInterval int, esAddress string) {
	tickChan := time.NewTicker(time.Second * time.Duration(queryInterval))
	defer tickChan.Stop()
	for range tickChan.C {
		log.Debug("collectMetrics triggered")
		metricsCtx.collectMetrics(esAddress)
		log.Debug("collectMetrics ended")
	}
}

func (pm *shardAwarenessMetrics) collectMetrics(esAddress string) {
	collectionTime := time.Now()
	pm.shardsZoneAware, pm.shardsNotZoneAware = getShardsAwarenessStats(esAddress)
	pm.collectionTime = time.Since(collectionTime).Seconds()
}

type shardAttributes struct {
	Index  string
	Shard  string
	Prirep string
	Node   string
}

func getShardsAwarenessStats(esAddress string) (float64, float64) {
	var countAware float64
	var countUnAware float64
	var indexAttributes []shardAttributes
	var shardList []byte

	shardList = getEsShardsList(esAddress)
	clusterName = GetEsClusterName(esAddress)

	err := json.Unmarshal(shardList, &indexAttributes)
	if err != nil {
		log.Error(err)
	}

	log.Debug("Shard list retreived. Shards total: ", len(indexAttributes))

	shardMap := make(map[string][]string)

	// Map shards to zones
	for _, shard := range indexAttributes {
		indexShard := strings.Join([]string{shard.Index, shard.Shard}, "-")
		zone := strings.Split(shard.Node, "-")[1]
		if shardas, ok := shardMap[indexShard]; ok {
			if shardas[0] != zone {
				shardMap[indexShard] = append(shardMap[indexShard], zone)
			}
		} else {
			shardMap[indexShard] = append(shardMap[indexShard], zone)
		}
	}

	// Shard is considered 'zone-unaware' if it has only one zone attatched
	for _, value := range shardMap {
		if len(value) == 1 {
			countUnAware++
		} else {
			countAware++
		}
	}

	return countAware, countUnAware
}

func getEsShardsList(esAddress string) []byte {
	endpoint := []string{esAddress, "/_cat/shards?h=index,shard,prirep,node&format=json"}
	url := strings.Join(endpoint, "")
	return getJSON(url)
}

func GetEsClusterName(esAddress string) string {
	var f interface{}

	clusterInfo := getJSON(esAddress)
	err := json.Unmarshal(clusterInfo, &f)
	if err != nil {
		log.Error(err)
	}
	m := f.(map[string]interface{})
	for k, v := range m {
		if k == "cluster_name" {
			clusterName := fmt.Sprintf("%v", v)
			log.Debug("Cluster name retreived: ", clusterName)
			return clusterName
		}
	}
	return "Unknown"
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func getJSON(url string) []byte {
	r, err := httpClient.Get(url)
	if err != nil {
		log.Error(err)
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
	}

	return body
}
