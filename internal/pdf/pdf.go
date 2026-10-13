// Package pdf is a minimal, dependency-free PDF 1.4 writer that supports text in the
// standard Helvetica fonts (WinAnsi encoding), lines, filled rectangles and URI links.
//
// All coordinates are in points, measured from the TOP-LEFT of the page.
// Text positions refer to the text baseline.
package pdf

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// A4 page size in points.
const (
	PageWidth  = 595.28
	PageHeight = 841.89
)

type Font int

const (
	Regular Font = iota
	Bold
)

type Color struct{ R, G, B float64 }

// RGB builds a Color from 0-255 components.
func RGB(r, g, b uint8) Color {
	return Color{float64(r) / 255, float64(g) / 255, float64(b) / 255}
}

func (c Color) op() string { return num(c.R) + " " + num(c.G) + " " + num(c.B) }

type link struct {
	x, y, w, h float64 // PDF coordinates (bottom-left origin)
	uri        string
}

type page struct {
	content bytes.Buffer
	links   []link
}

type Document struct {
	Title     string
	Author    string
	CreatedAt time.Time // optional; omitted when zero for reproducible output
	pages     []*page
}

func New(title, author string) *Document { return &Document{Title: title, Author: author} }

func (d *Document) AddPage()       { d.pages = append(d.pages, &page{}) }
func (d *Document) PageCount() int { return len(d.pages) }

func (d *Document) current() *page {
	if len(d.pages) == 0 {
		d.AddPage()
	}
	return d.pages[len(d.pages)-1]
}

// Text draws s with its baseline at (x, baseline).
func (d *Document) Text(x, baseline float64, f Font, size float64, c Color, s string) {
	if s == "" {
		return
	}
	fmt.Fprintf(&d.current().content, "BT %s rg /F%d %s Tf %s %s Td (%s) Tj ET\n",
		c.op(), int(f)+1, num(size), num(x), num(PageHeight-baseline), escape(Encode(s)))
}

// Line draws a straight line.
func (d *Document) Line(x1, y1, x2, y2, width float64, c Color) {
	fmt.Fprintf(&d.current().content, "q %s RG %s w %s %s m %s %s l S Q\n",
		c.op(), num(width), num(x1), num(PageHeight-y1), num(x2), num(PageHeight-y2))
}

// FillRect fills a rectangle whose top-left corner is (x, top).
func (d *Document) FillRect(x, top, w, h float64, c Color) {
	fmt.Fprintf(&d.current().content, "q %s rg %s %s %s %s re f Q\n",
		c.op(), num(x), num(PageHeight-top-h), num(w), num(h))
}

// Link makes the rectangle with top-left corner (x, top) a clickable URI link.
func (d *Document) Link(x, top, w, h float64, uri string) {
	p := d.current()
	p.links = append(p.links, link{x: x, y: PageHeight - top - h, w: w, h: h, uri: uri})
}

// TextWidth returns the rendered width of s in points.
func TextWidth(f Font, size float64, s string) float64 {
	var total float64
	for _, b := range Encode(s) {
		total += float64(widths[f][b])
	}
	return total * size / 1000
}

// Encode converts a UTF-8 string to WinAnsi bytes. Unsupported runes become '?'.
func Encode(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r == '\t':
			out = append(out, ' ')
		case r < 32:
			// drop control characters
		case r < 128:
			out = append(out, byte(r))
		default:
			if b, ok := winAnsi[r]; ok {
				out = append(out, b)
			} else if r >= 160 && r <= 255 {
				out = append(out, byte(r))
			} else {
				out = append(out, '?')
			}
		}
	}
	return out
}

var winAnsi = map[rune]byte{
	'€': 0x80, '‚': 0x82, '„': 0x84, '…': 0x85, '‘': 0x91, '’': 0x92,
	'“': 0x93, '”': 0x94, '•': 0x95, '–': 0x96, '—': 0x97, '™': 0x99,
}

// escape produces an ASCII-only PDF literal string body.
func escape(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		switch {
		case c == '(' || c == ')' || c == '\\':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		case c < 32 || c > 126:
			fmt.Fprintf(&sb, "\\%03o", c)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func num(v float64) string {
	s := strconv.FormatFloat(v, 'f', 3, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}

// WriteTo writes the complete PDF to w.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(d.Bytes())
	return int64(n), err
}

// Bytes serializes the document.
func (d *Document) Bytes() []byte {
	if len(d.pages) == 0 {
		d.AddPage()
	}
	const (
		catalogID = 1
		pagesID   = 2
		font1ID   = 3
		font2ID   = 4
		infoID    = 5
	)
	type pageIDs struct {
		page, content int
		annots        []int
	}
	ids := make([]pageIDs, len(d.pages))
	next := 6
	for i, p := range d.pages {
		ids[i].page, ids[i].content = next, next+1
		next += 2
		for range p.links {
			ids[i].annots = append(ids[i].annots, next)
			next++
		}
	}
	total := next - 1

	var buf bytes.Buffer
	offsets := make([]int, total+1)
	obj := func(id int, body string) {
		offsets[id] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, body)
	}

	buf.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	obj(catalogID, "<< /Type /Catalog /Pages 2 0 R >>")

	kids := make([]string, len(ids))
	for i, p := range ids {
		kids[i] = fmt.Sprintf("%d 0 R", p.page)
	}
	obj(pagesID, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(ids)))
	obj(font1ID, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	obj(font2ID, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")

	info := fmt.Sprintf("<< /Title (%s) /Author (%s) /Producer (DevResume)",
		escape(Encode(d.Title)), escape(Encode(d.Author)))
	if !d.CreatedAt.IsZero() {
		info += fmt.Sprintf(" /CreationDate (D:%sZ)", d.CreatedAt.UTC().Format("20060102150405"))
	}
	obj(infoID, info+" >>")

	for i, p := range d.pages {
		annots := ""
		if len(ids[i].annots) > 0 {
			refs := make([]string, len(ids[i].annots))
			for j, a := range ids[i].annots {
				refs[j] = fmt.Sprintf("%d 0 R", a)
			}
			annots = " /Annots [" + strings.Join(refs, " ") + "]"
		}
		obj(ids[i].page, fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %s %s] /Resources << /Font << /F1 3 0 R /F2 4 0 R >> >> /Contents %d 0 R%s >>",
			num(PageWidth), num(PageHeight), ids[i].content, annots))

		content := p.content.Bytes()
		offsets[ids[i].content] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n<< /Length %d >>\nstream\n", ids[i].content, len(content))
		buf.Write(content)
		buf.WriteString("\nendstream\nendobj\n")

		for j, l := range p.links {
			obj(ids[i].annots[j], fmt.Sprintf(
				"<< /Type /Annot /Subtype /Link /Rect [%s %s %s %s] /Border [0 0 0] /A << /S /URI /URI (%s) >> >>",
				num(l.x), num(l.y), num(l.x+l.w), num(l.y+l.h), escape(Encode(l.uri))))
		}
	}

	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", total+1)
	for id := 1; id <= total; id++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[id])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R /Info 5 0 R >>\nstartxref\n%d\n%%%%EOF\n", total+1, xref)
	return buf.Bytes()
}
