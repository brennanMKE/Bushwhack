package pathdata

import (
	"math"
	"testing"
)

func end(t *testing.T, d string) (Point, []Subpath) {
	t.Helper()
	subs, err := Parse(d, 0)
	if err != nil {
		t.Fatalf("%q: %v", d, err)
	}
	if len(subs) == 0 {
		t.Fatalf("%q: no subpaths", d)
	}
	last := subs[len(subs)-1]
	return last.Segs[len(last.Segs)-1].P3, subs
}

func eq(a, b Point) bool { return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9 }

func TestCommands(t *testing.T) {
	cases := []struct {
		d    string
		want Point
	}{
		{"M1 2 L3 4", Point{3, 4}},
		{"m1 2 l3 4", Point{4, 6}},
		{"M0 0 H5 V7", Point{5, 7}},
		{"M1 1 h5 v7", Point{6, 8}},
		{"M0 0 C1 1 2 2 3 3", Point{3, 3}},
		{"M0 0 c1 1 2 2 3 3 s1 1 2 2", Point{5, 5}},
		{"M0 0 Q1 1 2 0 T4 0", Point{4, 0}},
		{"M0 0 q1 1 2 0 t2 0", Point{4, 0}},
		{"M0 0 A5 5 0 0 1 10 0", Point{10, 0}},
		{"M0 0 a5 5 0 0 1 10 0", Point{10, 0}},
		{"M0 0 1 1 2 2", Point{2, 2}},          // implicit lineto after M
		{"m0 0 1 1 1 1", Point{2, 2}},          // implicit relative lineto
		{"M.5.5L-1-2", Point{-1, -2}},          // compact numbers
		{"M1e-3 0L1E1,2e0", Point{10, 2}},      // exponents
		{"M0 0a5 5 0 1 0 10 0", Point{10, 0}},  // flags spaced
		{"M0 0a5 5 0 1010 0", Point{10, 0}},    // flags packed
		{"M0 0a5,5,0,0,1,10,0", Point{10, 0}},  // commas
		{"M0 0L10 0 10 10Z l1 1", Point{1, 1}}, // command after Z starts at subpath start
	}
	for _, c := range cases {
		got, _ := end(t, c.d)
		if !eq(got, c.want) {
			t.Errorf("%q ends at %v, want %v", c.d, got, c.want)
		}
	}
}

func TestArcMidpoint(t *testing.T) {
	// Half circle radius 5 from (0,0) to (10,0), sweep=1: goes through (5,-5)
	// in SVG's y-down space (clockwise on screen).
	_, subs := end(t, "M0 0 A5 5 0 0 1 10 0")
	segs := subs[0].Segs
	if len(segs) != 2 {
		t.Fatalf("got %d cubic pieces, want 2", len(segs))
	}
	if !eq(segs[0].P3, Point{5, -5}) {
		t.Errorf("midpoint %v, want (5,-5)", segs[0].P3)
	}
}

func TestArcRadiusTooSmall(t *testing.T) {
	// Radii are scaled up so the arc still reaches the end point.
	got, _ := end(t, "M0 0 A1 1 0 0 1 10 0")
	if !eq(got, Point{10, 0}) {
		t.Errorf("got %v", got)
	}
}

func TestClosedAndSubpaths(t *testing.T) {
	subs, err := Parse("M0 0H1V1Z M2 2H3V3", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 2 || !subs[0].Closed || subs[1].Closed {
		t.Fatalf("got %+v", subs)
	}
}

func TestErrors(t *testing.T) {
	for _, d := range []string{"L1 1", "M1", "M0 0 A5 5 0 2 1 10 0", "M0 0 Z 5"} {
		if _, err := Parse(d, 0); err == nil {
			t.Errorf("%q: expected error", d)
		}
	}
	if _, err := Parse("M0 0 L1 1 L2 2", 2); err != ErrTooManyCommands {
		t.Errorf("limit not enforced: %v", err)
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"M0 0L1 1Z", "m.5.5a1 1 0 1010 10", "M1e5-2C1 2 3 4 5 6s1 2 3 4q1 2 3 4t5 6"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, d string) {
		Parse(d, 10000)
	})
}
