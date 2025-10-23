module github.com/2025-demo-01/svc-trading-api

go 1.22

require (
	github.com/gorilla/mux v1.8.1
	github.com/prometheus/client_golang v1.19.0
	github.com/google/uuid v1.6.0
	go.uber.org/zap v1.27.0                         //  구조화 Logging 
	github.com/segmentio/kafka-go v0.4.47           //  Kafka producer (Go Lang)
	github.com/jackc/pgx/v5 v5.6.0                  //  Aurora(Postgres) Driver
	go.opentelemetry.io/otel v1.28.0                //  OpenTelemetry API
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 
)
