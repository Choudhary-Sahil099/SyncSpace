package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type Metrics struct {
	Requests           uint64
	Errors             uint64
	TotalLatencyMicros uint64
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		atomic.AddUint64(&m.Requests, 1)
		atomic.AddUint64(&m.TotalLatencyMicros, uint64(time.Since(start).Microseconds()))
	})
}

func (m *Metrics) Handler(w http.ResponseWriter, _ *http.Request) {
	requests := atomic.LoadUint64(&m.Requests)
	errors := atomic.LoadUint64(&m.Errors)
	total := atomic.LoadUint64(&m.TotalLatencyMicros)
	avg := uint64(0)
	if requests > 0 {
		avg = total / requests
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "syncspace_requests_total %d\n", requests)
	fmt.Fprintf(w, "syncspace_errors_total %d\n", errors)
	fmt.Fprintf(w, "syncspace_request_latency_microseconds_avg %d\n", avg)
}
