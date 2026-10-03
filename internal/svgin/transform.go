package svgin

import (
	"fmt"
	"math"
	"strings"

	"github.com/brennanMKE/Bushwhack/internal/pathdata"
)

// matrix is an SVG affine transform [a b c d e f]:
// x' = a*x + c*y + e, y' = b*x + d*y + f.
type matrix [6]float64

var identity = matrix{1, 0, 0, 1, 0, 0}

// mul returns m * n (n applied first).
func (m matrix) mul(n matrix) matrix {
	return matrix{
		m[0]*n[0] + m[2]*n[1],
		m[1]*n[0] + m[3]*n[1],
		m[0]*n[2] + m[2]*n[3],
		m[1]*n[2] + m[3]*n[3],
		m[0]*n[4] + m[2]*n[5] + m[4],
		m[1]*n[4] + m[3]*n[5] + m[5],
	}
}

func (m matrix) apply(p pathdata.Point) pathdata.Point {
	return pathdata.Point{X: m[0]*p.X + m[2]*p.Y + m[4], Y: m[1]*p.X + m[3]*p.Y + m[5]}
}

func (m matrix) applySubpath(sp pathdata.Subpath) pathdata.Subpath {
	if m == identity {
		return sp
	}
	sp.Start = m.apply(sp.Start)
	for i := range sp.Segs {
		s := &sp.Segs[i]
		s.P3 = m.apply(s.P3)
		if s.Cubic {
			s.P1 = m.apply(s.P1)
			s.P2 = m.apply(s.P2)
		}
	}
	return sp
}

// parseTransform parses a transform list such as
// "translate(10,20) rotate(45 5 5) scale(2)".
func parseTransform(s string) (matrix, error) {
	m := identity
	rest := strings.TrimSpace(s)
	for rest != "" {
		open := strings.IndexByte(rest, '(')
		if open < 0 {
			return identity, fmt.Errorf("missing '(' in %q", s)
		}
		closeIdx := strings.IndexByte(rest, ')')
		if closeIdx < open {
			return identity, fmt.Errorf("missing ')' in %q", s)
		}
		name := strings.TrimSpace(rest[:open])
		args, err := pathdata.Numbers(rest[open+1 : closeIdx])
		if err != nil {
			return identity, err
		}
		t, err := transformFunc(name, args)
		if err != nil {
			return identity, err
		}
		m = m.mul(t)
		rest = strings.TrimLeft(rest[closeIdx+1:], " \t\r\n,")
	}
	return m, nil
}

func transformFunc(name string, a []float64) (matrix, error) {
	bad := fmt.Errorf("%s() has %d arguments", name, len(a))
	rad := func(deg float64) float64 { return deg * math.Pi / 180 }
	switch name {
	case "matrix":
		if len(a) != 6 {
			return identity, bad
		}
		return matrix{a[0], a[1], a[2], a[3], a[4], a[5]}, nil
	case "translate":
		switch len(a) {
		case 1:
			return matrix{1, 0, 0, 1, a[0], 0}, nil
		case 2:
			return matrix{1, 0, 0, 1, a[0], a[1]}, nil
		}
	case "scale":
		switch len(a) {
		case 1:
			return matrix{a[0], 0, 0, a[0], 0, 0}, nil
		case 2:
			return matrix{a[0], 0, 0, a[1], 0, 0}, nil
		}
	case "rotate":
		if len(a) == 1 || len(a) == 3 {
			c, s := math.Cos(rad(a[0])), math.Sin(rad(a[0]))
			r := matrix{c, s, -s, c, 0, 0}
			if len(a) == 3 {
				return matrix{1, 0, 0, 1, a[1], a[2]}.mul(r).mul(matrix{1, 0, 0, 1, -a[1], -a[2]}), nil
			}
			return r, nil
		}
	case "skewX":
		if len(a) == 1 {
			return matrix{1, 0, math.Tan(rad(a[0])), 1, 0, 0}, nil
		}
	case "skewY":
		if len(a) == 1 {
			return matrix{1, math.Tan(rad(a[0])), 0, 1, 0, 0}, nil
		}
	default:
		return identity, fmt.Errorf("unknown transform %q", name)
	}
	return identity, bad
}
