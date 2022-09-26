package logging

import (
	"context"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Start() context.Context {
	return WithLoggingContext(context.Background())
}

func WithLoggingContext(ctx context.Context) context.Context {
	zerolog.ErrorFieldName = "error.message"
	level := zerolog.InfoLevel
	levelStr := LogLevel()

	switch strings.ToLower(levelStr) {

	case "err":
		fallthrough
	case "error":
		level = zerolog.ErrorLevel
	case "warn":
		fallthrough
	case "warning":
		level = zerolog.WarnLevel
	case "info":
		fallthrough
	case "information":
		level = zerolog.InfoLevel
	case "trace":
		level = zerolog.TraceLevel
	case "debug":
		level = zerolog.DebugLevel
	default:
		level = zerolog.InfoLevel
	}

	logger := log.Output(os.Stdout)
	logger = logger.Level(level)

	return logger.WithContext(ctx)

}

func LoggerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				t2 := time.Now()

				// Recover and record stack traces in case of a panic
				if rec := recover(); rec != nil {
					log.Error().Str("type", "error").Timestamp().Interface("recover_info", rec).Bytes("debug_stack", debug.Stack()).Msg("error_request")
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				// log end request
				log.Debug().
					Str("type", "access").
					Timestamp().
					Fields(map[string]interface{}{
						"remote_ip":  r.RemoteAddr,
						"host":       r.Host,
						"url":        r.URL.Path,
						"proto":      r.Proto,
						"method":     r.Method,
						"user_agent": r.Header.Get("User-Agent"),
						"status":     ww.Status(),
						"latency_ms": float64(t2.Sub(t1).Nanoseconds()) / 1000000.0,
						"bytes_in":   r.Header.Get("Content-Length"),
						"bytes_out":  ww.BytesWritten(),
					}).Msg("incoming_request")
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func MergeContextKeys(target context.Context, source context.Context, keys ...string) zerolog.Logger {
	logger := *log.Ctx(target)
	for _, k := range keys {
		v, ok := source.Value(k).(string)
		if ok {
			log.Debug().Msgf("Could not find key: %s in context", k)
		}
		logger = logger.With().Str(k, v).Logger()
	}
	return logger
}
