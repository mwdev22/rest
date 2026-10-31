package middleware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/mwdev22/logging"
	"github.com/mwdev22/rest/cctx"
	"github.com/mwdev22/rest/jsonutil"
	"github.com/mwdev22/rest/utils/errs"
)

type ApiHandler func(w http.ResponseWriter, r *http.Request) error
type MiddlewareToChain func(next http.Handler) http.Handler

// allows handlers to return errors
func Wrap(final ApiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := final(w, r); err != nil {
			var e errs.ApiError
			l := logging.FromContext(r.Context())
			if errors.As(err, &e) {
				jsonutil.Write(w, e.StatusCode, e.Map())
				l.Printf("%sAPI ERROR%s: %s", colorRed, colorReset, e.Log)
			} else {
				jsonutil.Write(w, http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
				l.Printf("%sUNKNOWN ERROR%s: %s", colorRed, colorReset, err.Error())
			}
		}
	}
}

func Pagination(defaultLimit int, maxLimit int) MiddlewareToChain {
	const minLimit = 1
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				pageQ := r.URL.Query().Get("page")
				limitQ := r.URL.Query().Get("limit")
				limit, offset := defaultLimit, 0

				if limitQ != "" {
					l, err := strconv.Atoi(limitQ)
					if err != nil {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "limit must be a number",
						})
						return
					}
					if l < minLimit {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "limit must be at least 1",
						})
						return
					}
					if l > maxLimit {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": fmt.Sprintf("limit must not exceed %d", maxLimit),
						})
						return
					}
					limit = l
				}

				if pageQ != "" {
					p, err := strconv.Atoi(pageQ)
					if err != nil {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "page must be a number",
						})
						return
					}
					if p < 1 {
						jsonutil.Write(w, http.StatusBadRequest, map[string]string{
							"error": "page must be at least 1",
						})
						return
					}
					offset = (p - 1) * limit
				}

				ctx := context.WithValue(r.Context(), cctx.Offset, offset)
				ctx = context.WithValue(ctx, cctx.Limit, limit)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
	}
}

func Logger(l logging.Logger) MiddlewareToChain {
	if l == nil {
		l = logging.DefaultLogger()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			before := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			ctx := logging.ToContext(r.Context(), l)
			defer func() {
				duration := time.Since(before)
				l.Printf("[%s] %s %s %.2fms",
					colorMethod(r.Method),
					r.RequestURI,
					colorStatus(ww.Status()),
					float64(duration.Microseconds())/1000.0)
			}()
			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}
func RateLimit(limit int, windowLength time.Duration) MiddlewareToChain {
	return func(next http.Handler) http.Handler {
		return httprate.Limit(limit, windowLength, httprate.WithKeyFuncs(keyByStoredIP))(next)
	}
}

// httprate.KeyByRealIP trusts client headers, so key on what RealIP resolved
func keyByStoredIP(r *http.Request) (string, error) {
	if ip := cctx.RealIP(r.Context()); ip != "" {
		return ip, nil
	}
	return httprate.KeyByIP(r)
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

// RealIP stores the TCP peer address and ignores forwarding headers, use RealIPFrom behind a proxy
func RealIP(next http.Handler) http.Handler {
	return RealIPFrom()(next)
}

// RealIPFrom trusts X-Forwarded-For / X-Real-IP only from peers in trusted,
// taking the right-most X-Forwarded-For entry that is not a trusted proxy
func RealIPFrom(trusted ...netip.Prefix) MiddlewareToChain {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r, trusted)
			next.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), cctx.RealIpKey, ip)),
			)
		})
	}
}

func clientIP(r *http.Request, trusted []netip.Prefix) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	peer, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}
	peer = peer.Unmap()
	if !isTrusted(peer, trusted) {
		return peer.String()
	}

	// proxies append to XFF, so anything left of the first untrusted hop is client-controlled
	var hops []string
	for _, h := range r.Header.Values("X-Forwarded-For") {
		hops = append(hops, strings.Split(h, ",")...)
	}
	client := peer
	for i := len(hops) - 1; i >= 0; i-- {
		hop, err := netip.ParseAddr(strings.TrimSpace(hops[i]))
		if err != nil {
			return client.String()
		}
		client = hop.Unmap()
		if !isTrusted(client, trusted) {
			return client.String()
		}
	}
	if len(hops) == 0 {
		if xrip, err := netip.ParseAddr(strings.TrimSpace(r.Header.Get("X-Real-IP"))); err == nil {
			return xrip.Unmap().String()
		}
	}
	return client.String()
}

func isTrusted(ip netip.Addr, trusted []netip.Prefix) bool {
	for _, p := range trusted {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

func Internal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _ := r.Context().Value(cctx.RealIpKey).(string)
		if ip != "" {
			l := logging.FromContext(r.Context())
			l.Printf("internal route ‑ caller IP: %s", ip)
		}

		addr, err := netip.ParseAddr(ip)
		if err != nil || !(addr.IsPrivate() || addr.IsLoopback()) {
			_ = jsonutil.Write(w, http.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
