// Package metrics contains structs and helpers for
// interacting with Prometheus.
package metrics

import "github.com/prometheus/client_golang/prometheus"

// CustomMetrics holds names and types of all custom prometheus metrics
// used in this application.
type Counters struct {
	PackagingErrorCounter    *prometheus.CounterVec
	PackagingProcessDuration *prometheus.CounterVec
}

func (c Counters) AddPackagingErrors(labels prometheus.Labels, count float64) {
	c.PackagingErrorCounter.With(labels).Add(count)
}

func (c Counters) AddPackagingProcessDuration(labels prometheus.Labels, count float64) {
	c.PackagingProcessDuration.With(labels).Add(count)
}
