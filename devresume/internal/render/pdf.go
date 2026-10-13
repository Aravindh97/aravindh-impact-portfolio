// Package render lays out a resume as a PDF document.
package render

import (
	"strings"
	"time"

	"github.com/Aravindh97/aravindh-impact-portfolio/devresume/internal/resume"
	"github.com/Aravindh97/aravindh-impact-portfolio/internal/pdf"
)

var (
	ink    = pdf.RGB(31, 36, 44)
	muted  = pdf.RGB(90, 99, 112)
	accent = pdf.RGB(11, 94, 125)
	rule   = pdf.RGB(200, 206, 214)
)

const (
	margin       = 48.0
	bottomMargin = 48.0
	contentW     = pdf.PageWidth - 2*margin
	lineFactor   = 1.35
)

type Options struct {
	IncludePlanned bool      // include projects with status "planned"
	CreatedAt      time.Time // optional PDF creation date
}

// PDF renders the resume as a single-column A4 PDF.
func PDF(r *resume.Resume, opts Options) []byte {
	l := &layout{d: pdf.New(r.Name+" - Resume", r.Name), y: margin}
	l.d.CreatedAt = opts.CreatedAt
	l.d.AddPage()

	l.header(r)
	if r.Summary != "" {
		l.section("Summary")
		l.para(r.Summary, pdf.Regular, 10, ink, 0)
	}
	if len(r.Skills) > 0 {
		l.section("Skills")
		for _, g := range r.Skills {
			l.skillRow(g)
		}
	}
	if projects := pdfProjects(r, opts.IncludePlanned); len(projects) > 0 {
		l.section("Selected projects")
		for _, p := range projects {
			l.project(p, strings.TrimSuffix(r.SiteURL, "/"))
		}
	}
	if len(r.Experience) > 0 {
		l.section("Experience")
		for _, j := range r.Experience {
			l.rowLR(join(" — ", j.Role, j.Company), pdf.Bold, 10.5, dates(j.Start, j.End))
			if j.Location != "" {
				l.line(0, pdf.Regular, 9, muted, j.Location)
			}
			l.bullets(j.Bullets, 9.5)
			l.y += 4
		}
	}
	if len(r.Education) > 0 {
		l.section("Education")
		for _, e := range r.Education {
			l.rowLR(join(" — ", e.Degree, e.School), pdf.Bold, 10, dates(e.Start, e.End))
		}
	}
	if len(r.Certifications) > 0 {
		l.section("Certifications")
		l.bullets(r.Certifications, 9.5)
	}
	return l.d.Bytes()
}

type layout struct {
	d *pdf.Document
	y float64 // current top position on the page
}

func (l *layout) ensure(h float64) {
	if l.y+h > pdf.PageHeight-bottomMargin {
		l.d.AddPage()
		l.y = margin
	}
}

func (l *layout) line(indent float64, f pdf.Font, size float64, c pdf.Color, s string) {
	l.ensure(size * lineFactor)
	l.d.Text(margin+indent, l.y+size, f, size, c, s)
	l.y += size * lineFactor
}

func (l *layout) para(s string, f pdf.Font, size float64, c pdf.Color, indent float64) {
	for _, ln := range wrap(s, f, size, contentW-indent) {
		l.line(indent, f, size, c, ln)
	}
}

func (l *layout) header(r *resume.Resume) {
	l.d.Text(margin, l.y+22, pdf.Bold, 22, ink, r.Name)
	l.y += 30
	if r.Title != "" {
		l.line(0, pdf.Regular, 12, accent, r.Title)
	}
	contact := join("  |  ", r.Location, r.Email, r.Phone, shortURL(r.SiteURL))
	l.para(contact, pdf.Regular, 9, muted, 0)
	if len(r.Links) > 0 {
		l.linkRow(r.Links, 9)
	}
}

func (l *layout) section(title string) {
	l.ensure(44) // keep the heading with at least a couple of lines of content
	l.y += 10
	l.d.Text(margin, l.y+11, pdf.Bold, 11, accent, strings.ToUpper(title))
	l.y += 15
	l.d.Line(margin, l.y, margin+contentW, l.y, 0.6, rule)
	l.y += 6
}

// rowLR draws left text (wrapping if needed) with right-aligned muted text on the first line.
func (l *layout) rowLR(left string, f pdf.Font, size float64, right string) {
	rightW := 0.0
	if right != "" {
		rightW = pdf.TextWidth(pdf.Regular, 9, right) + 12
	}
	for i, ln := range wrap(left, f, size, contentW-rightW) {
		l.ensure(size * lineFactor)
		l.d.Text(margin, l.y+size, f, size, ink, ln)
		if i == 0 && right != "" {
			l.d.Text(margin+contentW-pdf.TextWidth(pdf.Regular, 9, right), l.y+size, pdf.Regular, 9, muted, right)
		}
		l.y += size * lineFactor
	}
}

func (l *layout) skillRow(g resume.SkillGroup) {
	const size = 9.5
	label := g.Category + ": "
	cw := pdf.TextWidth(pdf.Bold, size, label)
	for i, ln := range wrap(strings.Join(g.Items, ", "), pdf.Regular, size, contentW-cw) {
		l.ensure(size * lineFactor)
		if i == 0 {
			l.d.Text(margin, l.y+size, pdf.Bold, size, ink, label)
		}
		l.d.Text(margin+cw, l.y+size, pdf.Regular, size, ink, ln)
		l.y += size * lineFactor
	}
}

func (l *layout) bullets(items []string, size float64) {
	for _, item := range items {
		for i, ln := range wrap(item, pdf.Regular, size, contentW-14) {
			l.ensure(size * lineFactor)
			if i == 0 {
				l.d.Text(margin+4, l.y+size, pdf.Regular, size, accent, "•")
			}
			l.d.Text(margin+14, l.y+size, pdf.Regular, size, ink, ln)
			l.y += size * lineFactor
		}
		l.y++
	}
}

func (l *layout) project(p resume.Project, siteURL string) {
	right := ""
	if p.Status != resume.Live {
		right = p.Status.Label()
	}
	l.ensure(40)
	l.rowLR(join(" — ", p.Name, join(" · ", p.Subtitle, p.LangLabel())), pdf.Bold, 10.5, right)
	if len(p.Stack) > 0 {
		l.para("Stack: "+strings.Join(p.Stack, ", "), pdf.Regular, 9, muted, 0)
	}
	l.bullets(p.Highlights, 9.5)
	var links []resume.Link
	if siteURL != "" && p.Lang != "" {
		if p.Slug == "devresume" {
			links = append(links, resume.Link{Label: "Live editor", URL: siteURL + "/live/devresume/"})
		} else {
			links = append(links, resume.Link{Label: "Live app", URL: siteURL + "/systems/" + p.Slug},
				resume.Link{Label: "Source code", URL: siteURL + "/live/" + p.Slug + "/"})
		}
	}
	if p.Repo != "" {
		links = append(links, resume.Link{Label: shortURL(p.Repo), URL: p.Repo})
	}
	if p.Demo != "" {
		links = append(links, resume.Link{Label: shortURL(p.Demo), URL: p.Demo})
	}
	if len(links) > 0 {
		l.linkRow(links, 9)
	}
	l.y += 4
}

// linkRow draws clickable link labels separated by "|", wrapping as needed.
func (l *layout) linkRow(links []resume.Link, size float64) {
	const sep = "  |  "
	lh := size * 1.45
	l.ensure(lh)
	x := margin
	sepW := pdf.TextWidth(pdf.Regular, size, sep)
	for i, lk := range links {
		w := pdf.TextWidth(pdf.Regular, size, lk.Label)
		if i > 0 {
			if x+sepW+w > margin+contentW {
				l.y += lh
				l.ensure(lh)
				x = margin
			} else {
				l.d.Text(x, l.y+size, pdf.Regular, size, muted, sep)
				x += sepW
			}
		}
		l.d.Text(x, l.y+size, pdf.Regular, size, accent, lk.Label)
		l.d.Link(x, l.y, w, size*1.25, lk.URL)
		x += w
	}
	l.y += lh
}

func wrap(s string, f pdf.Font, size, maxW float64) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if cand := cur + " " + w; pdf.TextWidth(f, size, cand) <= maxW {
			cur = cand
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	return append(lines, cur)
}

func join(sep string, parts ...string) string {
	var kept []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, sep)
}

func dates(start, end string) string { return join(" – ", start, end) }

func shortURL(u string) string {
	u = strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
	return strings.TrimSuffix(u, "/")
}

// pdfProjects keeps the PDF short: featured projects when any are marked, otherwise all visible ones.
func pdfProjects(r *resume.Resume, includePlanned bool) []resume.Project {
	all := r.VisibleProjects(includePlanned)
	var featured []resume.Project
	for _, p := range all {
		if p.Featured {
			featured = append(featured, p)
		}
	}
	if len(featured) > 0 {
		return featured
	}
	return all
}
