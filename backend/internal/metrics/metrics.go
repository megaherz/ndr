package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// RPC metrics
	RPCRequestsTotal    *prometheus.CounterVec
	RPCRequestDuration  *prometheus.HistogramVec
	RPCRequestsInFlight prometheus.Gauge

	// Matchmaking metrics
	MatchmakingWaitTime  *prometheus.HistogramVec
	MatchmakingQueueSize *prometheus.GaugeVec
	MatchmakingTimeouts  *prometheus.CounterVec
	ActiveMatches        prometheus.Gauge
	MatchDuration        *prometheus.HistogramVec

	// Economy metrics
	HouseFuelBalance   prometheus.Gauge
	RakeFuelBalance    prometheus.Gauge
	TotalPrizesAwarded *prometheus.CounterVec
	TotalBurnRewards   *prometheus.CounterVec

	// TonCenter metrics
	TonCenterRequestsTotal   *prometheus.CounterVec
	TonCenterRequestDuration *prometheus.HistogramVec
	TonCenterErrors          *prometheus.CounterVec

	// Settlement metrics
	SettlementDuration *prometheus.HistogramVec
	SettlementErrors   *prometheus.CounterVec
}

// New creates a new Metrics instance with all metrics registered
func New() *Metrics {
	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),

		// RPC metrics
		RPCRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_requests_total",
				Help: "Total number of RPC requests",
			},
			[]string{"method", "status"},
		),
		RPCRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rpc_request_duration_seconds",
				Help:    "Duration of RPC requests in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"method"},
		),
		RPCRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rpc_requests_in_flight",
				Help: "Number of RPC requests currently being processed",
			},
		),

		// Matchmaking metrics
		MatchmakingWaitTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "matchmaking_wait_time_seconds",
				Help:    "Time players wait in matchmaking queue",
				Buckets: []float64{1, 2, 5, 10, 15, 20, 30, 45, 60},
			},
			[]string{"league"},
		),
		MatchmakingQueueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "matchmaking_queue_size",
				Help: "Number of players in matchmaking queue",
			},
			[]string{"league"},
		),
		MatchmakingTimeouts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "matchmaking_timeouts_total",
				Help: "Total number of matchmaking timeouts",
			},
			[]string{"league"},
		),
		ActiveMatches: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_matches",
				Help: "Number of matches currently in progress",
			},
		),
		MatchDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "match_duration_seconds",
				Help:    "Duration of matches from start to settlement",
				Buckets: []float64{60, 120, 180, 240, 300, 360, 420, 480, 600},
			},
			[]string{"league"},
		),

		// Economy metrics
		HouseFuelBalance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "house_fuel_balance",
				Help: "Current FUEL balance in house wallet",
			},
		),
		RakeFuelBalance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rake_fuel_balance",
				Help: "Current FUEL balance in rake wallet",
			},
		),
		TotalPrizesAwarded: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "total_prizes_awarded",
				Help: "Total FUEL prizes awarded to players",
			},
			[]string{"league"},
		),
		TotalBurnRewards: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "total_burn_rewards",
				Help: "Total BURN rewards awarded to players",
			},
			[]string{"league"},
		),

		// TonCenter metrics
		TonCenterRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "toncenter_requests_total",
				Help: "Total number of TonCenter API requests",
			},
			[]string{"method", "status"},
		),
		TonCenterRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "toncenter_request_duration_seconds",
				Help:    "Duration of TonCenter API requests",
				Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0},
			},
			[]string{"method"},
		),
		TonCenterErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "toncenter_errors_total",
				Help: "Total number of TonCenter API errors",
			},
			[]string{"method", "error_type"},
		),

		// Settlement metrics
		SettlementDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "settlement_duration_seconds",
				Help:    "Duration of match settlement process",
				Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"league"},
		),
		SettlementErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "settlement_errors_total",
				Help: "Total number of settlement errors",
			},
			[]string{"league", "error_type"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.RPCRequestsTotal,
		m.RPCRequestDuration,
		m.RPCRequestsInFlight,
		m.MatchmakingWaitTime,
		m.MatchmakingQueueSize,
		m.MatchmakingTimeouts,
		m.ActiveMatches,
		m.MatchDuration,
		m.HouseFuelBalance,
		m.RakeFuelBalance,
		m.TotalPrizesAwarded,
		m.TotalBurnRewards,
		m.TonCenterRequestsTotal,
		m.TonCenterRequestDuration,
		m.TonCenterErrors,
		m.SettlementDuration,
		m.SettlementErrors,
	)

	return m
}

// Handler returns the Prometheus metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// RecordHTTPRequest records metrics for an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordRPCRequest records metrics for an RPC request
func (m *Metrics) RecordRPCRequest(method, status string, duration time.Duration) {
	m.RPCRequestsTotal.WithLabelValues(method, status).Inc()
	m.RPCRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordMatchmakingWait records matchmaking wait time
func (m *Metrics) RecordMatchmakingWait(league string, duration time.Duration) {
	m.MatchmakingWaitTime.WithLabelValues(league).Observe(duration.Seconds())
}

// SetQueueSize sets the current queue size for a league
func (m *Metrics) SetQueueSize(league string, size float64) {
	m.MatchmakingQueueSize.WithLabelValues(league).Set(size)
}

// RecordMatchmakingTimeout records a matchmaking timeout
func (m *Metrics) RecordMatchmakingTimeout(league string) {
	m.MatchmakingTimeouts.WithLabelValues(league).Inc()
}

// SetActiveMatches sets the number of active matches
func (m *Metrics) SetActiveMatches(count float64) {
	m.ActiveMatches.Set(count)
}

// RecordMatchDuration records the duration of a completed match
func (m *Metrics) RecordMatchDuration(league string, duration time.Duration) {
	m.MatchDuration.WithLabelValues(league).Observe(duration.Seconds())
}

// SetHouseFuelBalance sets the current house FUEL balance
func (m *Metrics) SetHouseFuelBalance(balance float64) {
	m.HouseFuelBalance.Set(balance)
}

// SetRakeFuelBalance sets the current rake FUEL balance
func (m *Metrics) SetRakeFuelBalance(balance float64) {
	m.RakeFuelBalance.Set(balance)
}

// RecordPrizeAwarded records a prize awarded to a player
func (m *Metrics) RecordPrizeAwarded(league string, amount float64) {
	m.TotalPrizesAwarded.WithLabelValues(league).Add(amount)
}

// RecordBurnReward records a BURN reward awarded to a player
func (m *Metrics) RecordBurnReward(league string, amount float64) {
	m.TotalBurnRewards.WithLabelValues(league).Add(amount)
}

// RecordTonCenterRequest records metrics for a TonCenter API request
func (m *Metrics) RecordTonCenterRequest(method, status string, duration time.Duration) {
	m.TonCenterRequestsTotal.WithLabelValues(method, status).Inc()
	m.TonCenterRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordTonCenterError records a TonCenter API error
func (m *Metrics) RecordTonCenterError(method, errorType string) {
	m.TonCenterErrors.WithLabelValues(method, errorType).Inc()
}

// RecordSettlementDuration records the duration of a settlement process
func (m *Metrics) RecordSettlementDuration(league string, duration time.Duration) {
	m.SettlementDuration.WithLabelValues(league).Observe(duration.Seconds())
}

// RecordSettlementError records a settlement error
func (m *Metrics) RecordSettlementError(league, errorType string) {
	m.SettlementErrors.WithLabelValues(league, errorType).Inc()
}
