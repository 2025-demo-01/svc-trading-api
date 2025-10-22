package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", Health).Methods("GET")
	r.Handle("/metrics", Metrics())
	r.HandleFunc("/api/v1/trade/orders", CreateOrder).Methods("POST")
	return r
}
