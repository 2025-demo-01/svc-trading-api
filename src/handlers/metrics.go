package main 

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Metrics() http.Handler {
	return promhttp.Handler()
}
