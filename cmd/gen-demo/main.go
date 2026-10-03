// Command gen-demo writes the home page hero animation data. It runs the
// pipeline on one shape of a pattern (the Classic Jack mouth by default) and
// resamples the drawing, template, bushing-centre and cut outlines to the
// same point count, starting at the same spot, so the browser can tween
// between them.
//
//	go run ./cmd/gen-demo > web/src/lib/demo.json
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/brennanMKE/Bushwhack/internal/template"
	"github.com/brennanMKE/Bushwhack/internal/units"
)

type pt struct{ X, Y float64 }

func main() {
	src := flag.String("pattern", "patterns/halloween/classic-jack.svg", "pattern SVG")
	id := flag.String("id", "mouth", "element id to animate")
	n := flag.Int("n", 240, "points per outline")
	flag.Parse()

	raw, err := os.ReadFile(*src)
	check(err)
	svg := string(raw)
	root := regexp.MustCompile(`(?s)<svg[^>]*>`).FindString(svg)
	el := regexp.MustCompile(`(?s)<[a-z]+\s[^>]*id="` + regexp.QuoteMeta(*id) + `"[^>]*/>`).FindString(svg)
	if root == "" || el == "" {
		check(fmt.Errorf("no element with id %q in %s", *id, *src))
	}
	o := template.DefaultOptions()
	o.Scale = 1 // patterns are drawn in inches
	res, err := template.Process(context.Background(), []byte(root+el+"</svg>"), o)
	check(err)

	drawing := largest(parse(res.Layers.Drawing))
	tmpl := largest(parse(res.Layers.Template))
	center := largest(parse(res.Layers.Center))
	cut := largest(parse(res.Layers.Cut))

	// Common start: the point nearest the drawing's leftmost point.
	anchor := drawing[0]
	for _, p := range drawing {
		if p.X < anchor.X {
			anchor = p
		}
	}
	out := map[string]any{
		"offsetIn":  res.Offset,
		"bushingIn": o.BushingOD,
		"bitIn":     o.BitDia,
		"caption":   fmt.Sprintf("%s offset for a %s bushing and %s bit.", in(res.Offset), in(o.BushingOD), in(o.BitDia)),
		"drawing":   resample(drawing, anchor, *n),
		"template":  resample(tmpl, anchor, *n),
		"center":    resample(center, anchor, *n),
		"cut":       resample(cut, anchor, *n),
	}
	// View box: template bounds plus room for the bushing.
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range tmpl {
		minX, minY = math.Min(minX, p.X), math.Min(minY, p.Y)
		maxX, maxY = math.Max(maxX, p.X), math.Max(maxY, p.Y)
	}
	m := 0.6
	out["viewBox"] = []float64{r3(minX - m), r3(minY - m), r3(maxX - minX + 2*m), r3(maxY - minY + 2*m)}
	enc := json.NewEncoder(os.Stdout)
	check(enc.Encode(out))
}

func in(v float64) string { return strings.TrimSuffix(units.Fraction(v), `"`) + " in" }

func r3(v float64) float64 { return math.Round(v*1000) / 1000 }

func parse(d string) [][]pt {
	var rings [][]pt
	for _, sub := range strings.Split(d, "Z") {
		var ring []pt
		for _, c := range strings.FieldsFunc(sub, func(r rune) bool { return r == 'M' || r == 'L' }) {
			var p pt
			if _, err := fmt.Sscan(c, &p.X, &p.Y); err == nil {
				ring = append(ring, p)
			}
		}
		if len(ring) > 2 {
			rings = append(rings, ring)
		}
	}
	return rings
}

func area(r []pt) float64 {
	a := 0.0
	for i := range r {
		j := (i + 1) % len(r)
		a += r[i].X*r[j].Y - r[j].X*r[i].Y
	}
	return a / 2
}

func largest(rs [][]pt) []pt {
	if len(rs) == 0 {
		check(fmt.Errorf("empty outline"))
	}
	best := rs[0]
	for _, r := range rs {
		if math.Abs(area(r)) > math.Abs(area(best)) {
			best = r
		}
	}
	return best
}

// resample returns n points evenly spaced by arc length, clockwise on screen
// (positive area in y-down space), starting nearest anchor.
func resample(r []pt, anchor pt, n int) [][2]float64 {
	if area(r) < 0 {
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
	}
	start := 0
	for i, p := range r {
		if math.Hypot(p.X-anchor.X, p.Y-anchor.Y) < math.Hypot(r[start].X-anchor.X, r[start].Y-anchor.Y) {
			start = i
		}
	}
	r = append(append([]pt{}, r[start:]...), r[:start]...)
	r = append(r, r[0])
	total := 0.0
	for i := 1; i < len(r); i++ {
		total += math.Hypot(r[i].X-r[i-1].X, r[i].Y-r[i-1].Y)
	}
	out := make([][2]float64, 0, n)
	step := total / float64(n)
	seg, acc := 1, 0.0
	for k := 0; k < n; k++ {
		target := float64(k) * step
		for seg < len(r)-1 && acc+math.Hypot(r[seg].X-r[seg-1].X, r[seg].Y-r[seg-1].Y) < target {
			acc += math.Hypot(r[seg].X-r[seg-1].X, r[seg].Y-r[seg-1].Y)
			seg++
		}
		a, b := r[seg-1], r[seg]
		l := math.Hypot(b.X-a.X, b.Y-a.Y)
		t := 0.0
		if l > 0 {
			t = (target - acc) / l
		}
		out = append(out, [2]float64{r3(a.X + t*(b.X-a.X)), r3(a.Y + t*(b.Y-a.Y))})
	}
	return out
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen-demo:", err)
		os.Exit(1)
	}
}
