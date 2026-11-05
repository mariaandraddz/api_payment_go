package observability


var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of requests",
		}, []string{"method", "path", "status"})

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of http request",
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
			Help: "Total of insufficient funds", // modificar
		}
		)
)