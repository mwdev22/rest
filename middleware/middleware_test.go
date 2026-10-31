package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/mwdev22/rest/cctx"
	"github.com/mwdev22/rest/utils/errs"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name           string
		handler        ApiHandler
		expectedStatus int
		expectedBody   string
		checkJSON      bool
		expectedError  string
	}{
		{
			name: "ok - no error",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "error - invalid json",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.InvalidJson(errors.New("unexpected token"))
			},
			expectedStatus: http.StatusBadRequest,
			checkJSON:      true,
			expectedError:  "invalid json",
		},
		{
			name: "error - invalid path param",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.InvalidPathParam("id")
			},
			expectedStatus: http.StatusBadRequest,
			checkJSON:      true,
			expectedError:  "invalid path param: id",
		},
		{
			name: "error - invalid query param",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.InvalidQueryParam("id")
			},
			expectedStatus: http.StatusBadRequest,
			checkJSON:      true,
			expectedError:  "invalid query param: id",
		},
		{
			name: "error - not found",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errs.NotFound("object not found")
			},
			expectedStatus: http.StatusNotFound,
			checkJSON:      true,
			expectedError:  "not found",
		},
		{
			name: "generic error",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("something went wrong")
			},
			expectedStatus: http.StatusInternalServerError,
			checkJSON:      true,
			expectedError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Wrap(tt.handler)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkJSON {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("expected error message '%s', got '%s'", tt.expectedError, response["error"])
				}
			} else if tt.expectedBody != "" {
				if w.Body.String() != tt.expectedBody {
					t.Errorf("expected body '%s', got '%s'", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

func TestRealIP(t *testing.T) {
	proxies := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("fd00::/8"),
	}

	tests := []struct {
		name       string
		trusted    []netip.Prefix
		remoteAddr string
		xff        []string
		xRealIP    string
		expectedIP string
	}{
		{name: "RemoteAddr fallback", remoteAddr: "192.0.2.1:12345", expectedIP: "192.0.2.1"},
		{name: "IPv6 RemoteAddr", remoteAddr: "[::1]:8080", expectedIP: "::1"},
		{name: "IPv4-mapped IPv6 RemoteAddr", remoteAddr: "[::ffff:192.0.2.1]:8080", expectedIP: "192.0.2.1"},
		{name: "spoofed XFF from untrusted peer ignored", remoteAddr: "203.0.113.9:1234", xff: []string{"192.168.1.1"}, expectedIP: "203.0.113.9"},
		{name: "spoofed X-Real-IP from untrusted peer ignored", remoteAddr: "203.0.113.9:1234", xRealIP: "192.168.1.1", expectedIP: "203.0.113.9"},
		{name: "XFF from peer outside trusted set ignored", trusted: proxies, remoteAddr: "203.0.113.9:1234", xff: []string{"192.168.1.1"}, expectedIP: "203.0.113.9"},
		{name: "trusted proxy - single hop", trusted: proxies, remoteAddr: "10.0.0.2:1234", xff: []string{"198.51.100.7"}, expectedIP: "198.51.100.7"},
		{name: "trusted proxy chain - right-most untrusted wins", trusted: proxies, remoteAddr: "10.0.0.2:1234", xff: []string{"192.168.1.1, 198.51.100.7, 10.0.0.5"}, expectedIP: "198.51.100.7"},
		{name: "trusted proxy chain - multiple XFF headers", trusted: proxies, remoteAddr: "10.0.0.2:1234", xff: []string{"192.168.1.1", "198.51.100.7, 10.0.0.5"}, expectedIP: "198.51.100.7"},
		{name: "trusted IPv6 proxy", trusted: proxies, remoteAddr: "[fd00::1]:443", xff: []string{"2001:db8::7"}, expectedIP: "2001:db8::7"},
		{name: "trusted proxy - invalid XFF entry stops walk", trusted: proxies, remoteAddr: "10.0.0.2:1234", xff: []string{"198.51.100.7, garbage"}, expectedIP: "10.0.0.2"},
		{name: "trusted proxy - all hops trusted", trusted: proxies, remoteAddr: "10.0.0.2:1234", xff: []string{"10.0.0.9"}, expectedIP: "10.0.0.9"},
		{name: "trusted proxy - X-Real-IP without XFF", trusted: proxies, remoteAddr: "10.0.0.2:1234", xRealIP: "198.51.100.42", expectedIP: "198.51.100.42"},
		{name: "trusted proxy - no headers", trusted: proxies, remoteAddr: "10.0.0.2:1234", expectedIP: "10.0.0.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RealIPFrom(tt.trusted...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ip, ok := r.Context().Value(cctx.RealIpKey).(string)
				if !ok {
					t.Error("RealIpKey not found in context")
					return
				}
				if ip != tt.expectedIP {
					t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for _, v := range tt.xff {
				req.Header.Add("X-Forwarded-For", v)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestRealIP_DefaultIgnoresHeaders(t *testing.T) {
	const peer = "203.0.113.9"

	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := cctx.RealIP(r.Context()); ip != peer {
			t.Errorf("expected IP %s, got %s", peer, ip)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = peer + ":1234"
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	req.Header.Set("X-Real-IP", "192.168.1.1")
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestInternal_SpoofedHeaders(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		expectedStatus int
	}{
		{name: "public peer with spoofed private XFF denied", remoteAddr: "203.0.113.9:1234", expectedStatus: http.StatusForbidden},
		{name: "private peer allowed", remoteAddr: "192.168.1.10:1234", expectedStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RealIP(Internal(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})))

			req := httptest.NewRequest(http.MethodGet, "/internal", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Forwarded-For", "192.168.1.1")
			req.Header.Set("X-Real-IP", "192.168.1.1")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRateLimit_IgnoresSpoofedHeaders(t *testing.T) {
	const (
		limit  = 2
		window = time.Minute
	)

	tests := []struct {
		name       string
		withRealIP bool
	}{
		{name: "without RealIP keys on peer", withRealIP: false},
		{name: "with RealIP keys on stored IP", withRealIP: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler http.Handler = RateLimit(limit, window)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			if tt.withRealIP {
				handler = RealIP(handler)
			}

			var lastCode int
			for i := 0; i <= limit; i++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = "203.0.113.9:1234"
				req.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i))
				req.Header.Set("X-Real-IP", fmt.Sprintf("198.51.100.%d", i))
				req.Header.Set("True-Client-IP", fmt.Sprintf("198.51.100.%d", i))
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				lastCode = w.Code
			}

			if lastCode != http.StatusTooManyRequests {
				t.Errorf("expected status %d after rotating headers, got %d", http.StatusTooManyRequests, lastCode)
			}
		})
	}
}

func TestInternal(t *testing.T) {
	tests := []struct {
		name           string
		ip             string
		expectedStatus int
		expectedError  string
		shouldAllow    bool
	}{
		{
			name:           "allow 192.168 network",
			ip:             "192.168.1.100",
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},

		{
			name:           "block public IP",
			ip:             "203.0.113.1",
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			shouldAllow:    false,
		},
		{
			name:           "allow 10.x network",
			ip:             "10.1.2.3",
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "allow loopback",
			ip:             "127.0.0.1",
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "allow IPv6 loopback",
			ip:             "::1",
			expectedStatus: http.StatusOK,
			shouldAllow:    true,
		},
		{
			name:           "block 192.168 lookalike",
			ip:             "192.168.evil.com",
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			shouldAllow:    false,
		},
		{
			name:           "block empty IP",
			ip:             "",
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			shouldAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Internal(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("allowed"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/internal", nil)
			req = req.WithContext(context.WithValue(req.Context(), cctx.RealIpKey, tt.ip))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d for IP %s, got %d", tt.expectedStatus, tt.ip, w.Code)
			}

			if tt.shouldAllow {
				if w.Body.String() != "allowed" {
					t.Errorf("expected body 'allowed', got %s", w.Body.String())
				}
			} else {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("expected error '%s', got '%s'", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestLogger(t *testing.T) {
	loggerMiddleware := Logger(nil)
	handler := loggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("logged"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "logged" {
		t.Errorf("expected body 'logged', got %s", w.Body.String())
	}
}

func TestRecoverer(t *testing.T) {
	handler := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 after panic, got %d", w.Code)
	}
}

func TestColorMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{"GET", colorBlue + "GET" + colorReset},
		{"POST", colorGreen + "POST" + colorReset},
		{"PUT", colorYellow + "PUT" + colorReset},
		{"DELETE", colorRed + "DELETE" + colorReset},
		{"PATCH", colorCyan + "PATCH" + colorReset},
		{"OPTIONS", colorCyan + "OPTIONS" + colorReset},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := colorMethod(tt.method)
			if result != tt.expected {
				t.Errorf("colorMethod(%s) = %q, want %q", tt.method, result, tt.expected)
			}
		})
	}
}

func TestColorStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{200, colorGreen + "200" + colorReset},
		{299, colorGreen + "299" + colorReset},
		{301, colorYellow + "301" + colorReset},
		{404, colorRed + "404" + colorReset},
		{500, colorRed + "500" + colorReset},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.status), func(t *testing.T) {
			result := colorStatus(tt.status)
			if result != tt.expected {
				t.Errorf("colorStatus(%d) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}
