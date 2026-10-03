package units

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// testdata/units.json is shared with web/src/lib/units.test.ts so the Go and
// browser parsers agree.
func TestSharedCases(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/units.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input  string  `json:"input"`
		Inches float64 `json:"inches"`
		Error  bool    `json:"error"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		got, err := ParseLength(c.Input)
		switch {
		case c.Error && err == nil:
			t.Errorf("%q: want error, got %v", c.Input, got)
		case !c.Error && err != nil:
			t.Errorf("%q: %v", c.Input, err)
		case !c.Error && math.Abs(got-c.Inches) > 1e-9:
			t.Errorf("%q = %v, want %v", c.Input, got, c.Inches)
		}
	}
}

func TestFraction(t *testing.T) {
	for in, want := range map[float64]string{0.3125: `5/16"`, 0.09375: `3/32"`, 1.25: `1 1/4"`, 0.1: `0.100"`, 2: `2"`} {
		if got := Fraction(in); got != want {
			t.Errorf("Fraction(%v) = %s, want %s", in, got, want)
		}
	}
}
