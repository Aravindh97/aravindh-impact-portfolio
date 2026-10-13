package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Aravindh97/aravindh-impact-portfolio/devresume/internal/resume"
	"github.com/Aravindh97/aravindh-impact-portfolio/internal/pdf"
)

func sample() *resume.Resume {
	return &resume.Resume{
		Name: "Ada Lovelace", Title: "Engineer", Email: "ada@example.com",
		Summary: strings.Repeat("Builds reliable systems. ", 30),
		Links:   []resume.Link{{Label: "GitHub", URL: "https://github.com/ada"}},
		Projects: []resume.Project{
			{Slug: "live", Name: "LiveProj", Domain: "Climate", Status: resume.Live, Repo: "https://github.com/ada/live", Color: "#0f766e", Lang: "go", Featured: true},
			{Slug: "plan", Name: "PlannedProj", Domain: "Education", Status: resume.Planned, Featured: true},
		},
	}
}

func TestPDFPaginatesAndLinks(t *testing.T) {
	r := sample()
	for i := 0; i < 40; i++ {
		r.Experience = append(r.Experience, resume.Job{Role: "Engineer", Company: "Co", Start: "2020", End: "2021",
			Bullets: []string{strings.Repeat("Shipped things that mattered. ", 8)}})
	}
	r.SiteURL = "https://ada.github.io/portfolio/"
	b := PDF(r, Options{IncludePlanned: true})
	if !bytes.Contains(b, []byte("/URI (https://ada.github.io/portfolio/systems/live)")) {
		t.Error("missing live app link")
	}
	if !bytes.HasPrefix(b, []byte("%PDF-")) {
		t.Fatal("not a PDF")
	}
	if bytes.Contains(b, []byte("/Count 1 ")) {
		t.Error("expected multiple pages for long content")
	}
	if !bytes.Contains(b, []byte("/URI (https://github.com/ada/live)")) {
		t.Error("missing project link annotation")
	}
	if !bytes.Contains(b, []byte("PlannedProj")) {
		t.Error("featured project should be included")
	}
}

func TestWrapRespectsWidth(t *testing.T) {
	const maxW = 120
	for _, ln := range wrap(strings.Repeat("word ", 50), pdf.Regular, 10, maxW) {
		if w := pdf.TextWidth(pdf.Regular, 10, ln); w > maxW {
			t.Errorf("line %q width %v exceeds %v", ln, w, maxW)
		}
	}
	if wrap("   ", pdf.Regular, 10, maxW) != nil {
		t.Error("blank input should produce no lines")
	}
}
