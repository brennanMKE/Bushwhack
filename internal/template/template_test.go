package template

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brennanMKE/Bushwhack/internal/geom"
)

func run(t *testing.T, svg string, o Options) *Result {
	t.Helper()
	res, err := Process(context.Background(), []byte(svg), o)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func ring(xy ...float64) []geom.Pt {
	r := make([]geom.Pt, 0, len(xy)/2)
	for i := 0; i+1 < len(xy); i += 2 {
		r = append(r, geom.Pt{X: xy[i], Y: xy[i+1]})
	}
	return r
}

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.5f, want %.5f (±%g)", name, got, want, tol)
	}
}

// A 1" square in a 4" canvas with scale 1 in/unit.
const square = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"><rect x="1.5" y="1.5" width="1" height="1"/></svg>`

func scaled(o Options) Options { o.Scale = 1; return o }

func TestSquareOffsetHole(t *testing.T) {
	res := run(t, square, scaled(DefaultOptions()))
	near(t, "offset", res.Offset, 0.09375, 1e-9)
	s := res.Shapes[0]
	near(t, "template w", s.TemplateWidthIn, 1.1875, 0.001)
	near(t, "template h", s.TemplateHeightIn, 1.1875, 0.001)
	// Simulated cut: 1" square with bitR corner radii.
	lost := 4 * (1 - math.Pi/4) * 0.0625 * 0.0625 * 100
	near(t, "lost %", s.LostAreaPct, lost, 0.25)
	if s.OvercutPct > 0.01 {
		t.Errorf("overcut = %v", s.OvercutPct)
	}
}

func TestSquareCornerRadius(t *testing.T) {
	// The template's area tells us the corner radius: 1 + 4r + πr².
	o := scaled(DefaultOptions())
	res := run(t, square, o)
	if !strings.Contains(string(res.TemplateSVG), `width="4in"`) {
		t.Error("template must declare physical units")
	}
	sq := geom.FromRings([][]geom.Pt{ring(0, 0, 1, 0, 1, 1, 0, 1)})
	r := 0.09375
	near(t, "area", geom.Area(geom.Offset(sq, r)), 1+4*r+math.Pi*r*r, 0.001)
}

func TestPieceMode(t *testing.T) {
	o := scaled(DefaultOptions())
	o.Mode = ModePiece
	res := run(t, square, o)
	near(t, "offset", res.Offset, 0.21875, 1e-9)
	near(t, "template w", res.Shapes[0].TemplateWidthIn, 1.4375, 0.001)
	// Outside corners of a piece stay sharp: the bit's far edge wraps them.
	near(t, "lost %", res.Shapes[0].LostAreaPct, 0, 0.05)
}

func TestEvenOddHoleShrinks(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"><path d="M1 1H3V3H1Z M1.6 1.6V2.4H2.4V1.6Z"/></svg>`
	res := run(t, svg, scaled(DefaultOptions()))
	found := false
	for _, n := range res.Notes {
		if n.Kind == NoteIsland {
			found = true
		}
	}
	if !found {
		t.Errorf("expected island warning, got %v", res.Warnings)
	}
	inner := geom.FromRings([][]geom.Pt{ring(1, 1, 3, 1, 3, 3, 1, 3), ring(1.6, 1.6, 1.6, 2.4, 2.4, 2.4, 2.4, 1.6)})
	if len(inner) != 2 {
		t.Fatalf("even-odd should give outer + hole, got %d rings", len(inner))
	}
	off := geom.Offset(inner, 0.09375)
	// outer grows to 2.1875 with round corners; hole shrinks to 0.8-0.1875 square.
	hole := 0.8 - 2*0.09375
	want := (4 + 4*2*0.09375 + math.Pi*0.09375*0.09375) - hole*hole
	near(t, "area", geom.Area(off), want, 0.002)
}

func TestBridge(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"><rect x="0.5" y="1" width="1" height="1"/><rect x="1.8" y="1" width="1" height="1"/></svg>`
	res := run(t, svg, scaled(DefaultOptions()))
	near(t, "min bridge", res.MinBridgeIn, 0.3-2*0.09375, 0.001)
	if len(res.Warnings) == 0 || !strings.Contains(strings.Join(res.Warnings, "\n"), "is 0.113 in thick") {
		t.Errorf("expected thin-bridge warning, got %v", res.Warnings)
	}
}

func TestMerged(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"><rect x="0.5" y="1" width="1" height="1"/><rect x="1.6" y="1" width="1" height="1"/></svg>`
	res := run(t, svg, scaled(DefaultOptions()))
	if res.MinBridgeIn != 0 || !strings.Contains(strings.Join(res.Warnings, "\n"), "merged into one opening") {
		t.Errorf("expected merge, got bridge %v warnings %v", res.MinBridgeIn, res.Warnings)
	}
}

func TestFitCanvasAndArtwork(t *testing.T) {
	o := DefaultOptions()
	res := run(t, square, o)
	near(t, "scale", res.ScaleInPerUnit, 1.5, 1e-9) // 6 / 4
	near(t, "page w", res.PageWidthIn, 6, 1e-9)
	o.Fit = FitArtwork
	res = run(t, square, o)
	near(t, "scale", res.ScaleInPerUnit, 6, 1e-6)
	near(t, "shape w", res.Shapes[0].WidthIn, 6, 0.001)
}

func TestFitDocument(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="100mm" height="50mm" viewBox="0 0 100 50"><circle cx="50" cy="25" r="12.7"/></svg>`
	o := DefaultOptions()
	o.Fit = FitDocument
	res := run(t, svg, o)
	near(t, "circle dia", res.Shapes[0].WidthIn, 1.0, 0.002)
}

func TestPageGrows(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><rect x="0" y="0" width="1" height="1"/></svg>`
	res := run(t, svg, scaled(DefaultOptions()))
	if res.PageWidthIn <= 1.18 {
		t.Errorf("page should grow, got %v", res.PageWidthIn)
	}
}

func TestBadOptions(t *testing.T) {
	o := DefaultOptions()
	o.BitDia = 0.5
	if _, err := Process(context.Background(), []byte(square), o); err == nil {
		t.Error("bit larger than bushing must fail")
	}
}

func TestBitMustPassThroughBushing(t *testing.T) {
	for _, c := range []struct {
		od, id, bit float64
		ok          bool
	}{
		{5.0 / 16, 0, 1.0 / 8, true},         // standard 1/4 in inside
		{5.0 / 16, 0, 1.0 / 4, false},        // standard inside is exactly 1/4 in
		{5.0 / 16, 17.0 / 64, 1.0 / 4, true}, // Bosch RA1103: 17/64 in inside
		{1.0 / 2, 0, 7.0 / 16, false},        // standard 13/32 in inside
		{0.55, 0, 0.5, true},                 // not a standard size: only the OD is known
		{0.55, 0.45, 0.5, false},
		{1.0 / 2, 0.6, 1.0 / 8, false}, // inside bigger than outside
	} {
		o := DefaultOptions()
		o.BushingOD, o.BushingID, o.BitDia = c.od, c.id, c.bit
		_, err := Process(context.Background(), []byte(square), scaled(o))
		if (err == nil) != c.ok {
			t.Errorf("OD %v ID %v bit %v: err %v, want ok=%v", c.od, c.id, c.bit, err, c.ok)
		}
	}
}

func TestShapeLabels(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"><rect id="nose" x="0.5" y="1" width="1" height="1"/><rect id="path12" x="1.8" y="1" width="1" height="1"/></svg>`
	res := run(t, svg, scaled(DefaultOptions()))
	var bridge *Note
	for i := range res.Notes {
		if res.Notes[i].Kind == NoteBridge {
			bridge = &res.Notes[i]
		}
	}
	if bridge == nil || !strings.Contains(bridge.Text, "between the nose and shape 2") || bridge.Bridge != 0 {
		t.Errorf("bridge note %+v", bridge)
	}
	if res.Shapes[0].Name != "nose" {
		t.Errorf("name %q", res.Shapes[0].Name)
	}
}

func TestRejectsDoctype(t *testing.T) {
	svg := `<?xml version="1.0"?><!DOCTYPE svg [<!ENTITY x "y">]><svg xmlns="http://www.w3.org/2000/svg"/>`
	if _, err := Process(context.Background(), []byte(svg), DefaultOptions()); err == nil {
		t.Error("DOCTYPE must be rejected")
	}
}

func FuzzProcess(f *testing.F) {
	for _, s := range []string{
		square,
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><path d="M1 1h3v3zM5 5a2 2 0 1 0 1 1z"/><g transform="rotate(30)"><circle r="2" cx="5" cy="5"/></g></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg"><polygon points="0,0 1,0 0,1"/></svg>`,
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		done := make(chan struct{})
		go func() {
			defer close(done)
			Process(ctx, data, DefaultOptions())
		}()
		// The server's 10 s budget only holds if Process honours ctx promptly.
		select {
		case <-done:
		case <-time.After(6 * time.Second):
			if dir := os.Getenv("BUSHWHACK_SLOW_DIR"); dir != "" {
				os.WriteFile(filepath.Join(dir, fmt.Sprintf("slow-%d.svg", time.Now().UnixNano())), data, 0o644)
			}
			t.Fatalf("Process still running 6 s after start despite a 2 s deadline")
		}
	})
}

func TestTemplateOnePathPerOpening(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 6 4">
	  <rect id="nose" x="0.5" y="1" width="1" height="1"/>
	  <rect id="path12" x="2.5" y="1" width="1" height="1"/>
	  <path id="Left Eye!" d="M4.5 1H5.5V2H4.5Z"/>
	  <rect id="a" x="0.5" y="2.6" width="1" height="1"/><rect id="b" x="1.6" y="2.6" width="1" height="1"/>
	  <path id="ring" d="M3 2.4H5.6V3.9H3Z M3.6 2.9V3.4H5V2.9Z"/>
	</svg>`
	res := run(t, svg, scaled(DefaultOptions()))
	out := string(res.TemplateSVG)
	for _, want := range []string{`id="nose"`, `id="opening-2"`, `id="left-eye"`, `id="a-and-b"`, `id="ring"`} {
		if !strings.Contains(out, want) {
			t.Errorf("template missing %s:\n%s", want, out)
		}
	}
	if n := strings.Count(out, "<path "); n != 5 {
		t.Errorf("got %d paths, want 5 (one per opening)", n)
	}
	if strings.Contains(out, "<g") {
		t.Error("template should have no groups to ungroup")
	}
	// The ring's hole stays in the ring's own path.
	i := strings.Index(out, `id="ring"`)
	if strings.Count(out[i:strings.Index(out[i:], "/>")+i], "Z") != 2 {
		t.Error("ring opening should carry its hole")
	}
}

func TestPreviewOnePathPerShape(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 6 4">
	  <rect id="nose" x="0.5" y="1" width="1" height="1"/>
	  <rect x="2.5" y="1" width="1" height="1"/>
	  <rect id="a" x="0.5" y="2.6" width="1" height="1"/><rect id="b" x="1.6" y="2.6" width="1" height="1"/>
	</svg>`
	out := string(run(t, svg, scaled(DefaultOptions())).PreviewSVG)
	for _, want := range []string{
		`id="drawing-nose"`, `id="drawing-shape-2"`, `id="drawing-a"`, `id="drawing-b"`,
		`id="template-nose"`, `id="template-opening-2"`, `id="template-a-and-b"`,
		`id="cut-nose"`, `id="cut-opening-2"`, `id="cut-a-and-b"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("preview missing %s", want)
		}
	}
	// 4 drawing + 3 template + 3 cut paths, no groups.
	if n := strings.Count(out, "<path "); n != 10 {
		t.Errorf("got %d paths, want 10", n)
	}
	if strings.Contains(out, "<g>") || strings.Contains(out, "<g ") && strings.Count(out, "<g ") > 1 {
		t.Error("only the legend may be grouped")
	}
}
