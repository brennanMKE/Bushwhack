// Package geom wraps Clipper2 for the polygon work the pipeline needs:
// even-odd assembly, round-join offsets, booleans, areas, bounds and
// ring-to-ring distance. Coordinates are inches in float64 at the edges and
// fixed-point int64 (Scale units per inch) inside Clipper.
package geom

import (
	"math"

	clipper "github.com/bolom009/go-clipper2"
)

// Scale is the fixed-point resolution: 1e5 units per inch (0.00001").
const Scale = 1e5

// arcTolerance keeps round joins within 0.0005" of a true arc.
const arcTolerance = 0.0005 * Scale

// Pt is a point in inches.
type Pt struct{ X, Y float64 }

// Polys is a set of closed rings in Clipper's normalised form: outer rings
// have positive area, holes negative. Fill with non-zero.
type Polys = clipper.Paths64

// Rect is an axis-aligned box in inches.
type Rect struct{ MinX, MinY, MaxX, MaxY float64 }

func (r Rect) W() float64  { return r.MaxX - r.MinX }
func (r Rect) H() float64  { return r.MaxY - r.MinY }
func (r Rect) Empty() bool { return r.MaxX < r.MinX || r.MaxY < r.MinY }
func EmptyRect() Rect      { return Rect{math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)} }
func (r Rect) Add(p Pt) Rect {
	return Rect{math.Min(r.MinX, p.X), math.Min(r.MinY, p.Y), math.Max(r.MaxX, p.X), math.Max(r.MaxY, p.Y)}
}
func (r Rect) Union(o Rect) Rect {
	return Rect{math.Min(r.MinX, o.MinX), math.Min(r.MinY, o.MinY), math.Max(r.MaxX, o.MaxX), math.Max(r.MaxY, o.MaxY)}
}

// gap is the distance between two boxes (0 if they overlap).
func (r Rect) gap(o Rect) float64 {
	dx := math.Max(0, math.Max(o.MinX-r.MaxX, r.MinX-o.MaxX))
	dy := math.Max(0, math.Max(o.MinY-r.MaxY, r.MinY-o.MaxY))
	return math.Hypot(dx, dy)
}

func toInt(v float64) int64 { return int64(math.Round(v * Scale)) }

// FromRings converts float rings (inches) to normalised polygons using the
// even-odd rule, so a ring inside a ring becomes a hole.
func FromRings(rings [][]Pt) Polys {
	paths := make(clipper.Paths64, 0, len(rings))
	for _, r := range rings {
		if len(r) < 3 {
			continue
		}
		p := make(clipper.Path64, len(r))
		for i, pt := range r {
			p[i] = clipper.Point64{X: toInt(pt.X), Y: toInt(pt.Y)}
		}
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return nil
	}
	return clipper.UnionPaths64(paths, clipper.EvenOdd)
}

// ToRings converts polygons back to float rings in inches.
func ToRings(p Polys) [][]Pt {
	out := make([][]Pt, len(p))
	for i, path := range p {
		r := make([]Pt, len(path))
		for j, pt := range path {
			r[j] = Pt{float64(pt.X) / Scale, float64(pt.Y) / Scale}
		}
		out[i] = r
	}
	return out
}

// Offset grows (delta > 0) or shrinks (delta < 0) polygons by delta inches
// with round joins.
func Offset(p Polys, delta float64) Polys {
	if len(p) == 0 {
		return nil
	}
	if delta == 0 {
		return Union(p)
	}
	return clipper.InflatePaths64(p, delta*Scale, clipper.Round, clipper.Polygon,
		clipper.WithArcTolerance(arcTolerance))
}

// Union merges overlapping polygons.
func Union(p Polys) Polys {
	if len(p) == 0 {
		return nil
	}
	return clipper.UnionPaths64(p, clipper.NonZero)
}

// Difference returns a - b.
func Difference(a, b Polys) Polys {
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return a
	}
	return clipper.DifferenceWithClipPaths64(a, b, clipper.NonZero)
}

// Intersect returns a ∩ b.
func Intersect(a, b Polys) Polys {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	return clipper.IntersectWithClipPaths64(a, b, clipper.NonZero)
}

// Area returns the filled area in square inches (holes subtract).
func Area(p Polys) float64 {
	return clipper.AreaPaths64(p) / (Scale * Scale)
}

// Open removes features thinner than 2*r (morphological opening), used to
// ignore slivers when measuring lost area.
func Open(p Polys, r float64) Polys {
	return Offset(Offset(p, -r), r)
}

// OuterCount returns the number of outer rings (positive area).
func OuterCount(p Polys) int {
	n := 0
	for _, path := range p {
		if clipper.Area64(path) > 0 {
			n++
		}
	}
	return n
}

// Bounds returns the bounding box in inches.
func Bounds(p Polys) Rect {
	r := EmptyRect()
	for _, path := range p {
		for _, pt := range path {
			r = r.Add(Pt{float64(pt.X) / Scale, float64(pt.Y) / Scale})
		}
	}
	return r
}

// Translate shifts polygons by (dx, dy) inches.
func Translate(p Polys, dx, dy float64) Polys {
	return clipper.TranslatePaths64(p, toInt(dx), toInt(dy))
}

// VertexCount returns the total number of vertices.
func VertexCount(p Polys) int {
	n := 0
	for _, path := range p {
		n += len(path)
	}
	return n
}
