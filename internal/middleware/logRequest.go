package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/google/uuid"
)

func LogRequest(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestLogger := logger

			// In case some other part of our system calls another handler
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				id, err := uuid.NewV7()
				if err != nil {
					logger.Error("failed to generate request_id")
					// don't want to crash for this
					requestID = fmt.Sprintf("%d", time.Now().UnixNano())
				}
				requestID = id.String()
			}

			requestLogger = requestLogger.With(
				slog.Group("request",
					slog.String("id", requestID),
					slog.String("method", r.Method),
					slog.String("uri", r.URL.String()),
				),
			)
			ctx := context.WithValue(r.Context(), "logger", requestLogger)
			ctx = context.WithValue(ctx, "request_id", requestID)

			respMetrics := httpsnoop.CaptureMetricsFn(w, func(w http.ResponseWriter) {
				next.ServeHTTP(w, r.WithContext(ctx))
			})

			requestLogger.Info("handled request",
				slog.Group("request",
					slog.Int("code", respMetrics.Code),
					slog.Int64("duration_ms", respMetrics.Duration.Milliseconds()),
					slog.Int64("size", respMetrics.Written),
					slog.String("id", requestID),
					slog.String("ip", requestGetRemoteAddress(r)),
					slog.String("method", r.Method),
					slog.String("referer", r.Header.Get("Referer")),
					slog.String("uri", r.URL.String()),
					slog.String("user_agent", r.Header.Get("User-Agent")),
				),
			)
		}
		return http.HandlerFunc(fn)
	}
}

// requestGetRemoteAddress returns ip address of the client making the request,
// taking into account http proxies
func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		if colonLoc := strings.LastIndex(r.RemoteAddr, ":"); colonLoc == -1 {
			return r.RemoteAddr
		} else {
			return r.RemoteAddr[:colonLoc]
		}
	}
	if hdrForwardedFor != "" {
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIP
}
