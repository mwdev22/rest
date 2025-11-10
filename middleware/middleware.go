package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/mwdev22/rest/cctx"
	"github.com/mwdev22/rest/utils/errs"
	"github.com/mwdev22/rest/utils/jsonutil"
)

type appHandler func(w http.ResponseWriter, r *http.Request) error

func Logger(next http.Handler) http.Handler {
	before := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			log.Printf("[%s] %s %d %s", r.Method, r.RequestURI, ww.Status(), time.Since(before))
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

func Wrap(final appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := final(w, r); err != nil {
			var e errs.ApiError
			if errors.As(err, &e) {
				jsonutil.Write(w, e.StatusCode, map[string]string{
					"error": e.Error(),
				})
			} else {
				jsonutil.Write(w, http.StatusInternalServerError, "internal server error")
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
