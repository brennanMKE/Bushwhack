// Package svgin walks an SVG document and returns its filled shapes as
// subpaths in root user units, with every nested transform applied.
package svgin

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/brennanMKE/Bushwhack/internal/pathdata"
)

// Limits caps the work a single document may cause.
type Limits struct {
	MaxElements int // total start elements
	MaxCommands int // total path commands across all shapes
}

// Shape is one drawing element (<path>, <rect>, ...) in document order.
type Shape struct {
	Element  string
	ID       string
	Name     string             // inkscape:label, else id
	Subpaths []pathdata.Subpath // root user units
}

// Document is the parsed result.
type Document struct {
	// Canvas in root user units: viewBox if present, else 0 0 width height.
	MinX, MinY, Width, Height float64
	HasCanvas                 bool
	// Physical page size in inches if the root declares absolute units
	// (in, mm, cm, pt, pc), else 0.
	PhysicalWidthIn, PhysicalHeightIn float64
	Shapes                            []Shape
	Warnings                          []string
}

// ErrLimit marks errors caused by exceeding a Limits cap.
var ErrLimit = errors.New("limit exceeded")

// ErrUnsafe marks documents rejected for DOCTYPE/ENTITY declarations.
var ErrUnsafe = errors.New("DOCTYPE and ENTITY declarations are not allowed")

const svgNS = "http://www.w3.org/2000/svg"

// Parse reads an SVG document.
func Parse(data []byte, lim Limits) (*Document, error) {
	upper := bytes.ToUpper(data)
	if bytes.Contains(upper, []byte("<!DOCTYPE")) || bytes.Contains(upper, []byte("<!ENTITY")) {
		return nil, ErrUnsafe
	}
	w := &walker{
		dec:     xml.NewDecoder(bytes.NewReader(data)),
		lim:     lim,
		ignored: map[string]int{},
	}
	w.dec.CharsetReader = charsetReader
	if err := w.run(); err != nil {
		return nil, err
	}
	w.finishWarnings()
	return &w.doc, nil
}

type walker struct {
	dec      *xml.Decoder
	lim      Limits
	doc      Document
	elements int
	commands int
	ignored  map[string]int
	sawRoot  bool
	pathErrs int
}

func (w *walker) run() error {
	for {
		tok, err := w.dec.Token()
		if err == io.EOF {
			if !w.sawRoot {
				return errors.New("no <svg> root element found")
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("invalid XML: %w", err)
		}
		switch t := tok.(type) {
		case xml.Directive:
			return ErrUnsafe
		case xml.StartElement:
			if t.Name.Local != "svg" || w.sawRoot {
				return errors.New("root element must be <svg>")
			}
			w.sawRoot = true
			w.root(t)
			if err := w.children(identity); err != nil {
				return err
			}
		}
	}
}

func (w *walker) root(t xml.StartElement) {
	vb := attr(t, "viewBox")
	wAttr, hAttr := attr(t, "width"), attr(t, "height")
	width, wOK := userLength(wAttr)
	height, hOK := userLength(hAttr)
	if vb != "" {
		nums, err := pathdata.Numbers(vb)
		if err == nil && len(nums) == 4 && nums[2] > 0 && nums[3] > 0 {
			w.doc.MinX, w.doc.MinY, w.doc.Width, w.doc.Height = nums[0], nums[1], nums[2], nums[3]
			w.doc.HasCanvas = true
		} else {
			w.warn("Ignored an invalid viewBox")
		}
	}
	if !w.doc.HasCanvas && wOK && hOK && width > 0 && height > 0 {
		w.doc.Width, w.doc.Height = width, height
		w.doc.HasCanvas = true
	}
	pw, okw := physicalInches(wAttr)
	ph, okh := physicalInches(hAttr)
	if okw && okh && pw > 0 && ph > 0 {
		w.doc.PhysicalWidthIn, w.doc.PhysicalHeightIn = pw, ph
	}
}

var skipContainers = map[string]bool{
	"defs": true, "clipPath": true, "mask": true, "symbol": true, "marker": true,
	"pattern": true, "linearGradient": true, "radialGradient": true, "style": true,
	"script": true, "metadata": true, "title": true, "desc": true, "filter": true,
	"foreignObject": true,
}

var reportIgnored = map[string]bool{
	"text": true, "image": true, "use": true, "line": true, "foreignObject": true,
}

// children walks until the end of the current element.
func (w *walker) children(m matrix) error {
	for {
		tok, err := w.dec.Token()
		if err != nil {
			if err == io.EOF {
				return errors.New("invalid XML: unexpected end of document")
			}
			return fmt.Errorf("invalid XML: %w", err)
		}
		switch t := tok.(type) {
		case xml.EndElement:
			return nil
		case xml.Directive:
			return ErrUnsafe
		case xml.StartElement:
			if err := w.element(t, m); err != nil {
				return err
			}
		}
	}
}

func (w *walker) element(t xml.StartElement, parent matrix) error {
	w.elements++
	if w.lim.MaxElements > 0 && w.elements > w.lim.MaxElements {
		return fmt.Errorf("%w: more than %d elements", ErrLimit, w.lim.MaxElements)
	}
	name := t.Name.Local
	if (t.Name.Space != "" && t.Name.Space != svgNS) || skipContainers[name] || hidden(t) {
		if reportIgnored[name] {
			w.ignored[name]++
		}
		return w.dec.Skip()
	}
	m := parent
	if tr := attr(t, "transform"); tr != "" {
		local, err := parseTransform(tr)
		if err != nil {
			w.warn("Ignored an invalid transform: " + err.Error())
		} else {
			m = parent.mul(local)
		}
	}
	switch name {
	case "g", "a", "switch":
		return w.children(m)
	case "svg":
		// Nested viewport: honour x/y; a nested viewBox is not supported.
		x, _ := userLength(attr(t, "x"))
		y, _ := userLength(attr(t, "y"))
		if attr(t, "viewBox") != "" {
			w.warn("A nested <svg> viewBox was ignored; its shapes may be mis-scaled")
		}
		return w.children(m.mul(matrix{1, 0, 0, 1, x, y}))
	case "path", "rect", "circle", "ellipse", "polygon", "polyline":
		d, err := shapeData(t)
		if err != nil {
			w.warn(fmt.Sprintf("Skipped a <%s>: %v", name, err))
		} else if d != "" {
			if err := w.addShape(t, d, m); err != nil {
				return err
			}
		}
		return w.dec.Skip()
	default:
		if reportIgnored[name] {
			w.ignored[name]++
		}
		return w.dec.Skip()
	}
}

func (w *walker) addShape(t xml.StartElement, d string, m matrix) error {
	remaining := 0
	if w.lim.MaxCommands > 0 {
		remaining = w.lim.MaxCommands - w.commands
		if remaining <= 0 {
			return fmt.Errorf("%w: more than %d path commands", ErrLimit, w.lim.MaxCommands)
		}
	}
	subs, err := pathdata.Parse(d, remaining)
	if errors.Is(err, pathdata.ErrTooManyCommands) {
		return fmt.Errorf("%w: more than %d path commands", ErrLimit, w.lim.MaxCommands)
	}
	if err != nil {
		w.pathErrs++
	}
	for _, sp := range subs {
		w.commands += len(sp.Segs) + 1
	}
	if len(subs) == 0 {
		return nil
	}
	for i := range subs {
		subs[i] = m.applySubpath(subs[i])
	}
	name := attr(t, "id")
	for _, a := range t.Attr {
		if a.Name.Local == "label" && strings.Contains(a.Name.Space, "inkscape") && strings.TrimSpace(a.Value) != "" {
			name = strings.TrimSpace(a.Value)
		}
	}
	if len(name) > 60 {
		name = name[:60]
	}
	w.doc.Shapes = append(w.doc.Shapes, Shape{Element: t.Name.Local, ID: attr(t, "id"), Name: name, Subpaths: subs})
	return nil
}

func (w *walker) warn(s string) {
	for _, existing := range w.doc.Warnings {
		if existing == s {
			return
		}
	}
	w.doc.Warnings = append(w.doc.Warnings, s)
}

func (w *walker) finishWarnings() {
	names := make([]string, 0, len(w.ignored))
	for n := range w.ignored {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		why := "not a filled shape"
		switch n {
		case "text":
			why = "convert text to paths (outlines) first"
		case "image":
			why = "raster images can't be offset"
		case "use":
			why = "unlink clones/symbols first"
		case "line":
			why = "a line has no area"
		}
		w.warn(fmt.Sprintf("Ignored %d <%s> element(s): %s", w.ignored[n], n, why))
	}
	if w.pathErrs > 0 {
		w.warn(fmt.Sprintf("%d path(s) had malformed data and were read up to the error", w.pathErrs))
	}
}

func attr(t xml.StartElement, name string) string {
	for _, a := range t.Attr {
		if a.Name.Local == name && (a.Name.Space == "" || a.Name.Space == svgNS) {
			return strings.TrimSpace(a.Value)
		}
	}
	return ""
}

func hidden(t xml.StartElement) bool {
	if attr(t, "display") == "none" || attr(t, "visibility") == "hidden" {
		return true
	}
	style := strings.ReplaceAll(attr(t, "style"), " ", "")
	return strings.Contains(style, "display:none") || strings.Contains(style, "visibility:hidden")
}

// shapeData converts a basic shape element to path data.
func shapeData(t xml.StartElement) (string, error) {
	num := func(name string) float64 {
		v, _ := userLength(attr(t, name))
		return v
	}
	f := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	switch t.Name.Local {
	case "path":
		return attr(t, "d"), nil
	case "rect":
		x, y, wd, ht := num("x"), num("y"), num("width"), num("height")
		if wd <= 0 || ht <= 0 {
			return "", nil
		}
		rxs, rys := attr(t, "rx"), attr(t, "ry")
		rx, ry := num("rx"), num("ry")
		if rxs == "" || rxs == "auto" {
			rx = ry
		}
		if rys == "" || rys == "auto" {
			ry = rx
		}
		rx, ry = math.Min(math.Max(rx, 0), wd/2), math.Min(math.Max(ry, 0), ht/2)
		if rx == 0 || ry == 0 {
			return fmt.Sprintf("M%s %sH%sV%sH%sZ", f(x), f(y), f(x+wd), f(y+ht), f(x)), nil
		}
		a := fmt.Sprintf("A%s %s 0 0 1 ", f(rx), f(ry))
		return fmt.Sprintf("M%s %sH%s%s%s %sV%s%s%s %sH%s%s%s %sV%s%s%s %sZ",
			f(x+rx), f(y), f(x+wd-rx), a, f(x+wd), f(y+ry), f(y+ht-ry), a, f(x+wd-rx), f(y+ht),
			f(x+rx), a, f(x), f(y+ht-ry), f(y+ry), a, f(x+rx), f(y)), nil
	case "circle", "ellipse":
		cx, cy := num("cx"), num("cy")
		rx, ry := num("r"), num("r")
		if t.Name.Local == "ellipse" {
			rx, ry = num("rx"), num("ry")
			if attr(t, "rx") == "" || attr(t, "rx") == "auto" {
				rx = ry
			}
			if attr(t, "ry") == "" || attr(t, "ry") == "auto" {
				ry = rx
			}
		}
		if rx <= 0 || ry <= 0 {
			return "", nil
		}
		a := fmt.Sprintf("A%s %s 0 1 1 ", f(rx), f(ry))
		return fmt.Sprintf("M%s %s%s%s %s%s%s %sZ", f(cx-rx), f(cy), a, f(cx+rx), f(cy), a, f(cx-rx), f(cy)), nil
	case "polygon", "polyline":
		nums, err := pathdata.Numbers(attr(t, "points"))
		if err != nil {
			return "", err
		}
		if len(nums) < 4 {
			return "", nil
		}
		var b strings.Builder
		for i := 0; i+1 < len(nums); i += 2 {
			if i == 0 {
				b.WriteString("M")
			} else {
				b.WriteString("L")
			}
			b.WriteString(f(nums[i]) + " " + f(nums[i+1]))
		}
		if t.Name.Local == "polygon" {
			b.WriteString("Z")
		}
		return b.String(), nil
	}
	return "", nil
}

var unitPx = map[string]float64{
	"": 1, "px": 1, "in": 96, "mm": 96 / 25.4, "cm": 96 / 2.54, "pt": 96.0 / 72, "pc": 16,
}

func splitLength(s string) (float64, string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, "", false
	}
	i := len(s)
	for i > 0 && (s[i-1] >= 'a' && s[i-1] <= 'z' || s[i-1] >= 'A' && s[i-1] <= 'Z' || s[i-1] == '%') {
		i--
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s[:i]), 64)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, "", false
	}
	return v, strings.ToLower(s[i:]), true
}

// userLength converts a length attribute to user units (CSS px).
func userLength(s string) (float64, bool) {
	v, unit, ok := splitLength(s)
	if !ok {
		return 0, false
	}
	k, ok := unitPx[unit]
	if !ok {
		return 0, false
	}
	return v * k, true
}

// physicalInches converts a length with an absolute unit to inches. Unitless
// and px lengths are not physical: design tools disagree on px per inch.
func physicalInches(s string) (float64, bool) {
	v, unit, ok := splitLength(s)
	if !ok {
		return 0, false
	}
	switch unit {
	case "in":
		return v, true
	case "mm":
		return v / 25.4, true
	case "cm":
		return v / 2.54, true
	case "pt":
		return v / 72, true
	case "pc":
		return v / 6, true
	}
	return 0, false
}

func charsetReader(label string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(label) {
	case "utf-8", "utf8", "us-ascii", "ascii":
		return input, nil
	case "iso-8859-1", "latin1", "latin-1", "windows-1252", "cp1252":
		b, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		runes := make([]rune, len(b))
		for i, c := range b {
			runes[i] = rune(c)
		}
		return strings.NewReader(string(runes)), nil
	}
	return nil, fmt.Errorf("unsupported encoding %q", label)
}
