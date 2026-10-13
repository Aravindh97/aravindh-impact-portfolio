package app

import (
	"fmt"
	"math"
	"net/http"

	"github.com/Aravindh97/aravindh-impact-portfolio/internal/kit"
)

type RFDevice struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Kind         string  `json:"kind"`
	Band         string  `json:"band"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	CurrentPower float64 `json:"currentPower"`
	MinPower     float64 `json:"minPower"`
	DutyCycle    float64 `json:"dutyCycle"`
	Controllable bool    `json:"controllable"`
	Essential    bool    `json:"essential"`
}
type EcoRequest struct {
	CoverageTarget     float64 `json:"coverageTarget"`
	HabitatSensitivity float64 `json:"habitatSensitivity"`
	QuietHours         bool    `json:"quietHours"`
}
type EcoDevicePlan struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Band           string  `json:"band"`
	CurrentPower   float64 `json:"currentPower"`
	OptimizedPower float64 `json:"optimizedPower"`
	CurrentDuty    float64 `json:"currentDuty"`
	OptimizedDuty  float64 `json:"optimizedDuty"`
	Recommendation string  `json:"recommendation"`
	Essential      bool    `json:"essential"`
}
type EcoCell struct {
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Current   float64 `json:"current"`
	Optimized float64 `json:"optimized"`
	Habitat   bool    `json:"habitat"`
}
type EcoResult struct {
	CoverageTarget        float64         `json:"coverageTarget"`
	CurrentCoverage       float64         `json:"currentCoverage"`
	OptimizedCoverage     float64         `json:"optimizedCoverage"`
	CurrentHabitatProxy   float64         `json:"currentHabitatProxy"`
	OptimizedHabitatProxy float64         `json:"optimizedHabitatProxy"`
	EnergyReductionPct    float64         `json:"energyReductionPct"`
	Devices               []EcoDevicePlan `json:"devices"`
	Cells                 []EcoCell       `json:"cells"`
	Notice                string          `json:"notice"`
}

var rfDevices = []RFDevice{
	{"ap-north", "North access point", "router", "5 GHz", 22, 24, 100, 25, .72, true, false},
	{"ap-south", "South access point", "router", "2.4 GHz", 68, 68, 85, 20, .66, true, false},
	{"iot-gateway", "Emergency IoT gateway", "gateway", "868/915 MHz", 48, 45, 60, 30, .25, true, true},
	{"laptop-bank", "Shared laptops", "clients", "2.4/5 GHz", 32, 62, 45, 45, .18, false, false},
	{"phone-zone", "Phone charging area", "clients", "cellular/Wi-Fi", 78, 35, 55, 55, .24, false, false},
}

func levelSet(d RFDevice) []float64 {
	if !d.Controllable {
		return []float64{d.CurrentPower}
	}
	return []float64{d.MinPower, math.Round((d.MinPower + d.CurrentPower) / 2), d.CurrentPower}
}
func signalAt(x, y float64, powers map[string]float64) float64 {
	best := 0.0
	for _, d := range rfDevices {
		if !d.Controllable {
			continue
		}
		dist := math.Hypot(x-d.X, y-d.Y)
		v := (powers[d.ID] / 100) / (1 + math.Pow(dist/17, 2))
		if v > best {
			best = v
		}
	}
	return best
}
func exposureAt(x, y float64, powers, duties map[string]float64) float64 {
	v := 0.0
	for _, d := range rfDevices {
		dist := math.Hypot(x-d.X, y-d.Y)
		band := 1.0
		if d.Band == "5 GHz" {
			band = .72
		}
		v += (powers[d.ID] / 100) * duties[d.ID] * band / (1 + math.Pow(dist/11, 2))
	}
	return v
}
func metrics(powers, duties map[string]float64) (coverage, habitat, energy float64) {
	covered, total := 0, 0
	for y := 5; y <= 95; y += 10 {
		for x := 5; x <= 95; x += 10 {
			total++
			if signalAt(float64(x), float64(y), powers) >= .105 {
				covered++
			}
		}
	}
	coverage = 100 * float64(covered) / float64(total)
	for y := 10; y <= 40; y += 10 {
		for x := 65; x <= 95; x += 10 {
			habitat += exposureAt(float64(x), float64(y), powers, duties)
		}
	}
	habitat /= 16
	for _, d := range rfDevices {
		energy += powers[d.ID] * duties[d.ID]
	}
	return
}

func ecoOptimize(in EcoRequest) (EcoResult, error) {
	if in.CoverageTarget < 50 || in.CoverageTarget > 100 || in.HabitatSensitivity < 0 || in.HabitatSensitivity > 100 {
		return EcoResult{}, fmt.Errorf("coverage target 50-100 and habitat sensitivity 0-100 required")
	}
	currentP, currentD := map[string]float64{}, map[string]float64{}
	for _, d := range rfDevices {
		currentP[d.ID] = d.CurrentPower
		currentD[d.ID] = d.DutyCycle
	}
	cc, ch, ce := metrics(currentP, currentD)
	bestScore := math.Inf(1)
	var bestP, bestD map[string]float64
	var bc, bh, be float64
	a, b, c := levelSet(rfDevices[0]), levelSet(rfDevices[1]), levelSet(rfDevices[2])
	for _, p0 := range a {
		for _, p1 := range b {
			for _, p2 := range c {
				p := map[string]float64{}
				duty := map[string]float64{}
				for _, d := range rfDevices {
					p[d.ID] = d.CurrentPower
					duty[d.ID] = d.DutyCycle
				}
				p[rfDevices[0].ID], p[rfDevices[1].ID], p[rfDevices[2].ID] = p0, p1, p2
				if in.QuietHours {
					for _, d := range rfDevices {
						if !d.Essential {
							duty[d.ID] *= .45
						}
					}
				}
				cov, hab, energy := metrics(p, duty)
				if cov+1e-9 < in.CoverageTarget {
					continue
				}
				score := hab*(.5+in.HabitatSensitivity/20) + energy*(1.5-in.HabitatSensitivity/100)
				if score < bestScore {
					bestScore = score
					bestP, bestD = p, duty
					bc, bh, be = cov, hab, energy
				}
			}
		}
	}
	if bestP == nil {
		return EcoResult{}, fmt.Errorf("coverage target cannot be met with available access points")
	}
	res := EcoResult{CoverageTarget: in.CoverageTarget, CurrentCoverage: round(cc), OptimizedCoverage: round(bc), CurrentHabitatProxy: round(ch * 100), OptimizedHabitatProxy: round(bh * 100), EnergyReductionPct: round(100 * (ce - be) / math.Max(ce, 1)), Notice: "Relative RF activity proxy—not a biological dose or wildlife safety certification. Validate with calibrated instruments and qualified ecology/RF experts."}
	for _, d := range rfDevices {
		rec := "Observe only; operating system or user controls this radio"
		if d.Controllable {
			rec = fmt.Sprintf("Set configured power from %.0f%% to %.0f%%", d.CurrentPower, bestP[d.ID])
			if in.QuietHours && !d.Essential {
				rec += " and reduce idle airtime during quiet hours"
			}
		}
		res.Devices = append(res.Devices, EcoDevicePlan{d.ID, d.Name, d.Band, d.CurrentPower, bestP[d.ID], d.DutyCycle, bestD[d.ID], rec, d.Essential})
	}
	for y := 5; y <= 95; y += 10 {
		for x := 5; x <= 95; x += 10 {
			res.Cells = append(res.Cells, EcoCell{x, y, round(exposureAt(float64(x), float64(y), currentP, currentD) * 100), round(exposureAt(float64(x), float64(y), bestP, bestD) * 100), x >= 65 && y <= 45})
		}
	}
	return res, nil
}
func registerEcoSignal(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/ecosignal/devices", func(w http.ResponseWriter, _ *http.Request) {
		kit.JSON(w, 200, map[string]any{"devices": rfDevices, "habitat_zone": "coarse demonstration area"})
	})
	mux.HandleFunc("POST /v1/ecosignal/optimize", func(w http.ResponseWriter, r *http.Request) {
		var in EcoRequest
		if !kit.Decode(w, r, &in) {
			return
		}
		out, err := ecoOptimize(in)
		if err != nil {
			kit.Error(w, 422, err.Error())
			return
		}
		kit.JSON(w, 200, out)
	})
}
