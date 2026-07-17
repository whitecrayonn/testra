package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var defaultBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05,
	0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// Registry is a lightweight, Prometheus-compatible metrics store.
type Registry struct {
	mu          sync.RWMutex
	jobCounter  map[string]uint64
	jobDuration map[string]*histogram
	mlCounter   map[string]uint64
	mlDuration  map[string]*histogram
	queueStatus map[string]int64
	db          *sql.DB
}

type histogram struct {
	buckets []float64
	counts  []uint64
	sum     float64
	count   uint64
}

func newHistogram() *histogram {
	return &histogram{
		buckets: defaultBuckets,
		counts:  make([]uint64, len(defaultBuckets)+1),
	}
}

func (h *histogram) observe(v float64) {
	idx := len(h.buckets)
	for i, b := range h.buckets {
		if v <= b {
			idx = i
			break
		}
	}
	for i := idx; i < len(h.counts); i++ {
		h.counts[i]++
	}
	h.sum += v
	h.count++
}

func newRegistry(db *sql.DB) *Registry {
	return &Registry{
		jobCounter:  make(map[string]uint64),
		jobDuration: make(map[string]*histogram),
		mlCounter:   make(map[string]uint64),
		mlDuration:  make(map[string]*histogram),
		queueStatus: make(map[string]int64),
		db:          db,
	}
}

var defaultRegistry = newRegistry(nil)

// RecordJob records the duration and outcome of a queue job.
// status is one of: success, retry, dead_letter.
func RecordJob(jobType, status string, duration time.Duration) {
	defaultRegistry.recordJob(jobType, status, duration)
}

// RecordMLCall records the latency and outcome of a call to the Python ML engine.
func RecordMLCall(method, status string, duration time.Duration) {
	defaultRegistry.recordMLCall(method, status, duration)
}

// Handler returns an http.Handler that exposes /metrics in Prometheus text format.
func Handler(db *sql.DB) http.Handler {
	if db != nil {
		defaultRegistry.mu.Lock()
		defaultRegistry.db = db
		defaultRegistry.mu.Unlock()
	}
	return http.HandlerFunc(defaultRegistry.serveHTTP)
}

func (r *Registry) recordJob(jobType, status string, duration time.Duration) {
	key := jobType + "|" + status
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobCounter[key]++
	h := r.jobDuration[key]
	if h == nil {
		h = newHistogram()
		r.jobDuration[key] = h
	}
	h.observe(duration.Seconds())
}

func (r *Registry) recordMLCall(method, status string, duration time.Duration) {
	key := method + "|" + status
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mlCounter[key]++
	h := r.mlDuration[key]
	if h == nil {
		h = newHistogram()
		r.mlDuration[key] = h
	}
	h.observe(duration.Seconds())
}

func (r *Registry) refreshQueueStatus(ctx context.Context) {
	r.mu.RLock()
	db := r.db
	r.mu.RUnlock()
	if db == nil {
		return
	}

	rows, err := db.QueryContext(ctx, "SELECT status, COUNT(*) FROM queue_jobs GROUP BY status")
	if err != nil {
		return
	}
	defer rows.Close()

	status := make(map[string]int64)
	for rows.Next() {
		var st string
		var n int64
		if err := rows.Scan(&st, &n); err != nil {
			continue
		}
		status[st] = n
	}
	if err := rows.Err(); err != nil {
		return
	}

	r.mu.Lock()
	r.queueStatus = status
	r.mu.Unlock()
}

func splitKey(key string) (string, string) {
	parts := strings.SplitN(key, "|", 2)
	if len(parts) != 2 {
		return key, ""
	}
	return parts[0], parts[1]
}

func sortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (r *Registry) serveHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	r.refreshQueueStatus(ctx)

	r.mu.RLock()
	jobCounter := copyMapUint64(r.jobCounter)
	jobDuration := copyHistMap(r.jobDuration)
	mlCounter := copyMapUint64(r.mlCounter)
	mlDuration := copyHistMap(r.mlDuration)
	queueStatus := copyMapInt64(r.queueStatus)
	r.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder

	b.WriteString("# HELP testra_worker_jobs_total Total number of processed queue jobs\n")
	b.WriteString("# TYPE testra_worker_jobs_total counter\n")
	for _, key := range sortedStringKeys(jobCounter) {
		jobType, status := splitKey(key)
		fmt.Fprintf(&b, "testra_worker_jobs_total{job_type=%q,status=%q} %d\n", jobType, status, jobCounter[key])
	}

	writeHistogram(&b, "testra_worker_job_duration_seconds", "Queue job execution duration", jobDuration)

	b.WriteString("# HELP testra_ml_requests_total Total number of requests to the ML engine\n")
	b.WriteString("# TYPE testra_ml_requests_total counter\n")
	for _, key := range sortedStringKeys(mlCounter) {
		method, status := splitKey(key)
		fmt.Fprintf(&b, "testra_ml_requests_total{method=%q,status=%q} %d\n", method, status, mlCounter[key])
	}

	writeHistogram(&b, "testra_ml_request_duration_seconds", "ML engine request duration", mlDuration)

	b.WriteString("# HELP testra_worker_queue_jobs_status Number of queue jobs by status\n")
	b.WriteString("# TYPE testra_worker_queue_jobs_status gauge\n")
	for _, status := range sortedStringKeys(queueStatus) {
		fmt.Fprintf(&b, "testra_worker_queue_jobs_status{status=%q} %d\n", status, queueStatus[status])
	}

	fmt.Fprint(w, b.String())
}

func writeHistogram(b *strings.Builder, name, help string, hist map[string]*histogram) {
	b.WriteString("# HELP ")
	b.WriteString(name)
	b.WriteString(" ")
	b.WriteString(help)
	b.WriteString("\n# TYPE ")
	b.WriteString(name)
	b.WriteString(" histogram\n")

	for _, key := range sortedStringKeys(hist) {
		labelA, labelB := splitKey(key)
		h := hist[key]
		for i, bucket := range h.buckets {
			le := strconv.FormatFloat(bucket, 'g', -1, 64)
			fmt.Fprintf(b, "%s_bucket{%s,le=%q} %d\n", name, labelPair(name, labelA, labelB), le, h.counts[i])
		}
		fmt.Fprintf(b, "%s_bucket{%s,le=\"+Inf\"} %d\n", name, labelPair(name, labelA, labelB), h.counts[len(h.counts)-1])
		fmt.Fprintf(b, "%s_sum{%s} %g\n", name, labelPair(name, labelA, labelB), h.sum)
		fmt.Fprintf(b, "%s_count{%s} %d\n", name, labelPair(name, labelA, labelB), h.count)
	}
}

func labelPair(name, a, b string) string {
	switch name {
	case "testra_ml_request_duration_seconds":
		return fmt.Sprintf("method=%q,status=%q", a, b)
	default:
		return fmt.Sprintf("job_type=%q,status=%q", a, b)
	}
}

func copyMapUint64(m map[string]uint64) map[string]uint64 {
	out := make(map[string]uint64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyMapInt64(m map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyHistMap(m map[string]*histogram) map[string]*histogram {
	out := make(map[string]*histogram, len(m))
	for k, h := range m {
		counts := make([]uint64, len(h.counts))
		copy(counts, h.counts)
		out[k] = &histogram{
			buckets: h.buckets,
			counts:  counts,
			sum:     h.sum,
			count:   h.count,
		}
	}
	return out
}
