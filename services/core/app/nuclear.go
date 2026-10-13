package app

import (
	"fmt"
	"math"
	"net/http"

	"github.com/Aravindh97/aravindh-impact-portfolio/internal/kit"
)

type NuclearRequest struct {
	ReleaseIndex       float64 `json:"releaseIndex"`
	WindSpeed          float64 `json:"windSpeed"`
	Rainfall           float64 `json:"rainfall"`
	HoursSince         int     `json:"hoursSince"`
	MonitoringCoverage float64 `json:"monitoringCoverage"`
}

var nuclearCases = []map[string]any{
	{"name": "Three Mile Island", "year": 1979, "country": "United States", "focus": "Communication, instrumentation and emergency planning", "source": "https://www.nrc.gov/reading-rm/doc-collections/fact-sheets/3mile-isle"},
	{"name": "Chernobyl", "year": 1986, "country": "Ukraine (then Soviet Union)", "focus": "Long-term contamination, relocation and international monitoring", "source": "https://www.iaea.org/topics/chornobyl"},
	{"name": "Fukushima Daiichi", "year": 2011, "country": "Japan", "focus": "Compound natural hazards, evacuation and long-term recovery", "source": "https://www.iaea.org/topics/response/fukushima-daiichi-nuclear-accident"},
}

func nuclearScenario(x NuclearRequest) (map[string]any, error) {
	if x.ReleaseIndex < 0 || x.ReleaseIndex > 100 || x.WindSpeed < 0 || x.WindSpeed > 100 || x.Rainfall < 0 || x.Rainfall > 300 || x.HoursSince < 0 || x.HoursSince > 8760 || x.MonitoringCoverage < 0 || x.MonitoringCoverage > 100 {
		return nil, fmt.Errorf("scenario values out of range")
	}
	decay := math.Exp(-float64(x.HoursSince) / 240)
	transport := 1 + x.WindSpeed/30
	deposition := 1 + x.Rainfall/120
	index := math.Min(100, x.ReleaseIndex*decay*transport*deposition)
	zones := []map[string]any{
		{"name": "Immediate monitoring priority", "radius_km": round(math.Sqrt(index) * 1.2), "index": round(index)},
		{"name": "Food/water sampling priority", "radius_km": round(math.Sqrt(index) * 3), "index": round(index * .55)},
		{"name": "Long-term environmental survey", "radius_km": round(math.Sqrt(index) * 6), "index": round(index * .25)},
	}
	recovery := []map[string]any{
		{"phase": "hours", "action": "Follow official shelter/evacuation communication and deploy certified monitoring"},
		{"phase": "days", "action": "Sample food, water and soil through authorized laboratories"},
		{"phase": "years", "action": "Track exposure, cleanup, health support and community recovery transparently"},
	}
	return map[string]any{
		"scenario_index":     index,
		"zones":              zones,
		"monitoring_gap_pct": round(100 - x.MonitoringCoverage),
		"recovery":           recovery,
		"notice":             "Educational civil-safety scenario only. It is not a dose calculation, medical instruction or official emergency guidance.",
	}, nil
}

func registerNuclear(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/nuclear/cases", func(w http.ResponseWriter, _ *http.Request) {
		kit.JSON(w, 200, map[string]any{"cases": nuclearCases})
	})
	mux.HandleFunc("POST /v1/nuclear/scenario", func(w http.ResponseWriter, r *http.Request) {
		var in NuclearRequest
		if !kit.Decode(w, r, &in) {
			return
		}
		out, err := nuclearScenario(in)
		if err != nil {
			kit.Error(w, 422, err.Error())
			return
		}
		kit.JSON(w, 200, out)
	})
}
