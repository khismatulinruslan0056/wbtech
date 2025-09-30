package middleware

import (
	"L0/internal/transport/dto"
	"L0/internal/transport/httpserver/common"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"time"
)

var reOrderID = regexp.MustCompile(`^[A-Za-z0-9_-]{12,64}$`)

func ValidateOrderID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("orderID")
		if orderID == "" || !reOrderID.MatchString(orderID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			resp := dto.ErrorResponse{ErrMsg: "invalid or empty orderID"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		ctx := context.WithValue(r.Context(), common.OrderIDKey, orderID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			if rw.statusCode >= 400 {
				log.Error("log Middleware",
					"Method", r.Method,
					"Path", r.URL.Path,
					"statusCode", rw.statusCode,
					"duration", duration,
					"RemoteAddr", r.RemoteAddr,
					"UserAgent", r.UserAgent(),
					"errMsg", rw.errMsg,
				)
			} else {
				log.Info("log Middleware",
					"Method", r.Method,
					"Path", r.URL.Path,
					"statusCode", rw.statusCode,
					"duration", duration,
					"RemoteAddr", r.RemoteAddr,
					"UserAgent", r.UserAgent(),
				)
			}
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	errMsg     string
	statusCode int
	written    bool
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.ResponseWriter.WriteHeader(code)
		w.written = true
	}
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	if w.statusCode >= 400 {
		const maxLen = 512
		if len(b) > maxLen {
			w.errMsg = string(b[:maxLen]) + "..."
		} else {
			w.errMsg = string(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

func RecoverPanic(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					log.Error("panic", "panic", v)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
