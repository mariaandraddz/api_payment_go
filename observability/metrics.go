package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of requests",
		}, []string{"method", "path", "status"})

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of http request",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"})

	auth = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_success",
			Help: "Total of successful authentication requests",
		}, []string{"method", "path", "status"})

	declineInsufficientFunds = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "decline_insufficient_funds",
			Help: "Total of decline transations due to insufficient founds", // modificar
		}, []string{"method", "path"},
	)

	declineInactiveCards = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "decline_inactive_cards_total",
			Help: "Total of declined transactions due to inactive cards",
		}, []string{"method", "path"},
	)

	healthcheck = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "healthcheck_total",
			Help: "Total of healthchecks",
		}, []string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(auth)
	prometheus.MustRegister(declineInsufficientFunds)
	prometheus.MustRegister(declineInactiveCards)
	prometheus.MustRegister(healthcheck)
}

func Metric() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())

		path := c.FullPath()
		if path == "" {
			path = "NOT_FOUND"
		}

		httpRequests.WithLabelValues(c.Request.Method, path, status).Inc()
		HttpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
func RecordAuthSuccess(method string, path string, status int) {
	auth.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}

func RecordDeclineInsufficientFunds(method string, path string) {
	declineInsufficientFunds.WithLabelValues(method, path).Inc()
}

func RecordDeclineInactiveCards(method string, path string) {
	declineInactiveCards.WithLabelValues(method, path).Inc()
}

func RecordHealthcheck(method string, path string) {
	healthcheck.WithLabelValues(method, path).Inc()
}
