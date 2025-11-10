package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/mwdev22/rest/cctx"
	"github.com/mwdev22/rest/jsonutil"
	"github.com/mwdev22/rest/utils/errs"
)

type HandlerWithErr func(w http.ResponseWriter, r *http.Request) error

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func colorMethod(method string) string {
	switch method {
	case "GET":
		return colorBlue + method + colorReset
	case "POST":
		return colorGreen + method + colorReset
	case "PUT":
		return colorYellow + method + colorReset
	case "DELETE":
		return colorRed + method + colorReset
	case "PATCH":
		return colorCyan + method + colorReset
	case "OPTIONS":
		return colorCyan + method + colorReset
	default:
		return method
	}
}

func colorStatus(status int) string {
	statusStr := fmt.Sprintf("%v", status)
	switch {
	case status >= 200 && status < 300:
		return colorGreen + statusStr + colorReset
	case status >= 300 && status < 400:
		return colorYellow + statusStr + colorReset
	case status >= 400 && status < 500:
		return colorRed + statusStr + colorReset
	case status >= 500:
		return colorRed + statusStr + colorReset
	default:
		return statusStr
	}
}

func Logger(next http.Handler) http.Handler {
	before := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			duration := time.Since(before)
			log.Printf("[%s] %s %s %dms",
				colorMethod(r.Method),
				r.RequestURI,
				colorStatus(ww.Status()),
				duration.Milliseconds())
		}()

		next.ServeHTTP(ww, r)
	})
}
func RateLimit(limit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return httprate.LimitByRealIP(limit, windowLength)(next)
	}
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func Wrap(final HandlerWithErr) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := final(w, r); err != nil {
			var e errs.ApiError
			if errors.As(err, &e) {
				jsonutil.Write(w, e.StatusCode, e.Map())
				log.Printf("%sAPI ERROR%s: %s", colorRed, colorReset, e.Log)
			} else {
				jsonutil.Write(w, http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
				log.Printf("%sUNKNOWN ERROR%s: %s", colorRed, colorReset, err.Error())
			}
		}
	}
}

func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := func(r *http.Request) string {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				log.Printf("X-Forwarded-For: %s", xff)
				return strings.TrimSpace(strings.Split(xff, ",")[0])
			}
			if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
				log.Printf("X-Real-IP: %s", xrip)
				return xrip
			}
			host, _, _ := strings.Cut(r.RemoteAddr, ":")
			return host
		}(r)

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), cctx.RealIpKey, ip)),
		)
	})
}

func Internal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _ := r.Context().Value(cctx.RealIpKey).(string)
		if ip != "" {
			log.Printf("Internal route ‑ caller IP: %s", ip)
		}

		if !strings.HasPrefix(ip, "192.168.") && !strings.HasPrefix(ip, "10.") {
			_ = jsonutil.Write(w, http.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
