package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	OrdersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_orders_total",
			Help: "orders result counter",
		},
		[]string{"result"},
	)
	OrdersLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "trading_orders_latency_ms",
		Help:    "orders endpoint latency",
		Buckets: []float64{50, 100, 200, 300, 500, 800, 1200, 2000},
	})
)

func MustRegister() {
	prometheus.MustRegister(OrdersTotal, OrdersLatency)
}
