package geom

import (
	"context"
	"math"
)

type ring struct {
	pts []Pt
	box Rect
}

func rings(p Polys) []ring {
	out := make([]ring, 0, len(p))
	for _, r := range ToRings(p) {
		if len(r) < 2 {
			continue
		}
		box := EmptyRect()
		for _, pt := range r {
			box = box.Add(pt)
		}
		out = append(out, ring{r, box})
	}
	return out
}

// Distance returns the minimum distance between the boundaries of a and b and
// the closest pair of points. It does not detect containment or overlap; use
// Intersect for that. limit prunes work: pairs farther apart than limit are
// skipped, and if nothing is closer than limit the result is (+Inf, ...).
// Pass +Inf for an exact answer.
func Distance(ctx context.Context, a, b Polys, limit float64) (float64, Pt, Pt, error) {
	best := limit
	var pa, pb Pt
	ra, rb := rings(a), rings(b)
	checks := 0
	for _, x := range ra {
		for _, y := range rb {
			if x.box.gap(y.box) >= best {
				continue
			}
			for i := range x.pts {
				a0, a1 := x.pts[i], x.pts[(i+1)%len(x.pts)]
				sa := EmptyRect().Add(a0).Add(a1)
				if sa.gap(y.box) >= best {
					continue
				}
				for j := range y.pts {
					b0, b1 := y.pts[j], y.pts[(j+1)%len(y.pts)]
					if sa.gap(EmptyRect().Add(b0).Add(b1)) >= best {
						continue
					}
					d, p, q := segSeg(a0, a1, b0, b1)
					if d < best {
						best, pa, pb = d, p, q
					}
				}
				checks++
				if checks&1023 == 0 {
					if err := ctx.Err(); err != nil {
						return 0, Pt{}, Pt{}, err
					}
				}
			}
		}
	}
	if best >= limit {
		return math.Inf(1), Pt{}, Pt{}, nil
	}
	return best, pa, pb, nil
}

func segSeg(a0, a1, b0, b1 Pt) (float64, Pt, Pt) {
	if p, ok := segIntersect(a0, a1, b0, b1); ok {
		return 0, p, p
	}
	best := math.Inf(1)
	var pa, pb Pt
	try := func(d float64, p, q Pt) {
		if d < best {
			best, pa, pb = d, p, q
		}
	}
	q, d := closest(a0, b0, b1)
	try(d, a0, q)
	q, d = closest(a1, b0, b1)
	try(d, a1, q)
	q, d = closest(b0, a0, a1)
	try(d, q, b0)
	q, d = closest(b1, a0, a1)
	try(d, q, b1)
	return best, pa, pb
}

func closest(p, a, b Pt) (Pt, float64) {
	dx, dy := b.X-a.X, b.Y-a.Y
	l2 := dx*dx + dy*dy
	t := 0.0
	if l2 > 0 {
		t = math.Max(0, math.Min(1, ((p.X-a.X)*dx+(p.Y-a.Y)*dy)/l2))
	}
	q := Pt{a.X + t*dx, a.Y + t*dy}
	return q, math.Hypot(p.X-q.X, p.Y-q.Y)
}

func segIntersect(a0, a1, b0, b1 Pt) (Pt, bool) {
	rx, ry := a1.X-a0.X, a1.Y-a0.Y
	sx, sy := b1.X-b0.X, b1.Y-b0.Y
	den := rx*sy - ry*sx
	if den == 0 {
		return Pt{}, false
	}
	qx, qy := b0.X-a0.X, b0.Y-a0.Y
	t := (qx*sy - qy*sx) / den
	u := (qx*ry - qy*rx) / den
	if t < 0 || t > 1 || u < 0 || u > 1 {
		return Pt{}, false
	}
	return Pt{a0.X + t*rx, a0.Y + t*ry}, true
}

// BoundsGap is the distance between the bounding boxes of a and b.
func BoundsGap(a, b Polys) float64 { return Bounds(a).gap(Bounds(b)) }
