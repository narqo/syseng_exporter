package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func readConstGaugeMetric(m prometheus.Metric) float64 {
	pb := &dto.Metric{}
	m.Write(pb)
	return pb.GetGauge().GetValue()
}

func readConstSummaryMetric(m prometheus.Metric) (uint64, float64) {
	pb := &dto.Metric{}
	m.Write(pb)
	return pb.GetSummary().GetSampleCount(), pb.GetSummary().GetSampleSum()
}

type testSvc struct {
	*httptest.Server
	Response string
	Up       float64
	Count    uint64
	Sum      float64
}

func newTestSvc(resp string) testSvc {
	s := testSvc{
		Response: resp,
	}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, s.Response)
	}))
	return s
}

func testExporterCollect(t *testing.T, tsvc *testSvc, ch chan prometheus.Metric) {
	e := NewSysengExporter(context.Background(), "test", tsvc.URL)

	go func() {
		defer close(ch)
		e.Collect(ch)
	}()

	up := readConstGaugeMetric((<-ch).(prometheus.Metric))
	if up != tsvc.Up {
		t.Fatalf("up: want %f, got %f", tsvc.Up, up)
	}
}

func TestExporter(t *testing.T) {
	testResp := `
{
  "requestCounters": {
    "200": 65221,
    "404": 14066,
    "500": 12618
  },
  "requestRates": {
    "200": 100,
    "404": 1
  },
  "duration": {
    "count": 91905,
    "sum": 4484.3037570333245,
    "average": 0.024613801985478054
  }
}`
	tsvc := newTestSvc(testResp)
	tsvc.Up = 1.0
	tsvc.Count = 91905
	tsvc.Sum = 4484.3037570333245
	defer tsvc.Close()

	ch := make(chan prometheus.Metric)

	testExporterCollect(t, &tsvc, ch)

	// can't guaranty the order in which codes will be added to ch, so only check total length and that values are not zero
	for n := 0; n < 3; n++ {
		if reqTotalCode := readConstGaugeMetric((<-ch).(prometheus.Metric)); reqTotalCode == 0 {
			t.Errorf("%d reqTotal: bad request total: %f", n, reqTotalCode)
		}
	}

	reqDurationCount, reqDurationSum := readConstSummaryMetric((<-ch).(prometheus.Metric))
	if reqDurationCount != tsvc.Count {
		t.Errorf("request_duration: want %d, got %d", tsvc.Count, reqDurationCount)
	}
	if reqDurationSum != tsvc.Sum {
		t.Errorf("request_duration: want %f, got %f", tsvc.Sum, reqDurationSum)
	}
}

func TestExporterSvcDown(t *testing.T) {
	tsvc := newTestSvc("500 Internal Server Error")
	defer tsvc.Close()

	ch := make(chan prometheus.Metric)

	testExporterCollect(t, &tsvc, ch)
}
