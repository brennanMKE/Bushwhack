// Package pathdata parses SVG path "d" strings into absolute subpaths made of
// lines and cubic beziers. Quadratics are elevated to cubics exactly and
// elliptical arcs are converted to cubics, so every segment survives an affine
// transform unchanged in kind.
package pathdata

import (
	"errors"
	"fmt"
	"math"
)

// Point is a 2D point in user units.
type Point struct{ X, Y float64 }

// Segment is a line (Cubic == false, only P3 used) or a cubic bezier from the
// previous segment's end point through P1, P2 to P3.
type Segment struct {
	Cubic      bool
	P1, P2, P3 Point
}

// Subpath is one continuous run of segments starting at Start.
type Subpath struct {
	Start  Point
	Segs   []Segment
	Closed bool
}

// ErrTooManyCommands is returned when the command limit is exceeded.
var ErrTooManyCommands = errors.New("path has too many commands")

// Parse parses d. maxCommands <= 0 means no limit. On a syntax error it
// returns the subpaths parsed so far together with the error, matching how
// browsers render a path up to the first error.
func Parse(d string, maxCommands int) ([]Subpath, error) {
	p := parser{s: d}
	err := p.run(maxCommands)
	p.flush()
	return p.out, err
}

type parser struct {
	s   string
	pos int

	out     []Subpath
	cur     *Subpath
	pt      Point // current point
	start   Point // start of current subpath
	lastCtl Point // last control point, for S/T reflection
	lastOp  byte  // last command letter, upper-cased
}

func (p *parser) run(maxCommands int) error {
	var op byte
	count := 0
	for {
		p.skipSep()
		if p.pos >= len(p.s) {
			return nil
		}
		c := p.s[p.pos]
		if isCommand(c) {
			if op == 0 && c != 'M' && c != 'm' {
				return fmt.Errorf("path data must start with a moveto, not %q", c)
			}
			op = c
			p.pos++
		} else if op == 0 {
			return fmt.Errorf("path data must start with a command at offset %d", p.pos)
		} else if op == 'Z' || op == 'z' {
			return fmt.Errorf("unexpected number after Z at offset %d", p.pos)
		}
		count++
		if maxCommands > 0 && count > maxCommands {
			return ErrTooManyCommands
		}
		if err := p.command(op); err != nil {
			return err
		}
		// Implicit repeats after a moveto are linetos.
		if op == 'M' {
			op = 'L'
		} else if op == 'm' {
			op = 'l'
		}
	}
}

func isCommand(c byte) bool {
	switch c {
	case 'M', 'm', 'L', 'l', 'H', 'h', 'V', 'v', 'C', 'c', 'S', 's', 'Q', 'q', 'T', 't', 'A', 'a', 'Z', 'z':
		return true
	}
	return false
}

func (p *parser) command(op byte) error {
	rel := op >= 'a'
	up := op &^ 0x20
	var base Point
	if rel {
		base = p.pt
	}
	switch up {
	case 'Z':
		if p.cur != nil {
			p.cur.Closed = true
			p.flush()
		}
		p.pt = p.start
		p.lastCtl = p.pt
	case 'M':
		pt, err := p.point(base)
		if err != nil {
			return err
		}
		p.flush()
		p.cur = &Subpath{Start: pt}
		p.pt, p.start, p.lastCtl = pt, pt, pt
	case 'L':
		pt, err := p.point(base)
		if err != nil {
			return err
		}
		p.line(pt)
	case 'H':
		x, err := p.number()
		if err != nil {
			return err
		}
		p.line(Point{x + base.X, p.pt.Y})
	case 'V':
		y, err := p.number()
		if err != nil {
			return err
		}
		p.line(Point{p.pt.X, y + base.Y})
	case 'C':
		pts, err := p.points(3, base)
		if err != nil {
			return err
		}
		p.cubic(pts[0], pts[1], pts[2])
	case 'S':
		pts, err := p.points(2, base)
		if err != nil {
			return err
		}
		c1 := p.pt
		if p.lastOp == 'C' || p.lastOp == 'S' {
			c1 = reflect(p.lastCtl, p.pt)
		}
		p.cubic(c1, pts[0], pts[1])
	case 'Q':
		pts, err := p.points(2, base)
		if err != nil {
			return err
		}
		p.quad(pts[0], pts[1])
	case 'T':
		pt, err := p.point(base)
		if err != nil {
			return err
		}
		c := p.pt
		if p.lastOp == 'Q' || p.lastOp == 'T' {
			c = reflect(p.lastCtl, p.pt)
		}
		p.quad(c, pt)
	case 'A':
		rx, err := p.number()
		if err != nil {
			return err
		}
		ry, err := p.number()
		if err != nil {
			return err
		}
		rot, err := p.number()
		if err != nil {
			return err
		}
		large, err := p.flag()
		if err != nil {
			return err
		}
		sweep, err := p.flag()
		if err != nil {
			return err
		}
		end, err := p.point(base)
		if err != nil {
			return err
		}
		p.arc(rx, ry, rot, large, sweep, end)
	}
	p.lastOp = up
	return nil
}

func reflect(ctl, about Point) Point {
	return Point{2*about.X - ctl.X, 2*about.Y - ctl.Y}
}

// ensure starts an implicit subpath at the current point when drawing
// follows a Z (or nothing) without a moveto.
func (p *parser) ensure() {
	if p.cur == nil {
		p.cur = &Subpath{Start: p.pt}
		p.start = p.pt
	}
}

func (p *parser) flush() {
	if p.cur != nil {
		if len(p.cur.Segs) > 0 {
			p.out = append(p.out, *p.cur)
		}
		p.cur = nil
	}
}

func (p *parser) line(pt Point) {
	p.ensure()
	p.cur.Segs = append(p.cur.Segs, Segment{P3: pt})
	p.pt, p.lastCtl = pt, pt
}

func (p *parser) cubic(c1, c2, pt Point) {
	p.ensure()
	p.cur.Segs = append(p.cur.Segs, Segment{Cubic: true, P1: c1, P2: c2, P3: pt})
	p.pt, p.lastCtl = pt, c2
}

func (p *parser) quad(c, pt Point) {
	p0 := p.pt
	c1 := Point{p0.X + 2.0/3*(c.X-p0.X), p0.Y + 2.0/3*(c.Y-p0.Y)}
	c2 := Point{pt.X + 2.0/3*(c.X-pt.X), pt.Y + 2.0/3*(c.Y-pt.Y)}
	p.cubic(c1, c2, pt)
	p.lastCtl = c
}

// arc implements SVG 1.1 Appendix F.6 endpoint-to-center conversion and emits
// cubic approximations of at most 90 degrees each.
func (p *parser) arc(rx, ry, phiDeg float64, large, sweep bool, end Point) {
	p0 := p.pt
	if p0 == end {
		return
	}
	rx, ry = math.Abs(rx), math.Abs(ry)
	if rx == 0 || ry == 0 {
		p.line(end)
		return
	}
	phi := phiDeg * math.Pi / 180
	cosPhi, sinPhi := math.Cos(phi), math.Sin(phi)

	// F.6.5.1
	dx2, dy2 := (p0.X-end.X)/2, (p0.Y-end.Y)/2
	x1p := cosPhi*dx2 + sinPhi*dy2
	y1p := -sinPhi*dx2 + cosPhi*dy2

	// F.6.6 radius correction
	lambda := (x1p*x1p)/(rx*rx) + (y1p*y1p)/(ry*ry)
	if lambda > 1 {
		s := math.Sqrt(lambda)
		rx *= s
		ry *= s
	}

	// F.6.5.2
	num := rx*rx*ry*ry - rx*rx*y1p*y1p - ry*ry*x1p*x1p
	den := rx*rx*y1p*y1p + ry*ry*x1p*x1p
	coef := 0.0
	if den != 0 && num > 0 {
		coef = math.Sqrt(num / den)
	}
	if large == sweep {
		coef = -coef
	}
	cxp := coef * rx * y1p / ry
	cyp := -coef * ry * x1p / rx

	// F.6.5.3
	cx := cosPhi*cxp - sinPhi*cyp + (p0.X+end.X)/2
	cy := sinPhi*cxp + cosPhi*cyp + (p0.Y+end.Y)/2

	// F.6.5.5, F.6.5.6
	ux, uy := (x1p-cxp)/rx, (y1p-cyp)/ry
	vx, vy := (-x1p-cxp)/rx, (-y1p-cyp)/ry
	theta1 := math.Atan2(uy, ux)
	dtheta := math.Atan2(ux*vy-uy*vx, ux*vx+uy*vy)
	if !sweep && dtheta > 0 {
		dtheta -= 2 * math.Pi
	} else if sweep && dtheta < 0 {
		dtheta += 2 * math.Pi
	}

	n := int(math.Ceil(math.Abs(dtheta) / (math.Pi / 2)))
	if n < 1 {
		n = 1
	}
	step := dtheta / float64(n)
	k := 4.0 / 3 * math.Tan(step/4)
	ellipse := func(t float64) (Point, Point) {
		ct, st := math.Cos(t), math.Sin(t)
		pt := Point{cx + rx*ct*cosPhi - ry*st*sinPhi, cy + rx*ct*sinPhi + ry*st*cosPhi}
		d := Point{-rx*st*cosPhi - ry*ct*sinPhi, -rx*st*sinPhi + ry*ct*cosPhi}
		return pt, d
	}
	t := theta1
	a, da := ellipse(t)
	for i := 0; i < n; i++ {
		b, db := ellipse(t + step)
		if i == n-1 {
			b = end // land exactly on the requested end point
		}
		c1 := Point{a.X + k*da.X, a.Y + k*da.Y}
		c2 := Point{b.X - k*db.X, b.Y - k*db.Y}
		p.cubic(c1, c2, b)
		t += step
		a, da = b, db
	}
}

func (p *parser) points(n int, base Point) ([]Point, error) {
	pts := make([]Point, n)
	for i := range pts {
		pt, err := p.point(base)
		if err != nil {
			return nil, err
		}
		pts[i] = pt
	}
	return pts, nil
}

func (p *parser) point(base Point) (Point, error) {
	x, err := p.number()
	if err != nil {
		return Point{}, err
	}
	y, err := p.number()
	if err != nil {
		return Point{}, err
	}
	return Point{x + base.X, y + base.Y}, nil
}

func (p *parser) skipSep() {
	for p.pos < len(p.s) {
		switch p.s[p.pos] {
		case ' ', '\t', '\n', '\r', '\f', ',':
			p.pos++
		default:
			return
		}
	}
}

// flag reads an arc flag, which is a single 0 or 1 that may be packed against
// what follows ("10010 10").
func (p *parser) flag() (bool, error) {
	p.skipSep()
	if p.pos < len(p.s) {
		switch p.s[p.pos] {
		case '0':
			p.pos++
			return false, nil
		case '1':
			p.pos++
			return true, nil
		}
	}
	return false, fmt.Errorf("expected arc flag at offset %d", p.pos)
}

// number reads an SVG number: sign, digits, optional fraction, optional
// exponent. ".5.5" reads as 0.5 then 0.5; "-1-2" as -1 then -2.
func (p *parser) number() (float64, error) {
	p.skipSep()
	s, i := p.s, p.pos
	start := i
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		digits++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
			digits++
		}
	}
	if digits == 0 {
		return 0, fmt.Errorf("expected number at offset %d", start)
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		k := j
		for k < len(s) && s[k] >= '0' && s[k] <= '9' {
			k++
		}
		if k > j { // only consume the exponent if it has digits
			i = k
		}
	}
	v, err := parseFloat(s[start:i])
	if err != nil {
		return 0, fmt.Errorf("bad number %q at offset %d", s[start:i], start)
	}
	p.pos = i
	return v, nil
}
