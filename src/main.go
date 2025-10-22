package main

import (
	"context"          // [ADDED]
	"log"
	"net/http"
	"os"               // [ADDED]
	"os/signal"        // [ADDED]
	"syscall"          // [ADDED]
	"time"             // [ADDED]
)

func main() {
	r := setupRouter()

	// [ADDED] 서버 타임아웃 & 헤더 타임아웃
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// [ADDED] graceful shutdown
	go func() {
		log.Println("svc-trading-api listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down http server...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("bye")
}
