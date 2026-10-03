// Package units parses user-entered lengths such as "5/16", "1 1/4",
// "0.3125", "8mm" or "1/2in" into inches.
package units

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseLength returns s in inches. A bare number is inches.
func ParseLength(s string) (float64, error) {
	orig := s
	s = strings.ToLower(strings.TrimSpace(s))
	factor := 1.0
	for _, u := range []struct {
		suffix string
		f      float64
	}{{"mm", 1 / 25.4}, {"cm", 1 / 2.54}, {"inches", 1}, {"inch", 1}, {"in", 1}, {"\"", 1}, {"″", 1}} {
		if strings.HasSuffix(s, u.suffix) {
			s = strings.TrimSpace(strings.TrimSuffix(s, u.suffix))
			factor = u.f
			break
		}
	}
	if s == "" {
		return 0, fmt.Errorf("enter a length, e.g. 5/16 or 8mm")
	}
	v, err := parseMixed(s)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, fmt.Errorf("%q is not a length; use e.g. 5/16, 1 1/4, 0.3125 or 8mm", orig)
	}
	return v * factor, nil
}

// parseMixed handles "a", "a/b", "w a/b" and "w-a/b".
func parseMixed(s string) (float64, error) {
	if !strings.Contains(s, "/") {
		return strconv.ParseFloat(s, 64)
	}
	whole := 0.0
	frac := s
	if i := strings.LastIndexAny(s, " -"); i > 0 {
		w, err := strconv.ParseFloat(strings.TrimSpace(s[:i]), 64)
		if err != nil {
			return 0, err
		}
		whole, frac = w, strings.TrimSpace(s[i+1:])
	}
	parts := strings.Split(frac, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("bad fraction")
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, err
	}
	d, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil || d == 0 {
		return 0, fmt.Errorf("bad fraction")
	}
	if whole < 0 {
		return whole - n/d, nil
	}
	return whole + n/d, nil
}

// Fraction formats inches as the nearest 1/64 fraction when exact to within
// 0.0005", else as a decimal: 0.3125 -> 5/16", 0.1 -> 0.100".
func Fraction(in float64) string {
	sixtyFourths := math.Round(in * 64)
	if math.Abs(sixtyFourths/64-in) > 0.0005 || sixtyFourths == 0 {
		return strconv.FormatFloat(in, 'f', 3, 64) + `"`
	}
	n := int(sixtyFourths)
	whole, num, den := n/64, n%64, 64
	for num != 0 && num%2 == 0 {
		num /= 2
		den /= 2
	}
	switch {
	case num == 0:
		return fmt.Sprintf(`%d"`, whole)
	case whole == 0:
		return fmt.Sprintf(`%d/%d"`, num, den)
	default:
		return fmt.Sprintf(`%d %d/%d"`, whole, num, den)
	}
}
