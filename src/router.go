package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRouter() http.Handler {
	r := mux.NewRouter()

	// Health/Ready/Metrics
	r.HandleFunc("/healthz", Health).Methods("GET")
	r.HandleFunc("/readyz", Ready).Methods("GET") // [ADDED]
	r.Handle("/metrics", Metrics())

	// Business endpoints (stub)
	r.HandleFunc("/api/v1/trade/orders", CreateOrder).Methods("POST")

	return r
}
