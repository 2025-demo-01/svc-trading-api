package kafka

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

var w *kafka.Writer

func MustInit(brokers string) {
	if brokers == "" {
		panic("KAFKA_BROKERS empty")
	}
	w = &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,            // [ADDED] 안전
		Async:        false,
		BatchTimeout: 5 * time.Millisecond,
		BatchSize:    100,
	}
}

func Ping(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// metadata fetch 대용으로 no-op write with cancel
	return w.WriteMessages(ctx) == nil
}

type orderEvt struct {
	OrderID string  `json:"order_id"`
	Symbol  string  `json:"symbol"`
	Side    string  `json:"side"`
	Price   float64 `json:"price"`
	Qty     float64 `json:"qty"`
	Ts      int64   `json:"ts"`
}

func PublishOrder(ctx context.Context, orderID, symbol, side string, price, qty float64) error {
	topic := getenv("ORDERS_TOPIC", "orders.in")
	payload, _ := json.Marshal(orderEvt{
		OrderID: orderID, Symbol: symbol, Side: side, Price: price, Qty: qty, Ts: time.Now().UnixMilli(),
	})
	msg := kafka.Message{
		Key:   []byte(orderID), // [ADDED] 파티셔닝 키
		Value: payload,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "traceparent", Value: []byte(getenv("TRACEPARENT", ""))},
		},
	}
	return w.WriteMessages(ctx, msg)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" { return v }
	return d
}
