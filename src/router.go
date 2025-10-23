package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/2025-demo-01/svc-trading-api/src/middleware" // [ADDED]
)

func setupRouter() http.Handler {
	r := mux.NewRouter()

	// Global middlewares
	r.Use(middleware.WithRequestID) // [ADDED] 상관관계 추적
	r.Use(middleware.Recoverer)     // [ADDED] panic recover + 500
	r.Use(middleware.Timeout)       // [ADDED] 요청 타임아웃

	// Health/Ready/Metrics
	r.HandleFunc("/healthz", Health).Methods("GET")
	r.HandleFunc("/readyz", Ready).Methods("GET")
	r.Handle("/metrics", Metrics())

	// Business endpoints
	r.HandleFunc("/api/v1/trade/orders", CreateOrder).Methods("POST")

	// (선택) 버전 확인
	r.HandleFunc("/__version", Version).Methods("GET") // [ADDED]

	return r
}
