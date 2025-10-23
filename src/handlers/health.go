package main

import (
	"encoding/json" // [ADDED]
	"net"
	"net/http"
	"os"
	"time"

	"github.com/2025-demo-01/svc-trading-api/src/pkg/db"    // [ADDED]
	"github.com/2025-demo-01/svc-trading-api/src/pkg/kafka" // [ADDED]
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func Ready(w http.ResponseWriter, r *http.Request) {
	// [CHANGED] 실제 Ping 로직 → 내부 커넥터로 체크
	kok := kafka.Ping(500 * time.Millisecond)
	dok := db.Ping(r.Context(), 500*time.Millisecond)

	status := struct {
		Kafka  bool   `json:"kafka"`
		DB     bool   `json:"db"`
		Env    string `json:"env"`
		Status string `json:"status"`
	}{
		Kafka:  kok,
		DB:     dok,
		Env:    os.Getenv("APP_ENV"),
		Status: "ready",
	}

	if !(kok && dok) {
		w.WriteHeader(http.StatusServiceUnavailable)
		status.Status = "not-ready"
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}
