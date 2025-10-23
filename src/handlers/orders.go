package main

import (
	"encoding/json"
	"errors"   // [ADDED]
	"net/http"
	"strings"  // [ADDED]
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap" // [ADDED]

	"github.com/2025-demo-01/svc-trading-api/src/pkg/logger"   // [ADDED]
	"github.com/2025-demo-01/svc-trading-api/src/pkg/metrics"  // [ADDED]
	"github.com/2025-demo-01/svc-trading-api/src/pkg/kafka"    // [ADDED]
	"github.com/2025-demo-01/svc-trading-api/src/pkg/db"       // [ADDED]
)

type OrderReq struct {
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`   // buy/sell
	Price  float64 `json:"price"`
	Qty    float64 `json:"qty"`
}

type OrderResp struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

// [ADDED] 간단 검증
func (o *OrderReq) Validate() error {
	if o.Symbol == "" || o.Price <= 0 || o.Qty <= 0 {
		return errors.New("invalid payload")
	}
	if s := strings.ToLower(o.Side); s != "buy" && s != "sell" {
		return errors.New("invalid side")
	}
	// (옵션) 허용 심볼 화이트리스트 체크: ALLOWED_SYMBOLS="BTCUSDT,ETHUSDT"
	allowed := strings.Split(strings.TrimSpace(getEnvDefault("ALLOWED_SYMBOLS","")), ",")
	if len(allowed) > 0 && allowed[0] != "" {
		found := false
		for _, sym := range allowed {
			if strings.EqualFold(strings.TrimSpace(sym), o.Symbol) { found = true; break }
		}
		if !found { return errors.New("symbol not allowed") }
	}
	return nil
}

// [ADDED] 메모리 캐시(데모). 실제는 Aurora unique index or DynamoDB 권장.
var idemCache = map[string]OrderResp{}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	start := time.Now()                     // [ADDED]
	reqID := r.Context().Value("reqid")     // [ADDED]

	// [ADDED] Idempotency-Key
	if key := r.Header.Get("Idempotency-Key"); key != "" {
		if resp, ok := idemCache[key]; ok {
			metrics.OrdersTotal.WithLabelValues("idempotent").Inc()
			writeJSON(w, http.StatusOK, resp)
			return
		}
	}

	var req OrderReq
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil { // [ADDED] 1MB 제한
		metrics.OrdersTotal.WithLabelValues("bad_request").Inc()
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil { // [ADDED]
		metrics.OrdersTotal.WithLabelValues("invalid").Inc()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderID := uuid.NewString()
	// [ADDED] DB: pending 기록 (실패 시 500)
	if err := db.InsertPending(r.Context(), orderID, req.Symbol, req.Side, req.Price, req.Qty); err != nil {
		metrics.OrdersTotal.WithLabelValues("db_error").Inc()
		logger.L().Error("db insert failed", zap.Error(err), zap.String("req_id", reqID.(string)))
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// [ADDED] Kafka publish (orders.in) — exactly-once 지향 옵션은 내부에서 설정
	if err := kafka.PublishOrder(r.Context(), orderID, req.Symbol, req.Side, req.Price, req.Qty); err != nil {
		metrics.OrdersTotal.WithLabelValues("kafka_error").Inc()
		logger.L().Error("kafka publish failed", zap.Error(err), zap.String("req_id", reqID.(string)))
		// (옵션) DB 상태를 'failed'로 업데이트
		_ = db.MarkFailed(r.Context(), orderID, "kafka_publish_failed")
		http.Error(w, "queue error", http.StatusServiceUnavailable)
		return
	}

	resp := OrderResp{OrderID: orderID, Status: "accepted"}
	if key := r.Header.Get("Idempotency-Key"); key != "" {
		idemCache[key] = resp
	}

	metrics.OrdersLatency.Observe(float64(time.Since(start).Milliseconds())) // [ADDED]
	metrics.OrdersTotal.WithLabelValues("accepted").Inc()                    // [ADDED]

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// [ADDED] Version 노출(Debugging/realease 추적)
func Version(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{
		"version":  getEnvDefault("APP_VERSION", "0.1.0"),
		"git_sha":  getEnvDefault("GIT_SHA", "dev"),
		"build_ts": getEnvDefault("BUILD_TS", ""),
	})
}
