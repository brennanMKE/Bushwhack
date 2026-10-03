package svgin

import (
	"math"
	"testing"

	"github.com/brennanMKE/Bushwhack/internal/pathdata"
)

func TestTransforms(t *testing.T) {
	cases := []struct {
		tr   string
		in   pathdata.Point
		want pathdata.Point
	}{
		{"translate(10 20)", pathdata.Point{X: 1, Y: 1}, pathdata.Point{X: 11, Y: 21}},
		{"translate(10)", pathdata.Point{X: 1, Y: 1}, pathdata.Point{X: 11, Y: 1}},
		{"scale(2)", pathdata.Point{X: 1, Y: 3}, pathdata.Point{X: 2, Y: 6}},
		{"scale(2,3)", pathdata.Point{X: 1, Y: 1}, pathdata.Point{X: 2, Y: 3}},
		{"rotate(90)", pathdata.Point{X: 1, Y: 0}, pathdata.Point{X: 0, Y: 1}},
		{"rotate(90 1 1)", pathdata.Point{X: 2, Y: 1}, pathdata.Point{X: 1, Y: 2}},
		{"skewX(45)", pathdata.Point{X: 0, Y: 1}, pathdata.Point{X: 1, Y: 1}},
		{"skewY(45)", pathdata.Point{X: 1, Y: 0}, pathdata.Point{X: 1, Y: 1}},
		{"matrix(1 0 0 1 5 6)", pathdata.Point{X: 0, Y: 0}, pathdata.Point{X: 5, Y: 6}},
		{"translate(10,0) scale(2)", pathdata.Point{X: 1, Y: 1}, pathdata.Point{X: 12, Y: 2}},
	}
	for _, c := range cases {
		m, err := parseTransform(c.tr)
		if err != nil {
			t.Fatalf("%s: %v", c.tr, err)
		}
		got := m.apply(c.in)
		if math.Abs(got.X-c.want.X) > 1e-9 || math.Abs(got.Y-c.want.Y) > 1e-9 {
			t.Errorf("%s(%v) = %v, want %v", c.tr, c.in, got, c.want)
		}
	}
}

func TestNestedGroups(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
	  <g transform="translate(10 0)"><g transform="scale(2)"><rect x="1" y="1" width="1" height="1"/></g></g>
	  <defs><rect width="5" height="5"/></defs>
	  <text>hi</text>
	  <rect width="5" height="5" style="display: none"/>
	</svg>`
	doc, err := Parse([]byte(svg), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Shapes) != 1 {
		t.Fatalf("got %d shapes, want 1", len(doc.Shapes))
	}
	if s := doc.Shapes[0].Subpaths[0].Start; s.X != 12 || s.Y != 2 {
		t.Errorf("start %v, want (12,2)", s)
	}
	if len(doc.Warnings) != 1 {
		t.Errorf("warnings %v", doc.Warnings)
	}
}

func TestCanvas(t *testing.T) {
	doc, err := Parse([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="6in" height="3in"><circle r="1"/></svg>`), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Width != 576 || doc.PhysicalWidthIn != 6 {
		t.Errorf("width %v physical %v", doc.Width, doc.PhysicalWidthIn)
	}
}

func TestLimits(t *testing.T) {
	if _, err := Parse([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><g/><g/><g/></svg>`), Limits{MaxElements: 2}); err == nil {
		t.Error("element limit not enforced")
	}
}
