package observability

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tracerProvider *sdktrace.TracerProvider

func InitTracing(ctx context.Context) func(context.Context) error {
	tp := sdktrace.NewTracerProvider()
	tracerProvider = tp
	otel.SetTracerProvider(tp)

	svc := os.Getenv("SERVICE_NAME")
	if svc == "" { svc = "svc-trading-api" }
	otel.GetTextMapPropagator() // default W3C

	return tp.Shutdown
}

func CommonAttrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("service.name", "svc-trading-api"),
	}
}
