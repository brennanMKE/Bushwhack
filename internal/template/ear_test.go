package template

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	clipper "github.com/bolom009/go-clipper2"
	"github.com/brennanMKE/Bushwhack/internal/geom"
)

// The black cat's ears are mirror images, so its template must be too.
// go-clipper2's union once filled in the right ear at a 3/8 in bushing.
func TestSymmetricEarsSurvive(t *testing.T) {
	b, err := os.ReadFile("../../patterns/halloween/black-cat.svg")
	if err != nil {
		t.Fatal(err)
	}
	for _, bush := range []float64{0.3125, 0.375, 0.5} {
		o := DefaultOptions()
		o.BushingOD, o.Fit, o.Size = bush, FitArtwork, 5
		r, err := Process(context.Background(), b, o)
		if err != nil {
			t.Fatal(err)
		}
		tmpl := rawPaths(t, r.Layers.Template)
		// Count template material on a grid over each ear (point-in-polygon
		// only, so no boolean op can hide the bug).
		material := func(minX float64) (n int) {
			for x := minX; x < minX+0.6; x += 0.01 {
				for y := 3.0; y < 3.7; y += 0.01 {
					if !geom.Contains(tmpl, geom.Pt{X: x, Y: y}) {
						n++
					}
				}
			}
			return
		}
		left, right := material(1.0), material(4.4)
		if left < 200 || math.Abs(float64(left-right)) > 0.02*float64(left) {
			t.Errorf("bushing %.4f: ear material left %d, right %d grid points", bush, left, right)
		}
	}
}

// rawPaths reads our own M/L/Z path data without any boolean clean-up.
func rawPaths(t *testing.T, d string) geom.Polys {
	t.Helper()
	var out geom.Polys
	for _, sub := range strings.Split(d, "Z") {
		var path clipper.Path64
		for _, c := range strings.FieldsFunc(sub, func(r rune) bool { return r == 'M' || r == 'L' }) {
			var x, y float64
			if _, err := fmt.Sscan(c, &x, &y); err != nil {
				t.Fatalf("bad path %q: %v", c, err)
			}
			path = append(path, clipper.Point64{X: int64(math.Round(x * geom.Scale)), Y: int64(math.Round(y * geom.Scale))})
		}
		if len(path) > 2 {
			out = append(out, path)
		}
	}
	return out
}
