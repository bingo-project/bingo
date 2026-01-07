// ABOUTME: AI chat business metrics for monitoring and observability.
// ABOUTME: Tracks request duration, success rates, fallback usage, and quota.

package chat

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// aiRequestDuration tracks AI request duration by provider and model.
	aiRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_request_duration_seconds",
			Help:    "AI request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider", "model", "stream"},
	)

	// aiRequestsTotal tracks total AI requests by status.
	aiRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_requests_total",
			Help: "Total AI requests",
		},
		[]string{"provider", "model", "stream", "status"},
	)

	// aiFallbackTotal tracks fallback usage.
	aiFallbackTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_fallback_total",
			Help: "Total AI fallback activations",
		},
		[]string{"from_provider", "to_provider"},
	)

	// aiQuotaReservation tracks quota reservation operations.
	aiQuotaReservation = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_quota_reservation_total",
			Help: "Total quota reservation operations",
		},
		[]string{"operation"}, // operation: reserve, adjust, release
	)

	// aiCircuitBreakerState tracks circuit breaker state changes.
	aiCircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_circuit_breaker_state",
			Help: "Circuit breaker state (0=open, 0.5=half-open, 1=closed)",
		},
		[]string{"provider"},
	)

	// aiCircuitBreakerFailures tracks circuit breaker failure counts.
	aiCircuitBreakerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_circuit_breaker_failures_total",
			Help: "Total circuit breaker triggered failures",
		},
		[]string{"provider"},
	)

	// aiRPMRejections tracks requests rejected due to RPM limit.
	aiRPMRejections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_rpm_rejections_total",
			Help: "Total requests rejected due to rate limiting",
		},
		[]string{}, // no labels for now
	)
)

// RecordRequest records an AI request with duration and result.
func RecordRequest(provider, model string, isStream bool, duration float64, status string) {
	aiRequestDuration.WithLabelValues(provider, model, boolToString(isStream)).Observe(duration)
	aiRequestsTotal.WithLabelValues(provider, model, boolToString(isStream), status).Inc()
}

// RecordFallback records a fallback event.
func RecordFallback(fromProvider, toProvider string) {
	aiFallbackTotal.WithLabelValues(fromProvider, toProvider).Inc()
}

// RecordQuotaOperation records a quota operation.
func RecordQuotaOperation(operation string) {
	aiQuotaReservation.WithLabelValues(operation).Inc()
}

// SetCircuitBreakerState sets the circuit breaker state for monitoring.
func SetCircuitBreakerState(provider string, state CircuitBreakerState) {
	var value float64
	switch state {
	case CircuitOpen:
		value = 0
	case CircuitHalfOpen:
		value = 0.5
	case CircuitClosed:
		value = 1
	}
	aiCircuitBreakerState.WithLabelValues(provider).Set(value)
}

// RecordCircuitBreakerFailure records when circuit breaker rejects a request.
func RecordCircuitBreakerFailure(provider string) {
	aiCircuitBreakerFailures.WithLabelValues(provider).Inc()
}

// RecordRPMRejection records when a request is rejected due to rate limiting.
func RecordRPMRejection() {
	aiRPMRejections.WithLabelValues().Inc()
}

func boolToString(b bool) string {
	if b {
		return "true"
	}

	return "false"
}
