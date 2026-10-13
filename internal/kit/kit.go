// Package kit holds small shared helpers for every portfolio app: JSON I/O, request limits,
// rate limiting, security headers, a graceful HTTP server and a shared stylesheet.
package kit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// MaxBody is the default request body limit for JSON APIs.
const MaxBody = 64 << 10

// JSON writes v with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes {"error": msg}.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// Decode reads a size-limited JSON body into v, rejecting unknown fields.
// On failure it writes a 400/413 response and returns false.
func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			Error(w, http.StatusRequestEntityTooLarge, "request body too large")
		} else {
			Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		}
		return false
	}
	return true
}

// ClientIP returns the remote host. Forwarding headers are deliberately ignored
// because they are client-controlled unless a trusted proxy rewrites them.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Limiter is a per-client token bucket applied to state-changing requests.
type Limiter struct {
	mu      sync.Mutex
	rate    float64 // tokens per second
	burst   float64
	buckets map[string]*bucket
	now     func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewLimiter(ratePerSec, burst float64) *Limiter {
	return &Limiter{rate: ratePerSec, burst: burst, buckets: map[string]*bucket{}, now: time.Now}
}

// Allow reports whether key may make another request now.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.buckets) > 10000 { // bound memory; dropping state only makes limits more lenient briefly
		l.buckets = map[string]*bucket{}
	}
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Wrap rate-limits non-GET/HEAD requests per client IP.
func (l *Limiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !l.Allow(ClientIP(r)) {
			w.Header().Set("Retry-After", "1")
			Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CSP allows same-origin scripts/styles/fetch, inline code, WebAssembly and blob: previews.
const CSP = "default-src 'none'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; connect-src 'self'; frame-src 'self' blob:; worker-src 'self'; manifest-src 'self'; " +
	"media-src 'self' blob:; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'self'"

// SecurityHeaders adds baseline browser security headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", CSP)
		h.Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(self)")
		next.ServeHTTP(w, r)
	})
}

// Serve runs h on addr until SIGINT/SIGTERM. WriteTimeout is disabled so Server-Sent
// Events can stream; handlers that stream must bound their own lifetime.
func Serve(addr string, h http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	slog.Info("listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Main is the standard entrypoint for a standalone app: loopback by default, ADDR to override.
func Main(name string, newApp func(context.Context) http.Handler) {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h := SecurityHeaders(NewLimiter(5, 20).Wrap(newApp(ctx)))
	slog.Info("starting", "app", name)
	if err := Serve(addr, h); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// SSE prepares w for Server-Sent Events and returns a send function.
func SSE(w http.ResponseWriter) (send func(event string, v any) error, ok bool) {
	rc := http.NewResponseController(w)
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if err := rc.Flush(); err != nil {
		return nil, false
	}
	return func(event string, v any) error {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b); err != nil {
			return err
		}
		return rc.Flush()
	}, true
}
