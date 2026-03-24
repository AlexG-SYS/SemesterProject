package middleware

import (
	"compress/gzip"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AlexG-SYS/semesterproject/internal/helpers"
	"golang.org/x/time/rate"
)

type Middleware struct {
	App            *helpers.Application
	LimiterRPS     float64
	LimiterBurst   int
	LimiterEnabled bool
	TrustedOrigins []string
}

// Helper to capture status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (mw *metricsResponseWriter) WriteHeader(code int) {
	mw.statusCode = code
	mw.ResponseWriter.WriteHeader(code)
}

func New(app *helpers.Application) *Middleware {
	return &Middleware{App: app}
}

// logger
func (m *Middleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.App.Logger.Info("LOGGER: Received request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)

	})
}

// Metrics Middleware
func (m *Middleware) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Increment Total Requests
		m.App.TotalRequests.Add(1)

		// 2. Wrap the writer to capture the status code
		rec := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		// 3. Record Latency
		duration := time.Since(start)
		m.App.TotalLatency.Add(uint64(duration))

		// 4. Record Errors (any 4xx or 5xx)
		if rec.statusCode >= 400 {
			m.App.TotalErrors.Add(1)
		}

		// 5. Record Route Hits
		actualPath := r.URL.Path
		val, _ := m.App.RouteHits.LoadOrStore(actualPath, &atomic.Uint64{})
		val.(*atomic.Uint64).Add(1)
	})
}

func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	if !m.LimiterEnabled {
		return next
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP, excluding the port
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr // Fallback
		}

		mu.Lock()
		if _, found := clients[host]; !found {
			// This allows 2 requests per second with a burst of 6
			clients[host] = rate.NewLimiter(rate.Limit(m.LimiterRPS), m.LimiterBurst)
		}

		if !clients[host].Allow() {
			mu.Unlock()
			w.Header().Set("Retry-After", "1") // Inform client to wait 1 second
			m.App.ErrorJSON(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// CORS Middleware
func (m *Middleware) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		m.App.Logger.Info("CORS check", "request_origin", origin, "trusted", m.TrustedOrigins)
		if origin != "" {
			for i := range m.TrustedOrigins {
				if origin == m.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// Handle Preflight
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, POST, GET")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Response Compression (Gzip)
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (m *Middleware) Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Check if the client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Prepare the gzip writer
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// 3. Wrap the response writer and proceed
		gzw := gzipResponseWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(gzw, r)
	})
}
