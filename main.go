package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const namespace = "syseng"

type sysengSvcResp struct {
	ReqCounters map[string]float64 `json:"requestCounters"`
	ReqRates    map[string]float64 `json:"requestRates"`
	Duration    sysengDuration     `json:"duration"`
}

type sysengDuration struct {
	Count   uint64
	Sum     float64
	Average float64
}

type Exporter struct {
	uri       string
	ctx       context.Context
	svcClient *http.Client

	up *prometheus.Desc

	reqCount    *prometheus.Desc
	reqDuration *prometheus.Desc
}

var _ prometheus.Collector = &Exporter{}

func NewSysengExporter(ctx context.Context, namespace, uri string) prometheus.Collector {
	s := &Exporter{
		uri: uri,
		ctx: ctx,
		// NOTE: in real production environment, svcClient must have sane timeouts
		svcClient: http.DefaultClient,
	}
	s.up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last scrape successful.",
		nil, nil,
	)
	s.reqCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "http_requests_total"),
		"How many HTTP requests has been served by status code.",
		[]string{"code"},
		nil,
	)
	s.reqDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "http_request_duration_seconds"),
		"Summary of HTTP request durations.",
		nil,
		nil,
	)
	return s
}

// Describe sends the super-set of all collected descriptors of metrics to the provided channel.
// It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.reqCount
	ch <- e.reqDuration
}

// Collect fetches the stats from syseng svc and delivers them as Prometheus metrics.
// It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	resp := sysengSvcResp{}
	if err := e.scrapeSvc(&resp); err != nil {
		log.Errorf("Failed to scrap metrics from %q: %v", e.uri, err)
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	for code, val := range resp.ReqCounters {
		ch <- prometheus.MustNewConstMetric(e.reqCount, prometheus.CounterValue, val, code)
	}
	ch <- prometheus.MustNewConstSummary(
		e.reqDuration,
		resp.Duration.Count,
		resp.Duration.Sum,
		nil,
	)
}

func (e *Exporter) scrapeSvc(v interface{}) error {
	req, err := http.NewRequest("GET", e.uri, nil)
	if err != nil {
		return fmt.Errorf("cound not create svc request: %v", err)
	}
	// NOTE: e.ctx might have a timeout/deadline/trace
	req = req.WithContext(e.ctx)

	resp, err := e.svcClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

var revision = "dev"

var (
	addr        = flag.String("addr", ":8081", "listen address")
	metricsPath = flag.String("metrics-path", "/metrics", "path to expose metrics")
	svcAddr     = flag.String("syseng.stats-uri", "http://localhost:8080/stats", "address on which to scrape Syseng service")
)

func main() {
	flag.Parse()

	e := NewSysengExporter(context.TODO(), namespace, *svcAddr)
	prometheus.MustRegister(e)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", handleRoot)

	log.Infoln("Version " + revision + " listening on " + *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var rootHTML = []byte(`<!doctype html>
<html>
<head><title>SysEng Stats Exporter</title></head>
<body>
<h1>SysEng Stats Exporter</h1>
<p><a href="` + *metricsPath + `">Show Metrics</a></p>
</body>
</html>
`)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write(rootHTML)
}
