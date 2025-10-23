package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var conn *pgx.Conn

func MustInit(endpoint string) {
	if endpoint == "" {
		panic("DB_ENDPOINT empty")
	}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://%s", endpoint))
	if err != nil { panic(err) }
}

func Ping(ctx context.Context, d time.Duration) bool {
	c, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return conn.Ping(c) == nil
}

func InsertPending(ctx context.Context, orderID, symbol, side string, price, qty float64) error {
	_, err := conn.Exec(ctx, `insert into orders(order_id, symbol, side, price, qty, status) values ($1,$2,$3,$4,$5,'pending')`,
		orderID, symbol, side, price, qty)
	return err
}

func MarkFailed(ctx context.Context, orderID, reason string) error {
	_, err := conn.Exec(ctx, `update orders set status='failed', fail_reason=$2 where order_id=$1`, orderID, reason)
	return err
}
