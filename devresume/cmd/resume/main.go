// Command resume generates the PDF resume from the unified JSON data model.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Aravindh97/aravindh-impact-portfolio/devresume/internal/render"
	"github.com/Aravindh97/aravindh-impact-portfolio/devresume/internal/resume"
)

func main() {
	data := flag.String("data", "devresume/resume.json", "resume JSON")
	out := flag.String("out", "apps/web/public/resume.pdf", "output PDF")
	flag.Parse()
	r, err := resume.Load(*data)
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fail(err)
	}
	if err := os.WriteFile(*out, render.PDF(r, render.Options{IncludePlanned: false}), 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("wrote %s\n", *out)
}

func fail(err error) { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
