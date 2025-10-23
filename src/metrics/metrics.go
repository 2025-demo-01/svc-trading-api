package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	OrderAccepted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_accepted_total",
			Help: "Count of accepted orders",
		},
		[]string{"symbol","side"},
	)

	OrderRejected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_rejected_total",
			Help: "Count of rejected orders",
		},
		[]string{"reason"},
	)

	OrderLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "orders_request_duration_ms",
			Help:    "Order request duration in ms",
			Buckets: []float64{10,20,50,100,200,300,500,800,1200,2000},
		},
		[]string{"path"},
	)
)

func MustRegister() {
	prometheus.MustRegister(OrderAccepted, OrderRejected, OrderLatency)
}
