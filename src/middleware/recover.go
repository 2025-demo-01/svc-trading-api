package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
	"github.com/2025-demo-01/svc-trading-api/src/pkg/logger"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.L().Error("panic", zap.Any("err", rec), zap.ByteString("stack", debug.Stack()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
