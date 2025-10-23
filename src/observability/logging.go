package observability

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func InitLogger(service string) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	Log = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("svc", service).
		Logger()
}
