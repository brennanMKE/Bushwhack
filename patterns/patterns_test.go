package patterns

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/brennanMKE/Bushwhack/internal/geom"
	"github.com/brennanMKE/Bushwhack/internal/template"
)

func TestUpcoming(t *testing.T) {
	cases := map[string]string{
		"2026-10-02": "halloween", "2026-09-01": "halloween", "2026-10-31": "halloween",
		"2026-11-01": "winter", "2026-12-31": "winter", "2027-01-01": "valentines",
		"2027-02-14": "valentines", "2027-02-15": "spring", "2027-05-01": "july4",
		"2027-07-04": "july4", "2027-07-05": "halloween", "2027-08-31": "halloween",
	}
	for day, want := range cases {
		d, _ := time.Parse("2006-01-02", day)
		if got := Upcoming(d).Slug; got != want {
			t.Errorf("%s: upcoming %s, want %s", day, got, want)
		}
	}
	d, _ := time.Parse("2006-01-02", "2026-10-02")
	ord := Ordered(d)
	if ord[0].Slug != "halloween" || ord[1].Slug != "winter" || ord[len(ord)-1].Slug != "celebrations" {
		t.Errorf("order %v", ord)
	}
}

var allowedTransform = regexp.MustCompile(`^\s*((translate|scale)\([^)]*\)\s*)*$`)

// TestPatternRules enforces the design rules on every pattern.
func TestPatternRules(t *testing.T) {
	lib, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.All()) == 0 {
		t.Fatal("no patterns")
	}
	for _, p := range lib.All() {
		t.Run(p.Slug, func(t *testing.T) {
			checkMarkup(t, p.SVG)
			o := p.Options(template.DefaultOptions())
			res, err := template.Process(context.Background(), p.SVG, o)
			if err != nil {
				t.Fatal(err)
			}
			if msg := Check(res, 0.25); msg != "" {
				t.Errorf("at %.2f in: %s", o.Size, msg)
			}
			for _, n := range res.Notes {
				if n.Kind != template.NoteBridge {
					t.Errorf("warning: %s", n.Text)
				}
			}
			// Aspect close to square.
			if r := math.Max(res.PageWidthIn, res.PageHeightIn) / math.Min(res.PageWidthIn, res.PageHeightIn); r > 1.3 {
				t.Errorf("canvas aspect %.2f; keep it square or close", r)
			}
			// Minimum feature width 3/16 in at the recommended size: opening the
			// drawing with radius 3/32 keeps nearly all of it.
			for _, s := range res.Shapes {
				polys := pathPolys(t, s.Path)
				lost := geom.Area(geom.Difference(polys, geom.Open(polys, 3.0/32-0.002))) / geom.Area(polys)
				if lost > 0.05 {
					t.Errorf("shape %d (%s): %.1f%% is narrower than 3/16 in", s.Index, s.Name, lost*100)
				}
			}
		})
	}
}

// checkMarkup enforces: inch canvas, closed filled shapes only, a 0.25 in
// margin, no strokes/text/images, transforms limited to translate and scale.
func checkMarkup(t *testing.T, svg []byte) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(string(svg)))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		get := func(n string) string {
			for _, a := range se.Attr {
				if a.Name.Local == n {
					return a.Value
				}
			}
			return ""
		}
		switch se.Name.Local {
		case "svg":
			w, h, vb := get("width"), get("height"), get("viewBox")
			if !strings.HasSuffix(w, "in") || !strings.HasSuffix(h, "in") {
				t.Errorf("width/height must be in inches, got %q %q", w, h)
			}
			want := "0 0 " + strings.TrimSuffix(w, "in") + " " + strings.TrimSuffix(h, "in")
			if vb != want {
				t.Errorf("viewBox %q should be %q (inches)", vb, want)
			}
		case "text", "image", "line", "polyline", "use":
			t.Errorf("<%s> is not allowed", se.Name.Local)
		}
		if s := get("stroke"); s != "" && s != "none" {
			t.Errorf("<%s> has a stroke; patterns are filled shapes only", se.Name.Local)
		}
		if tr := get("transform"); tr != "" && !allowedTransform.MatchString(tr) {
			t.Errorf("transform %q: only translate and scale are allowed", tr)
		}
	}
	// Margin: the drawing at native (inch) scale stays 0.25 in inside the canvas.
	o := template.DefaultOptions()
	o.Scale = 1
	res, err := template.Process(context.Background(), svg, o)
	if err != nil {
		t.Fatal(err)
	}
	b := geom.Bounds(pathPolys(t, res.Layers.Drawing))
	if b.MinX < 0.25-1e-6 || b.MinY < 0.25-1e-6 || b.MaxX > res.PageWidthIn-0.25+1e-6 || b.MaxY > res.PageHeightIn-0.25+1e-6 {
		t.Errorf("drawing %+v is closer than 0.25 in to the %.3f x %.3f canvas edge", b, res.PageWidthIn, res.PageHeightIn)
	}
}

// pathPolys parses our own M/L/Z output back into polygons.
func pathPolys(t *testing.T, d string) geom.Polys {
	t.Helper()
	var rings [][]geom.Pt
	for _, sub := range strings.Split(d, "Z") {
		sub = strings.TrimSpace(sub)
		if sub == "" {
			continue
		}
		var ring []geom.Pt
		for _, cmd := range strings.FieldsFunc(sub, func(r rune) bool { return r == 'M' || r == 'L' }) {
			var x, y float64
			if _, err := fmt.Sscan(cmd, &x, &y); err != nil {
				t.Fatalf("bad path %q: %v", cmd, err)
			}
			ring = append(ring, geom.Pt{X: x, Y: y})
		}
		rings = append(rings, ring)
	}
	return geom.FromRings(rings)
}
