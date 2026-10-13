// Package app exposes Go operational APIs for TerraZero, Nuclear Legacy and EcoSignal.
package app

import (
	"context"
	"github.com/Aravindh97/aravindh-impact-portfolio/internal/kit"
	"hash/fnv"
	"math"
	"net/http"
	"sort"
	"time"
)

type Product struct {
	Slug     string   `json:"slug"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Signals  []string `json:"signals"`
}
type Signal struct {
	Label    string  `json:"label"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Trend    string  `json:"trend"`
	Severity string  `json:"severity"`
}
type SignalResponse struct {
	GeneratedAt time.Time `json:"generatedAt"`
	Simulated   bool      `json:"simulated"`
	Signals     []Signal  `json:"signals"`
}

var catalog = []Product{{"openschool", "OpenSchool Enterprise", "Enterprise open learning", []string{"Learner mastery", "Curriculum coverage", "Content review", "Offline sync"}}, {"faminelens", "FamineLens", "Historical famine analysis", []string{"Historical evidence", "Price stress", "Rainfall stress", "Response capacity"}}, {"kindred", "Kindred", "Animal welfare intelligence", []string{"Urgent reports", "Shelter capacity", "Rescue wait", "Sensitive locations"}}, {"terrazero", "TerraZero", "Multi-hazard impact zones", []string{"Hazard intensity", "Population exposure", "Infrastructure risk", "Response capacity"}}, {"nuclear-legacy", "Nuclear Legacy", "Civil nuclear safety", []string{"Monitoring status", "Scenario uncertainty", "Recovery phase", "Public guidance"}}, {"ecosignal", "EcoSignal", "Wildlife-aware RF efficiency", []string{"Network coverage", "Habitat RF proxy", "Energy reduction", "Calibration state"}}}
var units = map[string][]string{"openschool": {"%", "%", "items", "%"}, "faminelens": {"sources", "index", "index", "%"}, "kindred": {"reports", "places", "minutes", "coarse cells"}, "terrazero": {"index", "k people", "%", "%"}, "nuclear-legacy": {"stations", "range", "years", "notices"}, "ecosignal": {"%", "relative proxy", "%", "status"}}

func Products() []Product { return append([]Product(nil), catalog...) }
func SortedSlugs() []string {
	r := make([]string, 0, len(catalog))
	for _, p := range catalog {
		r = append(r, p.Slug)
	}
	sort.Strings(r)
	return r
}
func Signals(slug string, at time.Time) (SignalResponse, bool) {
	var p *Product
	for i := range catalog {
		if catalog[i].Slug == slug {
			p = &catalog[i]
			break
		}
	}
	if p == nil {
		return SignalResponse{}, false
	}
	window := at.UTC().Unix() / 300
	out := SignalResponse{at.UTC(), true, nil}
	for i, label := range p.Signals {
		h := fnv.New64a()
		_, _ = h.Write([]byte(slug))
		_, _ = h.Write([]byte{byte(i)})
		seed := float64((h.Sum64() + uint64(window)*uint64(i+7)) % 10000)
		value := math.Round((8+math.Abs(math.Sin(seed))*84)*10) / 10
		sev := "low"
		if value >= 70 {
			sev = "high"
		} else if value >= 42 {
			sev = "medium"
		}
		out.Signals = append(out.Signals, Signal{label, value, units[slug][i], []string{"down", "stable", "up"}[int(seed)%3], sev})
	}
	return out, true
}
func New(_ context.Context) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		kit.JSON(w, 200, map[string]any{"status": "ok", "projects": 6})
	})
	mux.HandleFunc("GET /v1/catalog", func(w http.ResponseWriter, _ *http.Request) { kit.JSON(w, 200, Products()) })
	mux.HandleFunc("GET /v1/signals/{slug}", func(w http.ResponseWriter, r *http.Request) {
		out, ok := Signals(r.PathValue("slug"), time.Now())
		if !ok {
			kit.Error(w, 404, "unknown project")
			return
		}
		kit.JSON(w, 200, out)
	})
	registerTerraZero(mux)
	registerNuclear(mux)
	registerEcoSignal(mux)
	return kit.SecurityHeaders(mux)
}

func round(x float64) float64 { return math.Round(x*10) / 10 }
