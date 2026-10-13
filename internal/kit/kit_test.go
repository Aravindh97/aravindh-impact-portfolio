package kit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewLimiter(1, 2)
	l.now = func() time.Time { return now }
	if !l.Allow("a") || !l.Allow("a") || l.Allow("a") {
		t.Fatal("burst of 2 expected")
	}
	if !l.Allow("b") {
		t.Fatal("keys must be independent")
	}
	now = now.Add(time.Second)
	if !l.Allow("a") {
		t.Fatal("token should refill after 1s")
	}
}

func TestWrapOnlyLimitsWrites(t *testing.T) {
	l := NewLimiter(0, 1)
	h := l.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	codes := []int{}
	for _, m := range []string{"POST", "POST", "GET"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(m, "/", nil))
		codes = append(codes, rec.Code)
	}
	if codes[0] != 200 || codes[1] != 429 || codes[2] != 200 {
		t.Fatalf("codes = %v", codes)
	}
}

func TestDecode(t *testing.T) {
	var v struct{ A int }
	rec := httptest.NewRecorder()
	if Decode(rec, httptest.NewRequest("POST", "/", strings.NewReader(`{"A":1,"B":2}`)), &v) || rec.Code != 400 {
		t.Fatal("unknown field should be rejected")
	}
	rec = httptest.NewRecorder()
	big := `{"A":1,"pad":"` + strings.Repeat("x", MaxBody) + `"}`
	if Decode(rec, httptest.NewRequest("POST", "/", strings.NewReader(big)), &v) || rec.Code != 413 {
		t.Fatalf("oversized body: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	if !Decode(rec, httptest.NewRequest("POST", "/", strings.NewReader(`{"A":3}`)), &v) || v.A != 3 {
		t.Fatal("valid body should decode")
	}
}
