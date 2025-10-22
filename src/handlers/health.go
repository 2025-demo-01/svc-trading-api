package main

import (
	"net"
	"net/http"
	"os"
	"time"
)

// [ADDED] 실제 환경에 맞게 ping 로직 교체(아래는 placeholder)
func kafkaHealthy() bool {
	// ex) net.DialTimeout("tcp", "<broker>:9094", 500*time.Millisecond)
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return false
	}
	// 가벼운 port ping ~ 예시(첫 broker만)
	conn, err := net.DialTimeout("tcp", brokers, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func dbHealthy() bool {
	endpoint := os.Getenv("DB_ENDPOINT")
	if endpoint == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", endpoint, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// [ADDED] Readiness: Kafka/DB 둘 다 붙어야 200 뜨더라 
func Ready(w http.ResponseWriter, r *http.Request) {
	if kafkaHealthy() && dbHealthy() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("not-ready"))
}
