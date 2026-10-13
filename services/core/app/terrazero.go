package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Aravindh97/aravindh-impact-portfolio/internal/kit"
	"io"
	"math"
	"net/http"
	"sync"
	"time"
)

type HazardEvent struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Magnitude float64   `json:"magnitude"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	DepthKM   float64   `json:"depthKm"`
	Time      time.Time `json:"time"`
	Source    string    `json:"source"`
	URL       string    `json:"url"`
}
type usgsFeed struct {
	Features []struct {
		ID         string `json:"id"`
		Properties struct {
			Mag   float64 `json:"mag"`
			Place string  `json:"place"`
			Time  int64   `json:"time"`
			URL   string  `json:"url"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}
type eventCache struct {
	mu     sync.Mutex
	at     time.Time
	events []HazardEvent
}

func (c *eventCache) get(ctx context.Context) ([]HazardEvent, bool) {
	c.mu.Lock()
	if len(c.events) > 0 && time.Since(c.at) < 10*time.Minute {
		x := append([]HazardEvent(nil), c.events...)
		c.mu.Unlock()
		return x, false
	}
	c.mu.Unlock()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/4.5_day.geojson", nil)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		var f usgsFeed
		if json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&f) == nil {
			var out []HazardEvent
			for _, x := range f.Features {
				if len(x.Geometry.Coordinates) >= 3 {
					out = append(out, HazardEvent{x.ID, x.Properties.Place, "earthquake", x.Properties.Mag, x.Geometry.Coordinates[0], x.Geometry.Coordinates[1], x.Geometry.Coordinates[2], time.UnixMilli(x.Properties.Time).UTC(), "USGS Earthquake Hazards Program", x.Properties.URL})
				}
			}
			if len(out) > 0 {
				c.mu.Lock()
				c.events, c.at = out, time.Now()
				c.mu.Unlock()
				return out, false
			}
		}
	}
	fallback := []HazardEvent{{"demo-1", "Public-feed fallback: offshore seismic event", "earthquake", 5.2, 142.1, 38.2, 35, time.Now().UTC(), "Embedded demonstration fallback", "https://earthquake.usgs.gov/"}}
	return fallback, true
}

type TerraRequest struct {
	Hazard           string  `json:"hazard"`
	Intensity        float64 `json:"intensity"`
	Population       int     `json:"population"`
	ResponseCapacity float64 `json:"responseCapacity"`
}
type ImpactZone struct {
	Name        string  `json:"name"`
	RadiusKM    float64 `json:"radiusKm"`
	ImpactIndex float64 `json:"impactIndex"`
	Guidance    string  `json:"guidance"`
}

func terraScenario(x TerraRequest) (map[string]any, error) {
	factors := map[string]float64{"earthquake": 1, "volcano": .85, "tsunami": 1.15, "flood": .72, "wildfire": .65}
	factor, ok := factors[x.Hazard]
	if !ok || x.Intensity < 1 || x.Intensity > 10 || x.Population < 0 || x.Population > 50000000 || x.ResponseCapacity < 0 || x.ResponseCapacity > 100 {
		return nil, fmt.Errorf("invalid scenario")
	}
	base := x.Intensity * factor
	zones := []ImpactZone{{"Severe impact", round(base * 1.8), round(math.Min(100, base*12)), "Official exclusion and life-safety actions only"}, {"Major disruption", round(base * 4.2), round(math.Min(100, base*8)), "Inspect critical infrastructure and evacuation capacity"}, {"Service pressure", round(base * 8), round(math.Min(100, base*4)), "Prepare shelters, hospitals, water and communications"}}
	exposure := math.Round(float64(x.Population) * math.Min(.9, base/14) * (1 - x.ResponseCapacity/180))
	return map[string]any{"hazard": x.Hazard, "zones": zones, "estimated_exposed_people": exposure, "infrastructure": map[string]int{"hospitals": int(base/2) + 1, "roads": int(base * 4), "water_sites": int(base * 2)}, "uncertainty": "wide; illustrative scenario", "notice": "Not an official prediction or evacuation order. Follow authorized local agencies."}, nil
}
func registerTerraZero(mux *http.ServeMux) {
	cache := &eventCache{}
	mux.HandleFunc("GET /v1/terrazero/events", func(w http.ResponseWriter, r *http.Request) {
		events, stale := cache.get(r.Context())
		kit.JSON(w, 200, map[string]any{"events": events, "stale": stale, "attribution": "USGS Earthquake Hazards Program"})
	})
	mux.HandleFunc("POST /v1/terrazero/scenario", func(w http.ResponseWriter, r *http.Request) {
		var in TerraRequest
		if !kit.Decode(w, r, &in) {
			return
		}
		out, err := terraScenario(in)
		if err != nil {
			kit.Error(w, 422, err.Error())
			return
		}
		kit.JSON(w, 200, out)
	})
}
