// Package flatten turns line/cubic subpaths into polylines by adaptive
// subdivision, so the chord error stays under a tolerance at any curvature.
package flatten

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/brennanMKE/Bushwhack/internal/pathdata"
)

type Point = pathdata.Point

// ErrTooManyVertices is returned when the vertex budget runs out.
var ErrTooManyVertices = errors.New("too many vertices after flattening")

// Budget bounds the total vertices produced and is checked against ctx.
type Budget struct {
	Ctx       context.Context
	Remaining int // <= 0 means unlimited
	limited   bool
	emitted   int
}

// NewBudget returns a budget of max vertices (0 = unlimited).
func NewBudget(ctx context.Context, max int) *Budget {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Budget{Ctx: ctx, Remaining: max, limited: max > 0}
}

func (b *Budget) take() error {
	b.emitted++
	if b.limited {
		b.Remaining--
		if b.Remaining < 0 {
			return ErrTooManyVertices
		}
	}
	if b.emitted&4095 == 0 {
		if err := b.Ctx.Err(); err != nil {
			return fmt.Errorf("processing timed out: %w", err)
		}
	}
	return nil
}

const maxDepth = 18

// Subpath flattens sp with maximum chord deviation tol. The returned ring
// starts at sp.Start; a closing point equal to the start is dropped.
func Subpath(sp pathdata.Subpath, tol float64, b *Budget) ([]Point, error) {
	out := []Point{sp.Start}
	if err := b.take(); err != nil {
		return nil, err
	}
	cur := sp.Start
	for _, s := range sp.Segs {
		if s.Cubic {
			var err error
			out, err = cubic(out, cur, s.P1, s.P2, s.P3, tol, 0, b)
			if err != nil {
				return nil, err
			}
		} else {
			if err := b.take(); err != nil {
				return nil, err
			}
			out = append(out, s.P3)
		}
		cur = s.P3
	}
	if len(out) > 1 && out[len(out)-1] == out[0] {
		out = out[:len(out)-1]
	}
	return out, nil
}

func cubic(out []Point, p0, p1, p2, p3 Point, tol float64, depth int, b *Budget) ([]Point, error) {
	if depth >= maxDepth || flat(p0, p1, p2, p3, tol) {
		if err := b.take(); err != nil {
			return nil, err
		}
		return append(out, p3), nil
	}
	mid := func(a, c Point) Point { return Point{X: (a.X + c.X) / 2, Y: (a.Y + c.Y) / 2} }
	p01, p12, p23 := mid(p0, p1), mid(p1, p2), mid(p2, p3)
	p012, p123 := mid(p01, p12), mid(p12, p23)
	m := mid(p012, p123)
	out, err := cubic(out, p0, p01, p012, m, tol, depth+1, b)
	if err != nil {
		return nil, err
	}
	return cubic(out, m, p123, p23, p3, tol, depth+1, b)
}

// flat reports whether both control points lie within tol of the chord. The
// curve lies inside the hull of its control points, so this bounds the error.
func flat(p0, p1, p2, p3 Point, tol float64) bool {
	return distToSegment(p1, p0, p3) <= tol && distToSegment(p2, p0, p3) <= tol
}

func distToSegment(p, a, c Point) float64 {
	dx, dy := c.X-a.X, c.Y-a.Y
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return math.Hypot(p.X-a.X, p.Y-a.Y)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / l2
	t = math.Max(0, math.Min(1, t))
	return math.Hypot(p.X-(a.X+t*dx), p.Y-(a.Y+t*dy))
}
