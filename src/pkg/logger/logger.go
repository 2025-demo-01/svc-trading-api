package logger

import "go.uber.org/zap"

var lg *zap.Logger

func Init() {
	lg, _ = zap.NewProduction()
}

func L() *zap.Logger {
	return lg
}
