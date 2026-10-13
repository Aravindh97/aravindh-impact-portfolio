// Package resume defines the resume data model and loads/validates resume.json.
// A single resume.json is the source of truth for both the website and the PDF.
package resume

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Status describes how far along a portfolio project is.
type Status string

const (
	Planned    Status = "planned"
	InProgress Status = "in-progress"
	Live       Status = "live"
)

// Label returns a human-readable status label.
func (s Status) Label() string {
	switch s {
	case Planned:
		return "Planned"
	case InProgress:
		return "In progress"
	case Live:
		return "Live"
	default:
		return string(s)
	}
}

func (s Status) valid() bool { return s == Planned || s == InProgress || s == Live }

type Resume struct {
	Name           string       `json:"name"`
	Title          string       `json:"title"`
	Location       string       `json:"location"`
	Email          string       `json:"email"`
	Phone          string       `json:"phone,omitempty"`
	Links          []Link       `json:"links"`
	Summary        string       `json:"summary"`
	Skills         []SkillGroup `json:"skills"`
	Experience     []Job        `json:"experience"`
	Projects       []Project    `json:"projects"`
	Education      []Education  `json:"education"`
	Certifications []string     `json:"certifications,omitempty"`
	// SiteURL is where the built site is hosted (e.g. https://<user>.github.io/<repo>/).
	// When set, the PDF links every project to its live app and source code.
	SiteURL string `json:"site_url,omitempty"`
}

type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type SkillGroup struct {
	Category string   `json:"category"`
	Items    []string `json:"items"`
}

type Job struct {
	Role     string   `json:"role"`
	Company  string   `json:"company"`
	Location string   `json:"location,omitempty"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
	Bullets  []string `json:"bullets"`
}

type Project struct {
	Slug                 string      `json:"slug"`
	Name                 string      `json:"name"`
	Lang                 string      `json:"lang,omitempty"` // "python" or "go": runnable app at apps/<slug>/, source at live/<slug>/
	Domain               string      `json:"domain"`
	Subtitle             string      `json:"subtitle,omitempty"`
	Tagline              string      `json:"tagline"`
	Problem              string      `json:"problem,omitempty"`
	Solution             string      `json:"solution,omitempty"`
	Stack                []string    `json:"stack"`
	Status               Status      `json:"status"`
	Featured             bool        `json:"featured,omitempty"` // included in the PDF
	Repo                 string      `json:"repo,omitempty"`
	Demo                 string      `json:"demo,omitempty"`
	Highlights           []string    `json:"highlights,omitempty"`
	Features             []string    `json:"features,omitempty"`
	DataSources          []string    `json:"data_sources,omitempty"`
	EngineeringDecisions []string    `json:"engineering_decisions,omitempty"`
	ProductionPath       string      `json:"production_path,omitempty"`
	Color                string      `json:"color,omitempty"` // #rrggbb
	Icon                 string      `json:"icon,omitempty"`  // inner SVG markup for a 24x24 stroke icon
	Playground           *Playground `json:"playground,omitempty"`
}

// Playground configures a Python project's in-browser runtime.
type Playground struct {
	Packages []string `json:"packages"`
	Open     string   `json:"open"`
}

// LangLabel is the display name of the project's language.
func (p Project) LangLabel() string {
	switch p.Lang {
	case "python":
		return "Python"
	case "go":
		return "Go"
	case "go+python":
		return "Python + Go"
	}
	return ""
}

type Education struct {
	Degree string `json:"degree"`
	School string `json:"school"`
	Start  string `json:"start,omitempty"`
	End    string `json:"end,omitempty"`
}

// Load reads and validates a resume JSON file.
func Load(path string) (*Resume, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read resume: %w", err)
	}
	return Parse(data)
}

// Parse decodes and validates resume JSON. Unknown fields are rejected so typos surface early.
func Parse(data []byte) (*Resume, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var r Resume
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("decode resume: %w", err)
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

var (
	slugRE   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	accentRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	// Icons are inlined as HTML, so only plain SVG shapes with geometry attributes are allowed.
	iconRE = regexp.MustCompile(`^(\s*<(path|circle|rect|line|polyline|polygon|ellipse)(\s+(d|cx|cy|r|rx|ry|x|y|x1|x2|y1|y2|width|height|points)="[-0-9a-zA-Z .,]*")*\s*/>)*\s*$`)
)

// Validate checks required fields, URL safety (http/https only) and slug uniqueness.
func (r *Resume) Validate() error {
	var errs []error
	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		errs = append(errs, fmt.Errorf("email %q is invalid", r.Email))
	}
	for i, l := range r.Links {
		if strings.TrimSpace(l.Label) == "" {
			errs = append(errs, fmt.Errorf("links[%d]: label is required", i))
		}
		if !validURL(l.URL) {
			errs = append(errs, fmt.Errorf("links[%d]: url %q must be an absolute http(s) URL", i, l.URL))
		}
	}
	if r.SiteURL != "" && !validURL(r.SiteURL) {
		errs = append(errs, fmt.Errorf("site_url %q must be an absolute http(s) URL", r.SiteURL))
	}
	seen := map[string]bool{}
	for i, p := range r.Projects {
		if !slugRE.MatchString(p.Slug) {
			errs = append(errs, fmt.Errorf("projects[%d]: slug %q must be lowercase-kebab-case", i, p.Slug))
		} else if seen[p.Slug] {
			errs = append(errs, fmt.Errorf("projects[%d]: duplicate slug %q", i, p.Slug))
		}
		seen[p.Slug] = true
		if strings.TrimSpace(p.Name) == "" {
			errs = append(errs, fmt.Errorf("projects[%d]: name is required", i))
		}
		if !p.Status.valid() {
			errs = append(errs, fmt.Errorf("projects[%d]: status %q must be planned, in-progress or live", i, p.Status))
		}
		if p.Color != "" && !accentRE.MatchString(p.Color) {
			errs = append(errs, fmt.Errorf("projects[%d]: color %q must be #rrggbb", i, p.Color))
		}
		if !iconRE.MatchString(p.Icon) {
			errs = append(errs, fmt.Errorf("projects[%d]: icon may only contain simple SVG shapes", i))
		}
		if p.Lang != "" && p.Lang != "python" && p.Lang != "go" && p.Lang != "go+python" {
			errs = append(errs, fmt.Errorf("projects[%d]: lang must be python, go or go+python", i))
		}
		for field, u := range map[string]string{"repo": p.Repo, "demo": p.Demo} {
			if u != "" && !validURL(u) {
				errs = append(errs, fmt.Errorf("projects[%d]: %s %q must be an absolute http(s) URL", i, field, u))
			}
		}
	}
	return errors.Join(errs...)
}

// VisibleProjects returns projects to show; planned ones are excluded unless includePlanned is set.
func (r *Resume) VisibleProjects(includePlanned bool) []Project {
	var out []Project
	for _, p := range r.Projects {
		if includePlanned || p.Status != Planned {
			out = append(out, p)
		}
	}
	return out
}

func validURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
