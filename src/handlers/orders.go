package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
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

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 까먹지마 까먹지마:
	// 1) validate → 400
	// 2) Aurora insert(status=pending)
	// 3) Kafka produce(topic=orders.in)
	// 4) return order_id

	resp := OrderResp{OrderID: uuid.NewString(), Status: "accepted"}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
