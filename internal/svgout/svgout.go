// Package svgout writes the true-size template SVG and the colour-coded
// reference (preview) SVG. Output is generated entirely from geometry; no
// input markup is ever echoed.
package svgout

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/brennanMKE/Bushwhack/internal/geom"
	"github.com/brennanMKE/Bushwhack/internal/units"
)

// Meta describes the job for titles and the legend.
type Meta struct {
	BushingOD, BitDia float64
	Mode              string
	Offset            float64
	PageW, PageH      float64
}

// Mark is a thin-bridge marker between two closest points.
type Mark struct{ P, Q geom.Pt }

// Layers are the preview's geometry, in page inches. Each piece becomes its
// own top-level <path>, so every shape can be selected on its own; ids must
// be unique across all three layers.
type Layers struct {
	Original []Opening // the drawing as submitted, one per shape
	Template []Opening // offset openings
	Cut      []Opening // final edge in the workpiece, one per cut-out
	Bridges  []Mark
	BushingR float64 // radius of the bridge marker circle
}

// Print colours: chalk blue, black and cut red on white. Red is reserved for
// the cut, so bridges are marked in brass.
const (
	ColorOriginal = "#2E6FD8"
	ColorTemplate = "#000000"
	ColorCut      = "#C93A2E"
	ColorBridge   = "#C4943A"
)

func num(v float64) string {
	s := strconv.FormatFloat(v, 'f', 4, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")
	if s == "-0" {
		return "0"
	}
	return s
}

// PathData renders polygons as one path "d" string.
func PathData(p geom.Polys) string {
	var b strings.Builder
	for _, ring := range geom.ToRings(p) {
		if len(ring) < 3 {
			continue
		}
		for i, pt := range ring {
			if i == 0 {
				b.WriteByte('M')
			} else {
				b.WriteByte('L')
			}
			b.WriteString(num(pt.X))
			b.WriteByte(' ')
			b.WriteString(num(pt.Y))
		}
		b.WriteByte('Z')
	}
	return b.String()
}

func describe(m Meta) string {
	what := "the opening (hole) matches the drawing"
	if m.Mode == "piece" {
		what = "the piece that falls out (inlay) matches the drawing"
	}
	return fmt.Sprintf("Bushing %s OD, bit %s, %s mode: %s. Template offset %s (%.4f in).",
		units.Fraction(m.BushingOD), units.Fraction(m.BitDia), m.Mode, what,
		units.Fraction(m.Offset), m.Offset)
}

// Opening is one cut-out in the template: an outer ring plus any holes
// inside it, written as its own <path> so laser and CNC software can select
// it and assign it to a layer on its own.
type Opening struct {
	ID    string // unique XML id
	Polys geom.Polys
}

// Template writes the cut-ready template: black hairline outlines at true
// size with physical units declared, one top-level <path> per opening (no
// wrapping group, so nothing needs ungrouping on import).
func Template(m Meta, openings []Opening) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%sin" height="%sin" viewBox="0 0 %s %s">
<title>Bushwhack router template</title>
<desc>%s Import at 100%% and do not rescale.</desc>
`, num(m.PageW), num(m.PageH), num(m.PageW), num(m.PageH), escape(describe(m)))
	for _, o := range openings {
		fmt.Fprintf(&b, `<path id="%s" d="%s" fill="none" stroke="#000000" stroke-width="0.01" fill-rule="evenodd"/>
`, escape(o.ID), PathData(o.Polys))
	}
	b.WriteString("</svg>\n")
	return []byte(b.String())
}

// Preview writes the printable reference: the original drawing dashed in
// chalk blue, the template in black, the simulated cut in red and thin
// bridges as brass bushing circles. A legend sits below the true-size page.
func Preview(m Meta, l Layers) []byte {
	const legendH = 0.55
	w, h := m.PageW, m.PageH
	totalH := h + legendH
	stroke := func(color string, width float64, extra string) string {
		return fmt.Sprintf(`fill="none" stroke="%s" stroke-width="%s" vector-effect="non-scaling-stroke" stroke-linejoin="round"%s`, color, num(width), extra)
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%sin" height="%sin" viewBox="0 0 %s %s">
<title>Bushwhack preview</title>
<desc>%s Dashed blue: your drawing. Black: the template. Red: where the router cuts. Brass circles: thin bridges.</desc>
<rect id="paper" x="0" y="0" width="%s" height="%s" fill="#ffffff"/>
<rect id="page-edge" x="0" y="0" width="%s" height="%s" fill="none" stroke="#c9ced6" stroke-width="1" vector-effect="non-scaling-stroke"/>
`, num(w), num(totalH), num(w), num(totalH), escape(describe(m)), num(w), num(totalH), num(w), num(h))
	paths := func(pieces []Opening, attrs string) {
		for _, o := range pieces {
			fmt.Fprintf(&b, `<path id="%s" d="%s" fill-rule="evenodd" %s/>
`, escape(o.ID), PathData(o.Polys), attrs)
		}
	}
	paths(l.Template, stroke(ColorTemplate, 1.2, ""))
	paths(l.Cut, stroke(ColorCut, 1.6, ""))
	paths(l.Original, stroke(ColorOriginal, 1.2, ` stroke-dasharray="5 3"`))
	for i, mk := range l.Bridges {
		cx, cy := (mk.P.X+mk.Q.X)/2, (mk.P.Y+mk.Q.Y)/2
		fmt.Fprintf(&b, `<circle id="thin-bridge-%d" cx="%s" cy="%s" r="%s" %s/>
`, i+1, num(cx), num(cy), num(l.BushingR), stroke(ColorBridge, 2, ""))
	}

	fs := 0.13
	if w < 5.5 {
		fs = 0.13 * w / 5.5
	}
	y := h + legendH*0.55
	items := []struct{ color, label, dash string }{
		{ColorOriginal, "your drawing", "5 3"},
		{ColorTemplate, "template", ""},
		{ColorCut, "router cut", ""},
		{ColorBridge, "thin bridge", ""},
	}
	x := 0.1 * fs / 0.13
	fmt.Fprintf(&b, `<g font-family="Helvetica, Arial, sans-serif" font-size="%s" fill="#23262A">`, num(fs))
	for _, it := range items {
		dash := ""
		if it.dash != "" {
			dash = fmt.Sprintf(` stroke-dasharray="%s"`, it.dash)
		}
		fmt.Fprintf(&b, `<line x1="%s" y1="%s" x2="%s" y2="%s" %s/><text x="%s" y="%s">%s</text>`,
			num(x), num(y-fs*0.35), num(x+fs*2.2), num(y-fs*0.35), stroke(it.color, 2, dash),
			num(x+fs*2.6), num(y), it.label)
		x += fs * (2.6 + 0.55*float64(len(it.label)) + 1.4)
	}
	// The description is ~150 characters; size it to fit the page width.
	desc := describe(m) + " Import at 100%."
	dfs := math.Min(fs*0.85, (w-0.2)/(0.5*float64(len(desc))))
	fmt.Fprintf(&b, `<text x="%s" y="%s" font-size="%s" fill="#555555">%s</text></g>
`, num(0.1*fs/0.13), num(y+fs*1.5), num(dfs), escape(desc))
	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func escape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
