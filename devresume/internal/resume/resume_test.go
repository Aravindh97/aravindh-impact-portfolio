package resume

import (
	"os"
	"strings"
	"testing"
)

const valid = `{"name":"Ada","email":"ada@example.com","links":[{"label":"GH","url":"https://github.com/ada"}],
"projects":[{"slug":"a","name":"A","status":"live"},{"slug":"b","name":"B","status":"planned"}]}`

func TestParseValid(t *testing.T) {
	r, err := Parse([]byte(valid))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(r.VisibleProjects(false)); got != 1 {
		t.Errorf("visible without planned = %d, want 1", got)
	}
	if got := len(r.VisibleProjects(true)); got != 2 {
		t.Errorf("visible with planned = %d, want 2", got)
	}
}

func TestValidationErrors(t *testing.T) {
	cases := map[string]string{
		"missing name":      `{"email":"a@b.co"}`,
		"bad email":         `{"name":"A","email":"nope"}`,
		"javascript url":    `{"name":"A","email":"a@b.co","links":[{"label":"x","url":"javascript:alert(1)"}]}`,
		"relative url":      `{"name":"A","email":"a@b.co","links":[{"label":"x","url":"/x"}]}`,
		"bad slug":          `{"name":"A","email":"a@b.co","projects":[{"slug":"Bad Slug","name":"A","status":"live"}]}`,
		"duplicate slug":    `{"name":"A","email":"a@b.co","projects":[{"slug":"a","name":"A","status":"live"},{"slug":"a","name":"B","status":"live"}]}`,
		"bad status":        `{"name":"A","email":"a@b.co","projects":[{"slug":"a","name":"A","status":"done"}]}`,
		"bad demo url":      `{"name":"A","email":"a@b.co","projects":[{"slug":"a","name":"A","status":"live","demo":"ftp://x"}]}`,
		"unknown field":     `{"name":"A","email":"a@b.co","nmae":"typo"}`,
		"malformed json":    `{"name":`,
		"bad color":         `{"name":"A","email":"a@b.co","projects":[{"slug":"a","name":"A","status":"live","color":"red;x"}]}`,
		"script icon":       `{"name":"A","email":"a@b.co","projects":[{"slug":"a","name":"A","status":"live","icon":"<script>alert(1)</script>"}]}`,
		"event attr icon":   `{"name":"A","email":"a@b.co","projects":[{"slug":"a","name":"A","status":"live","icon":"<path onload=\\"x\\"/>"}]}`,
		"bad lang":          `{"name":"A","email":"a@b.co","projects":[{"slug":"a","name":"A","status":"live","lang":"rust"}]}`,
		"missing proj name": `{"name":"A","email":"a@b.co","projects":[{"slug":"a","status":"live"}]}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(in)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestShippedResumeIsValid(t *testing.T) {
	if _, err := os.Stat("../../resume.json"); err != nil {
		t.Skip("resume.json not found")
	}
	if _, err := Load("../../resume.json"); err != nil {
		t.Fatalf("resume.json invalid: %v", err)
	}
}

func TestStatusLabel(t *testing.T) {
	if !strings.EqualFold(InProgress.Label(), "in progress") {
		t.Errorf("label = %q", InProgress.Label())
	}
}
