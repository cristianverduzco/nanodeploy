package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ManagedServicesTotal tracks the total number of ManagedServices by type and phase
	ManagedServicesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nanodeploy_managed_services_total",
			Help: "Total number of ManagedServices by type and phase",
		},
		[]string{"type", "phase"},
	)

	// ReconcileDuration tracks how long each reconcile loop takes
	ReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nanodeploy_reconcile_duration_seconds",
			Help:    "Duration of reconcile loops in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	// ReconcileErrorsTotal tracks reconcile errors by type
	ReconcileErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nanodeploy_reconcile_errors_total",
			Help: "Total number of reconcile errors by service type",
		},
		[]string{"type"},
	)
)

func init() {
	// Register metrics with controller-runtime's registry
	// so they're exposed on the existing :8080/metrics endpoint
	metrics.Registry.MustRegister(
		ManagedServicesTotal,
		ReconcileDuration,
		ReconcileErrorsTotal,
	)
}