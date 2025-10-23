package main

import (
	"context"     // [ADDED]
	"log"
	"net/http"
	"os"          // [ADDED]
	"os/signal"   // [ADDED]
	"syscall"     // [ADDED]
	"time"        // [ADDED]

	"go.uber.org/zap"                     // [ADDED]
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp" // [ADDED]

	"github.com/2025-demo-01/svc-trading-api/src/pkg/logger"   // [ADDED]
	"github.com/2025-demo-01/svc-trading-api/src/pkg/metrics"  // [ADDED]
	"github.com/2025-demo-01/svc-trading-api/src/pkg/db"       // [ADDED]
	"github.com/2025-demo-01/svc-trading-api/src/pkg/kafka"    // [ADDED]
)

func main() {
	logger.Init()               // [ADDED]
	defer logger.L().Sync()     // [ADDED]
	metrics.MustRegister()      // [ADDED]

	// [ADDED] 외부 의존성 초기화(실패 시 프로세스 종료 → 빠른 실패)
	db.MustInit(os.Getenv("DB_ENDPOINT"))
	kafka.MustInit(os.Getenv("KAFKA_BROKERS"))

	r := setupRouter()

	// [ADDED] OpenTelemetry HTTP wrapper
	handler := otelhttp.NewHandler(r, "svc-trading-api")

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handler, // [CHANGED]
		ReadHeaderTimeout: 5 * time.Second,
	}

	// [ADDED] graceful shutdown
	go func() {
		logger.L().Info("svc-trading-api listening", zap.String("addr", ":8080"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Fatal("listen error", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.L().Info("shutdown complete")
}
