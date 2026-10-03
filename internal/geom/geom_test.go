package geom

import "testing"

func sq(x, y, s float64) []Pt { return []Pt{{x, y}, {x + s, y}, {x + s, y + s}, {x, y + s}} }

func TestSplit(t *testing.T) {
	// An O (square with a counter), a dot inside the counter, and a separate square.
	p := Union(append(append(FromRings([][]Pt{sq(0, 0, 10), sq(2, 2, 6)}), FromRings([][]Pt{sq(4, 4, 2)})...),
		FromRings([][]Pt{sq(20, 0, 3)})...))
	pieces := Split(p)
	if len(pieces) != 3 {
		t.Fatalf("got %d pieces, want 3", len(pieces))
	}
	rings := map[int]int{}
	for _, pc := range pieces {
		rings[len(pc)]++
	}
	if rings[2] != 1 || rings[1] != 2 {
		t.Errorf("want one piece with a hole and two without, got %v", rings)
	}
	if !Contains(p, Pt{5, 5}) || Contains(p, Pt{3, 3}) || !Contains(p, Pt{1, 1}) {
		t.Error("Contains disagrees with the even-odd picture")
	}
}
