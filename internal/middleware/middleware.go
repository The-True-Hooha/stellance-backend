package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/The-True-Hooha/stellance-backend.git/pkg/logger"
	"github.com/google/uuid"
)

type ContextKey string
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Stack   string `json:"stack,omitempty"`
}

const (
	LoggerKey      ContextKey = "logger"
	CorrelationKey ContextKey = "correlation_id"
)

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return logger.Logger().With("warning", "missing logger in context")
}

func WriteLoggerToConText(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationId := r.Header.Get("X-Correlation-Id")
		if correlationId == "" {
			correlationId = uuid.New().String()
		}

		w.Header().Set("X-Correlation-Id", correlationId)
		log := logger.Logger().With("correlation-Id", correlationId)

		ctx := context.WithValue(r.Context(), LoggerKey, log)

		rw := responseWriter(w)
		start := time.Now()
		log.Info(
			"Request started: Method=%s, Path=%s, RemoteAddr=%s, UserAgent=%s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
		)

		h.ServeHTTP(rw, r.WithContext(ctx))
		duration := time.Since(start)

		logString := fmt.Sprintf("Request Response: Method=%s, Path=%s, Status=%d, Duration=%s, Size=%d", r.Method, r.URL.Path, rw.status, duration.String(), rw.size)

		if rw.status >= 500 {
			log.Error(logString)
		} else if rw.status >= 400 {
			log.Debug(logString)
		} else {
			log.Info(logString)
		}
	})
}

type responseWriterS struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

func responseWriter(w http.ResponseWriter) *responseWriterS {
	return &responseWriterS{
		status:         http.StatusOK,
		ResponseWriter: w,
		size:           0,
		written:        false,
	}
}

func (responseWrapper *responseWriterS) WriteHeader(code int) {
	responseWrapper.status = code
	if !responseWrapper.written {
		responseWrapper.ResponseWriter.WriteHeader(code)
		responseWrapper.written = true
	}
}

func (rw *responseWriterS) Write(data []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	log := logger.Logger()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := responseWriter(w)
		defer func() {
			if err := recover(); err != nil {
				env := os.Getenv("STAGE")
				errResponse := ErrorResponse{
					Status:  http.StatusInternalServerError,
					Message: "Internal Server Error",
				}
				dev := env == "dev" || env != "prod"

				if dev {
					errResponse.Error = fmt.Sprintf("%v", err)
					errResponse.Stack = string(debug.Stack())
				} else {
					errResponse.Error = "An unexpected error occurred."
				}

				log.Error("[ERROR] Recovered from panic: %v\n%s", err, debug.Stack())

				responseBytes, _ := json.Marshal(errResponse)
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(responseBytes)
			}

		}()
		next.ServeHTTP(rw, r)
	})
}
