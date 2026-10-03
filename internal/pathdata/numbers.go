package pathdata

import (
	"fmt"
	"math"
	"strconv"
)

func parseFloat(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, fmt.Errorf("bad number %q", s)
	}
	return v, nil
}

// Numbers parses a whitespace/comma separated list of SVG numbers, as used by
// viewBox, polygon points and transform arguments.
func Numbers(s string) ([]float64, error) {
	p := parser{s: s}
	var out []float64
	for {
		p.skipSep()
		if p.pos >= len(p.s) {
			return out, nil
		}
		v, err := p.number()
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
}
