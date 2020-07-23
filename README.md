# ElasticSearch shard allocation awareness exporter

## Description
This Exporter is heavily based on [postfix-exporter](https://github.com/vinted/postfix-exporter)  
It exports ElasticSearch Shard allocation awareness metrics.  
Current version depends on ElasticSearch node naming - zone attribute is extracted from hostname. 

Metrics exported:
* `es_awareness_zone_aware_shards_count`
* `es_awareness_zone_not_aware_shards_count`
  
Exporter uses separate collection thread, so scrape time will not be affected on highly loaded ES servers.

## Building

Checkout https://github.com/vinted/es-awareness-exporter repo.  
Build executable:  

 `go build`

## Using

Execute es-awareness-exporter:  

`./es-awareness-exporter`

By default exporter will bind to port `9709`.  

## Configuration

Following config parameters are available:  

```
  -telemetry.addr string
    	host:port Address to listen on for web interface (default ":9709")
  -query.interval int
      How often should daemon query metrics (default 15)
  -query.retries int
      How meny times metrics query should be retried (default 3)
  -query.timeout int
      Query timeout (default 15)
  -log.level string
      Logging level (default "info")
  -es.address sting
      ElasticSearch address http://host:port (default "http://localhost:9200")
```
