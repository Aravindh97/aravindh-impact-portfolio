package app

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestExactlySixProjectsAndStableSignals(t *testing.T) {
	if len(Products()) != 6 {
		t.Fatalf("projects=%d", len(Products()))
	}
	at := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	for _, slug := range SortedSlugs() {
		a, ok := Signals(slug, at)
		b, _ := Signals(slug, at.Add(3*time.Minute))
		if !ok || len(a.Signals) != 4 || !reflect.DeepEqual(a.Signals, b.Signals) {
			t.Errorf("%s signals", slug)
		}
	}
	if _, ok := Signals("missing", at); ok {
		t.Error("unknown slug accepted")
	}
}

func TestTerraZeroScenariosExposeUncertainty(t *testing.T) {
	low, _ := terraScenario(TerraRequest{"earthquake", 3, 100000, 80})
	high, err := terraScenario(TerraRequest{"earthquake", 8, 100000, 20})
	if err != nil || high["estimated_exposed_people"].(float64) <= low["estimated_exposed_people"].(float64) {
		t.Fatalf("impact not monotonic: %v", err)
	}
	if high["uncertainty"] == "" || len(high["zones"].([]ImpactZone)) != 3 {
		t.Error("missing uncertainty/zones")
	}
	if _, err := terraScenario(TerraRequest{"weapon", 10, 1, 1}); err == nil {
		t.Error("unsupported hazard accepted")
	}
}

func TestNuclearScenarioDecaysAndKeepsCivilBoundary(t *testing.T) {
	now, _ := nuclearScenario(NuclearRequest{50, 12, 10, 0, 70})
	later, err := nuclearScenario(NuclearRequest{50, 12, 10, 1000, 70})
	if err != nil || later["scenario_index"].(float64) >= now["scenario_index"].(float64) {
		t.Error("scenario did not decay")
	}
	if !strings.Contains(later["notice"].(string), "not a dose calculation") {
		t.Error("missing safety boundary")
	}
	if len(nuclearCases) != 3 {
		t.Error("historical cases missing")
	}
}

func TestEcoSignalReducesProxyWhileKeepingCoverage(t *testing.T) {
	r, err := ecoOptimize(EcoRequest{CoverageTarget: 65, HabitatSensitivity: 85, QuietHours: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.OptimizedCoverage < r.CoverageTarget || r.OptimizedHabitatProxy > r.CurrentHabitatProxy || r.EnergyReductionPct < 0 {
		t.Fatalf("bad optimization: %+v", r)
	}
	for _, d := range r.Devices {
		if d.Essential && d.OptimizedPower < 30 {
			t.Error("essential gateway dropped below minimum")
		}
	}
	if !strings.Contains(r.Notice, "not a biological dose") {
		t.Error("missing scientific boundary")
	}
}

func TestHTTP(t *testing.T) {
	h := New(context.Background())
	cases := []struct {
		method, path, body string
		code               int
	}{
		{"GET", "/healthz", "", 200}, {"GET", "/v1/catalog", "", 200},
		{"POST", "/v1/terrazero/scenario", `{"hazard":"flood","intensity":6,"population":100000,"responseCapacity":60}`, 200},
		{"GET", "/v1/nuclear/cases", "", 200},
		{"POST", "/v1/nuclear/scenario", `{"releaseIndex":20,"windSpeed":10,"rainfall":5,"hoursSince":12,"monitoringCoverage":80}`, 200},
		{"POST", "/v1/ecosignal/optimize", `{"coverageTarget":65,"habitatSensitivity":80,"quietHours":true}`, 200},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if rec.Code != tc.code || !json.Valid(rec.Body.Bytes()) {
			t.Errorf("%s %s: %d %s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
}
