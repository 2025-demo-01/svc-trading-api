func ProduceTradeEvent(p Producer, evt TradeEvent, traceID string) error {
    b, _ := json.Marshal(evt)
    msg := kafka.Message{
        Topic: "orders.in",
        Key:   []byte(evt.OrderID),
        Value: b,
        Headers: []kafka.Header{
            {Key: "schema-version", Value: []byte("v1")},
            {Key: "trace-id", Value: []byte(traceID)},
        },
    }
    return p.WriteMessages(ctx, msg)
}
