package pdf

import (
	"bytes"
	"regexp"
	"strconv"
	"testing"
)

func TestEncodeAndEscape(t *testing.T) {
	got := escape(Encode("a(b)\\ — •é✓"))
	want := `a\(b\)\\ \227 \225\351?`
	if got != want {
		t.Errorf("escape = %q, want %q", got, want)
	}
}

func TestTextWidth(t *testing.T) {
	// "Hi" in Helvetica: H=722, i=222 -> 944/1000 * 10pt
	if w := TextWidth(Regular, 10, "Hi"); w < 9.43 || w > 9.45 {
		t.Errorf("width = %v, want 9.44", w)
	}
	if TextWidth(Bold, 10, "Hi") <= TextWidth(Regular, 10, "Hi") {
		t.Error("bold should be wider than regular")
	}
}

// TestXrefOffsets verifies that every xref entry points at its "N 0 obj" header.
func TestXrefOffsets(t *testing.T) {
	d := New("T", "A")
	d.Text(10, 20, Regular, 12, RGB(0, 0, 0), "Hello (world)")
	d.Link(10, 10, 50, 12, "https://example.com")
	d.AddPage()
	d.Line(0, 0, 100, 100, 1, RGB(255, 0, 0))
	b := d.Bytes()

	if !bytes.HasPrefix(b, []byte("%PDF-1.4")) || !bytes.HasSuffix(b, []byte("%%EOF\n")) {
		t.Fatal("missing header or trailer")
	}
	m := regexp.MustCompile(`startxref\n(\d+)\n`).FindSubmatch(b)
	if m == nil {
		t.Fatal("no startxref")
	}
	xref, _ := strconv.Atoi(string(m[1]))
	if !bytes.HasPrefix(b[xref:], []byte("xref\n")) {
		t.Fatal("startxref does not point at xref table")
	}
	entries := regexp.MustCompile(`(\d{10}) 00000 n `).FindAllSubmatch(b[xref:], -1)
	// 5 fixed objects + page1 (page, content, annot) + page2 (page, content) = 10
	if len(entries) != 10 {
		t.Fatalf("xref entries = %d, want 10", len(entries))
	}
	for i, e := range entries {
		off, _ := strconv.Atoi(string(e[1]))
		want := strconv.Itoa(i+1) + " 0 obj"
		if !bytes.HasPrefix(b[off:], []byte(want)) {
			t.Errorf("object %d offset %d points at %q", i+1, off, b[off:off+12])
		}
	}
	if !bytes.Contains(b, []byte("/Count 2")) {
		t.Error("expected 2 pages")
	}
}
