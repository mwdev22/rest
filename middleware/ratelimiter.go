package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*client)
)

func cleanupClients() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, c := range clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

func getClient(ip string, r rate.Limit, b int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if c, exists := clients[ip]; exists {
		c.lastSeen = time.Now()
		return c.limiter
	}

	limiter := rate.NewLimiter(r, b)
	clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
	return limiter
}

// r = requests per second
// b = burst size (requests at once)
func RateLimiter(r rate.Limit, b int) func(http.Handler) http.Handler {
	go cleanupClients()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				http.Error(w, "unable to determine IP", http.StatusInternalServerError)
				return
			}

			limiter := getClient(ip, r, b)
			if !limiter.Allow() {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
