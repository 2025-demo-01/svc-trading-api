package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

func Timeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms := 2000 // default 2s
		if v := r.Context().Value("REQUEST_TIMEOUT_MS"); v != nil {
			if n, err := strconv.Atoi(v.(string)); err == nil { ms = n }
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(ms)*time.Millisecond)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
