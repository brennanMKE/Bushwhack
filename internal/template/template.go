// Package template is the Bushwhack pipeline: parse an SVG, scale it to
// inches, offset every shape for the guide bushing and bit, check the result,
// simulate the cut and write the template and preview SVGs. It does no I/O.
package template

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/brennanMKE/Bushwhack/internal/flatten"
	"github.com/brennanMKE/Bushwhack/internal/geom"
	"github.com/brennanMKE/Bushwhack/internal/pathdata"
	"github.com/brennanMKE/Bushwhack/internal/svgin"
	"github.com/brennanMKE/Bushwhack/internal/svgout"
	"github.com/brennanMKE/Bushwhack/internal/units"
)

const (
	ModeHole  = "hole"
	ModePiece = "piece"

	FitCanvas   = "canvas"
	FitArtwork  = "artwork"
	FitDocument = "document"
)

// Options are the user's inputs. All lengths are inches.
type Options struct {
	BushingOD float64
	BitDia    float64
	Mode      string  // "hole" | "piece"
	Fit       string  // "canvas" | "artwork" | "document"
	Size      float64 // longest side for canvas/artwork fit
	Scale     float64 // inches per user unit; 0 = unset (overrides Fit)
	MinBridge float64
}

// DefaultOptions matches a common 5/16" bushing and 1/8" bit.
func DefaultOptions() Options {
	return Options{BushingOD: 0.3125, BitDia: 0.125, Mode: ModeHole, Fit: FitCanvas, Size: 6, MinBridge: 0.25}
}

// Limits bound the work one request may cause.
type Limits struct {
	MaxElements int
	MaxCommands int
	MaxVertices int
}

// DefaultLimits are the server's caps.
var DefaultLimits = Limits{MaxElements: 20000, MaxCommands: 200000, MaxVertices: 500000}

// ShapeReport describes one input shape.
type ShapeReport struct {
	Index            int     `json:"index"`
	Element          string  `json:"element"`
	Name             string  `json:"name"` // from the SVG id or inkscape:label; may be empty
	Path             string  `json:"path"` // drawing outline, page inches
	WidthIn          float64 `json:"widthIn"`
	HeightIn         float64 `json:"heightIn"`
	TemplateWidthIn  float64 `json:"templateWidthIn"`
	TemplateHeightIn float64 `json:"templateHeightIn"`
	LostAreaPct      float64 `json:"lostAreaPct"`
	OvercutPct       float64 `json:"overcutPct"`
}

// Bridge is a narrow strip of template material between two openings.
type Bridge struct {
	A, B   int     // 1-based shape indices
	DistIn float64 `json:"distIn"`
	P, Q   geom.Pt // closest points, page coordinates
}

// Layers are SVG path data strings in page inches, for clients that draw
// their own preview.
type Layers struct {
	Drawing  string `json:"drawing"`
	Template string `json:"template"`
	Cut      string `json:"cut"`
	Center   string `json:"center"`
	Lost     string `json:"lost"`
}

// Result is everything the caller needs to show and download.
type Result struct {
	Layers         Layers
	TemplateSVG    []byte
	PreviewSVG     []byte
	PageWidthIn    float64
	PageHeightIn   float64
	ScaleInPerUnit float64
	Offset         float64
	Shapes         []ShapeReport
	MinBridgeIn    float64 // +Inf when there is only one opening
	Bridges        []Bridge
	Warnings       []string // Notes' text, for simple clients
	Notes          []Note
}

// Note kinds.
const (
	NoteMerged     = "merged"
	NoteBridge     = "bridge"
	NoteIsland     = "island"
	NoteLost       = "lost"
	NoteUncuttable = "uncuttable"
	NoteOvercut    = "overcut"
	NotePage       = "page"
	NoteInput      = "input"
)

// Note is one warning with enough structure for a UI to point at it.
type Note struct {
	Kind   string `json:"kind"`
	Text   string `json:"text"`
	Shapes []int  `json:"shapes,omitempty"` // 1-based shape indices involved
	Bridge int    `json:"bridge"`           // index into Result.Bridges, or -1
}

// InputError is a problem with the user's options or file, as opposed to an
// internal failure. Servers map it to 4xx.
type InputError struct{ Msg string }

func (e *InputError) Error() string { return e.Msg }

func inputErr(format string, a ...any) error { return &InputError{fmt.Sprintf(format, a...)} }

// flattenTol is the chord tolerance at final size (inches).
const flattenTol = 0.002

// sliver is the width below which lost or extra area is ignored.
const sliver = 0.005

// Process runs the pipeline with the default limits.
func Process(ctx context.Context, input []byte, o Options) (*Result, error) {
	return ProcessWithLimits(ctx, input, o, DefaultLimits)
}

// ProcessWithLimits runs the pipeline.
func ProcessWithLimits(ctx context.Context, input []byte, o Options, lim Limits) (*Result, error) {
	if err := validate(&o); err != nil {
		return nil, err
	}
	doc, err := svgin.Parse(input, svgin.Limits{MaxElements: lim.MaxElements, MaxCommands: lim.MaxCommands})
	if err != nil {
		if errors.Is(err, svgin.ErrLimit) || errors.Is(err, svgin.ErrUnsafe) {
			return nil, inputErr("%s", limitMessage(err))
		}
		return nil, inputErr("Bushwhack couldn't read this SVG (%v). Re-export it from your design app as plain SVG.", err)
	}
	if len(doc.Shapes) == 0 {
		return nil, inputErr("This file has no closed shapes. Bushwhack needs outlines it can cut around, like paths or polygons.")
	}
	p := &pipeline{ctx: ctx, o: o, doc: doc}
	for _, w := range doc.Warnings {
		p.warn("%s", w)
	}
	p.budget = flatten.NewBudget(ctx, lim.MaxVertices)
	return p.run()
}

func validate(o *Options) error {
	switch {
	case o.BushingOD <= 0 || o.BushingOD > 4:
		return inputErr("Enter a bushing outside diameter between 0 and 4 in.")
	case o.BitDia <= 0:
		return inputErr("Enter a bit diameter bigger than 0.")
	case o.BitDia >= o.BushingOD:
		return inputErr("A %s bit won't fit through a %s bushing. Use a smaller bit or a bigger bushing.",
			inches(o.BitDia), inches(o.BushingOD))
	case o.MinBridge < 0 || o.MinBridge > 10:
		return inputErr("Enter a minimum bridge between 0 and 10 in.")
	case o.Scale < 0 || math.IsNaN(o.Scale) || math.IsInf(o.Scale, 0):
		return inputErr("Scale must be a positive number of inches per SVG unit.")
	}
	if o.Mode == "" {
		o.Mode = ModeHole
	}
	if o.Mode != ModeHole && o.Mode != ModePiece {
		return inputErr("mode must be %q or %q", ModeHole, ModePiece)
	}
	if o.Fit == "" {
		o.Fit = FitCanvas
	}
	if o.Fit != FitCanvas && o.Fit != FitArtwork && o.Fit != FitDocument {
		return inputErr("fit must be canvas, artwork or document")
	}
	if o.Scale == 0 && o.Fit != FitDocument && (o.Size <= 0 || o.Size > 120) {
		return inputErr("Enter a longest side between 0 and 120 in.")
	}
	return nil
}

type pipeline struct {
	ctx    context.Context
	o      Options
	doc    *svgin.Document
	budget *flatten.Budget
	notes  []Note
}

func (p *pipeline) add(n Note) { p.notes = append(p.notes, n) }

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return fmt.Sprintf(many, n)
}

// inches formats a length the way the UI copy does: 1/8 in, 0.055 in.
func inches(v float64) string {
	return strings.TrimSuffix(units.Fraction(v), `"`) + " in"
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// label names a shape for copy: "the left eye" from id "left-eye", else
// "shape 3".
func (s shape) label() string {
	n := strings.TrimSpace(strings.NewReplacer("-", " ", "_", " ").Replace(s.name))
	if n == "" || strings.HasPrefix(strings.ToLower(n), "path") || strings.HasPrefix(strings.ToLower(n), "rect") ||
		strings.HasPrefix(strings.ToLower(n), "circle") || strings.HasPrefix(strings.ToLower(n), "ellipse") ||
		strings.HasPrefix(strings.ToLower(n), "polygon") || strings.HasPrefix(strings.ToLower(n), "layer") {
		return fmt.Sprintf("shape %d", s.index)
	}
	return "the " + strings.ToLower(n)
}

func (p *pipeline) warn(format string, a ...any) {
	p.add(Note{Kind: NoteInput, Bridge: -1, Text: fmt.Sprintf(format, a...)})
}

func (p *pipeline) check() error {
	if err := p.ctx.Err(); err != nil {
		return inputErr("This drawing took too long to process. Simplify it and try again.")
	}
	return nil
}

type shape struct {
	index   int
	element string
	name    string
	orig    geom.Polys
	offset  geom.Polys
}

func (p *pipeline) run() (*Result, error) {
	o := p.o
	doc := p.doc

	scale, originX, originY, pageW, pageH, err := p.scale()
	if err != nil {
		return nil, err
	}

	// Flatten into inches.
	open := 0
	var shapes []shape
	for i, s := range doc.Shapes {
		var rings [][]geom.Pt
		for _, sp := range s.Subpaths {
			if !sp.Closed {
				last := sp.Start
				if n := len(sp.Segs); n > 0 {
					last = sp.Segs[n-1].P3
				}
				if last != sp.Start {
					open++
				}
			}
			pts, err := flatten.Subpath(mapSubpath(sp, scale, originX, originY), flattenTol, p.budget)
			if err != nil {
				return nil, p.limitErr(err)
			}
			ring := make([]geom.Pt, len(pts))
			for j, pt := range pts {
				ring[j] = geom.Pt{X: pt.X, Y: pt.Y}
			}
			rings = append(rings, ring)
		}
		polys := geom.FromRings(rings)
		if geom.Area(polys) < 1e-6 {
			p.warn("Skipped <%s> #%d because it encloses no area", s.Element, i+1)
			continue
		}
		shapes = append(shapes, shape{index: len(shapes) + 1, element: s.Element, name: s.Name, orig: polys})
	}
	if open > 0 {
		p.warn("%d outline(s) weren't closed, so Bushwhack closed them with a straight line", open)
	}
	if len(shapes) == 0 {
		return nil, inputErr("This file has no closed shapes. Bushwhack needs outlines it can cut around, like paths or polygons.")
	}
	if err := p.check(); err != nil {
		return nil, err
	}

	bushingR, bitR := o.BushingOD/2, o.BitDia/2
	offset := bushingR - bitR
	if o.Mode == ModePiece {
		offset = bushingR + bitR
	}

	// Offset each opening, then merge.
	var all, origAll geom.Polys
	outerIn := 0
	for i := range shapes {
		shapes[i].offset = geom.Offset(shapes[i].orig, offset)
		outerIn += geom.OuterCount(shapes[i].offset)
		all = append(all, shapes[i].offset...)
		origAll = append(origAll, shapes[i].orig...)
	}
	tmpl := geom.Union(all)
	origUnion := geom.Union(origAll)
	if err := p.check(); err != nil {
		return nil, err
	}
	// Bridges between distinct openings.
	bridges, minBridge, err := p.bridges(shapes)
	if err != nil {
		return nil, err
	}
	if n := geom.OuterCount(tmpl); n < outerIn {
		involved := map[int]bool{}
		for _, b := range bridges {
			if b.DistIn == 0 {
				involved[b.A], involved[b.B] = true, true
			}
		}
		var ids []int
		for id := range involved {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		groups := len(ids) - (outerIn - n)
		switch {
		case len(ids) >= 2 && groups == 1:
			p.add(Note{Kind: NoteMerged, Shapes: ids, Bridge: -1,
				Text: fmt.Sprintf("%d shapes merged into one opening. Space them farther apart or make the design bigger.", len(ids))})
		case len(ids) >= 2 && groups > 1:
			p.add(Note{Kind: NoteMerged, Shapes: ids, Bridge: -1,
				Text: fmt.Sprintf("%d shapes merged into %d openings. Space them farther apart or make the design bigger.", len(ids), groups)})
		default:
			p.add(Note{Kind: NoteMerged, Bridge: -1,
				Text: "Parts of a shape merged into one opening. Make the design bigger or use a smaller bushing."})
		}
	}
	if islands := len(tmpl) - geom.OuterCount(tmpl); islands > 0 {
		p.add(Note{Kind: NoteIsland, Bridge: -1, Text: fmt.Sprintf(
			"%s of the template float free, like the middle of an O. Mount the template on a backer board or add bridges to hold them.",
			plural(islands, "1 piece", "%d pieces"))})
	}

	// Simulate the router.
	center := geom.Offset(tmpl, -bushingR)
	var cut geom.Polys
	if o.Mode == ModeHole {
		cut = geom.Offset(center, bitR)
	} else {
		cut = geom.Offset(center, -bitR)
	}
	if err := p.check(); err != nil {
		return nil, err
	}

	reports := make([]ShapeReport, len(shapes))
	var lostAll geom.Polys
	for i, s := range shapes {
		area := geom.Area(s.orig)
		lostPolys := geom.Open(geom.Difference(s.orig, cut), sliver/2)
		lostAll = append(lostAll, lostPolys...)
		lost := geom.Area(lostPolys)
		extra := geom.Area(geom.Open(geom.Difference(geom.Intersect(cut, s.offset), origUnion), sliver/2))
		ob, tb := geom.Bounds(s.orig), geom.Bounds(s.offset)
		reports[i] = ShapeReport{
			Index: s.index, Element: s.element, Name: s.name,
			WidthIn: ob.W(), HeightIn: ob.H(),
			TemplateWidthIn: tb.W(), TemplateHeightIn: tb.H(),
			LostAreaPct: 100 * lost / area,
			OvercutPct:  100 * extra / area,
		}
		switch r := reports[i]; {
		case r.LostAreaPct >= 99.5:
			p.add(Note{Kind: NoteUncuttable, Shapes: []int{s.index}, Bridge: -1, Text: fmt.Sprintf(
				"%s is narrower than the %s bit, so it can't be cut at all.", capitalize(s.label()), inches(o.BitDia))})
		case r.LostAreaPct >= 5:
			p.add(Note{Kind: NoteLost, Shapes: []int{s.index}, Bridge: -1, Text: fmt.Sprintf(
				"Some detail in %s is too small for a %s bit and won't show up in the cut. It's marked on the preview.", s.label(), inches(o.BitDia))})
		}
		if reports[i].OvercutPct >= 5 {
			p.add(Note{Kind: NoteOvercut, Shapes: []int{s.index}, Bridge: -1, Text: fmt.Sprintf(
				"The bushing can't follow the narrow notches in %s, so the router will cut them away. Make the design bigger or widen the notches.", s.label())})
		}
		if i%16 == 15 {
			if err := p.check(); err != nil {
				return nil, err
			}
		}
	}

	// Page: the scaled canvas, grown only if an opening crosses its edge.
	tb := geom.Bounds(tmpl)
	const margin = 0.125
	minX, minY, maxX, maxY := 0.0, 0.0, pageW, pageH
	if tb.MinX < 0 || tb.MinY < 0 || tb.MaxX > pageW || tb.MaxY > pageH {
		minX, minY = math.Min(0, tb.MinX-margin), math.Min(0, tb.MinY-margin)
		maxX, maxY = math.Max(pageW, tb.MaxX+margin), math.Max(pageH, tb.MaxY+margin)
		p.add(Note{Kind: NotePage, Bridge: -1, Text: fmt.Sprintf(
			"The template openings run past the edge of your drawing, so the page grew from %.3f × %.3f in to %.3f × %.3f in.", pageW, pageH, maxX-minX, maxY-minY)})
	}
	dx, dy := -minX, -minY
	pageW, pageH = maxX-minX, maxY-minY
	shift := func(g geom.Polys) geom.Polys {
		if dx == 0 && dy == 0 {
			return g
		}
		return geom.Translate(g, dx, dy)
	}
	for i := range bridges {
		b := &bridges[i]
		b.P.X, b.P.Y, b.Q.X, b.Q.Y = b.P.X+dx, b.P.Y+dy, b.Q.X+dx, b.Q.Y+dy
	}

	for i := range reports {
		reports[i].Path = svgout.PathData(shift(shapes[i].orig))
	}
	layers := Layers{
		Drawing:  svgout.PathData(shift(origUnion)),
		Template: svgout.PathData(shift(tmpl)),
		Cut:      svgout.PathData(shift(cut)),
		Center:   svgout.PathData(shift(center)),
		Lost:     svgout.PathData(shift(lostAll)),
	}

	meta := svgout.Meta{
		BushingOD: o.BushingOD, BitDia: o.BitDia, Mode: o.Mode, Offset: offset,
		PageW: pageW, PageH: pageH,
	}
	marks := make([]svgout.Mark, 0, len(bridges))
	for _, b := range bridges {
		if b.DistIn < o.MinBridge {
			marks = append(marks, svgout.Mark{P: b.P, Q: b.Q})
		}
	}
	opens := openings(shapes, shift(tmpl), shift)
	res := &Result{
		Layers:      layers,
		TemplateSVG: svgout.Template(meta, opens),
		PreviewSVG: svgout.Preview(meta, svgout.Layers{
			Original: drawingPieces(shapes, shift),
			Template: prefixed("template-", opens),
			Cut:      cutPieces(opens, shift(cut)),
			Bridges:  marks,
			BushingR: bushingR,
		}),
		PageWidthIn:    pageW,
		PageHeightIn:   pageH,
		ScaleInPerUnit: scale,
		Offset:         offset,
		Shapes:         reports,
		MinBridgeIn:    minBridge,
		Bridges:        bridges,
	}
	// Bridge notes point at bridges by index; bridges are sorted thinnest first.
	res.Notes = p.notes
	res.Warnings = make([]string, len(p.notes))
	for i, n := range p.notes {
		res.Warnings[i] = n.Text
	}
	return res, nil
}

func (p *pipeline) limitErr(err error) error {
	if errors.Is(err, flatten.ErrTooManyVertices) {
		return inputErr("This drawing has too much detail to process. Simplify it and try again.")
	}
	if p.ctx.Err() != nil {
		return inputErr("This drawing took too long to process. Simplify it and try again.")
	}
	return err
}

// scale picks inches per user unit and the page in inches.
func (p *pipeline) scale() (scale, originX, originY, pageW, pageH float64, err error) {
	o, doc := p.o, p.doc
	fit := o.Fit
	if o.Scale == 0 && fit == FitCanvas && !doc.HasCanvas {
		p.warn("The SVG has no page size (viewBox or width/height), so its shapes were sized instead")
		fit = FitArtwork
	}
	if o.Scale == 0 && fit == FitDocument && (doc.PhysicalWidthIn == 0 || !doc.HasCanvas) {
		return 0, 0, 0, 0, 0, inputErr("This SVG doesn't save a real-world size (width and height in in, mm or cm). Turn off \"Use the size saved in the SVG\" and set the longest side instead.")
	}

	var art geom.Rect
	if fit == FitArtwork || !doc.HasCanvas {
		art, err = p.artworkBounds()
		if err != nil {
			return
		}
	}
	switch {
	case o.Scale > 0:
		scale = o.Scale
	case fit == FitDocument:
		scale = doc.PhysicalWidthIn / doc.Width
		if sy := doc.PhysicalHeightIn / doc.Height; math.Abs(sy-scale)/scale > 0.01 {
			p.warn("The SVG's width/height and viewBox have different proportions, so the width set the scale")
		}
	case fit == FitArtwork:
		longest := math.Max(art.W(), art.H())
		if longest <= 0 {
			return 0, 0, 0, 0, 0, inputErr("This file has no closed shapes. Bushwhack needs outlines it can cut around, like paths or polygons.")
		}
		scale = o.Size / longest
	default:
		scale = o.Size / math.Max(doc.Width, doc.Height)
	}
	if doc.HasCanvas {
		originX, originY = doc.MinX, doc.MinY
		pageW, pageH = doc.Width*scale, doc.Height*scale
	} else {
		originX, originY = art.MinX, art.MinY
		pageW, pageH = art.W()*scale, art.H()*scale
	}
	if pageW > 240 || pageH > 240 {
		return 0, 0, 0, 0, 0, inputErr("At this size the page would be %.0f × %.0f in. Check the size or scale.", pageW, pageH)
	}
	return
}

// artworkBounds measures the drawn shapes in user units with a coarse
// flatten that doesn't count against the vertex budget much.
func (p *pipeline) artworkBounds() (geom.Rect, error) {
	r := geom.EmptyRect()
	ctl := geom.EmptyRect()
	for _, s := range p.doc.Shapes {
		for _, sp := range s.Subpaths {
			ctl = ctl.Add(geom.Pt{X: sp.Start.X, Y: sp.Start.Y})
			for _, seg := range sp.Segs {
				ctl = ctl.Add(geom.Pt{X: seg.P3.X, Y: seg.P3.Y})
				if seg.Cubic {
					ctl = ctl.Add(geom.Pt{X: seg.P1.X, Y: seg.P1.Y}).Add(geom.Pt{X: seg.P2.X, Y: seg.P2.Y})
				}
			}
		}
	}
	if ctl.Empty() {
		return r, inputErr("This file has no closed shapes. Bushwhack needs outlines it can cut around, like paths or polygons.")
	}
	tol := math.Max(ctl.W(), ctl.H()) / 20000
	if tol == 0 {
		return r, inputErr("This file has no closed shapes. Bushwhack needs outlines it can cut around, like paths or polygons.")
	}
	b := flatten.NewBudget(p.ctx, 0)
	for _, s := range p.doc.Shapes {
		for _, sp := range s.Subpaths {
			pts, err := flatten.Subpath(sp, tol, b)
			if err != nil {
				return r, p.limitErr(err)
			}
			for _, pt := range pts {
				r = r.Add(geom.Pt{X: pt.X, Y: pt.Y})
			}
		}
	}
	return r, nil
}

func mapSubpath(sp pathdata.Subpath, s, ox, oy float64) pathdata.Subpath {
	m := func(pt pathdata.Point) pathdata.Point { return pathdata.Point{X: (pt.X - ox) * s, Y: (pt.Y - oy) * s} }
	out := pathdata.Subpath{Start: m(sp.Start), Closed: sp.Closed, Segs: make([]pathdata.Segment, len(sp.Segs))}
	for i, seg := range sp.Segs {
		out.Segs[i] = pathdata.Segment{Cubic: seg.Cubic, P1: m(seg.P1), P2: m(seg.P2), P3: m(seg.P3)}
	}
	return out
}

// bridges measures the gap between every pair of offset openings. Pairs that
// overlap are reported at distance 0.
func (p *pipeline) bridges(shapes []shape) ([]Bridge, float64, error) {
	var out []Bridge
	best := math.Inf(1)
	for i := 0; i < len(shapes); i++ {
		for j := i + 1; j < len(shapes); j++ {
			a, b := shapes[i].offset, shapes[j].offset
			limit := math.Max(p.o.MinBridge, best)
			gap := geom.BoundsGap(a, b)
			if gap >= limit {
				continue
			}
			var d float64
			var pa, pb geom.Pt
			if gap == 0 && geom.Area(geom.Intersect(a, b)) > 0 {
				ib := geom.Bounds(geom.Intersect(a, b))
				c := geom.Pt{X: (ib.MinX + ib.MaxX) / 2, Y: (ib.MinY + ib.MaxY) / 2}
				d, pa, pb = 0, c, c
			} else {
				var err error
				d, pa, pb, err = geom.Distance(p.ctx, a, b, limit)
				if err != nil {
					return nil, 0, inputErr("This drawing took too long to process. Simplify it and try again.")
				}
				if math.IsInf(d, 1) {
					continue
				}
			}
			best = math.Min(best, d)
			out = append(out, Bridge{A: shapes[i].index, B: shapes[j].index, DistIn: d, P: pa, Q: pb})
		}
		if err := p.check(); err != nil {
			return nil, 0, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DistIn < out[j].DistIn })
	// Keep only pairs that matter to the caller: thin ones plus the minimum.
	kept := out[:0]
	for i, b := range out {
		if b.DistIn < p.o.MinBridge || i == 0 {
			kept = append(kept, b)
		}
	}
	byIndex := map[int]shape{}
	for _, s := range shapes {
		byIndex[s.index] = s
	}
	thin := 0
	for i, b := range kept {
		if b.DistIn >= p.o.MinBridge || b.DistIn == 0 {
			continue // zero-width bridges are covered by the merged note
		}
		thin++
		if thin > 8 {
			continue
		}
		p.add(Note{Kind: NoteBridge, Shapes: []int{b.A, b.B}, Bridge: i, Text: fmt.Sprintf(
			"The wood between %s and %s is %.3f in thick. Your bushing can break through anything under %s. Make the design bigger or move those shapes apart.",
			byIndex[b.A].label(), byIndex[b.B].label(), b.DistIn, inches(p.o.MinBridge))})
	}
	if thin > 8 {
		p.add(Note{Kind: NoteBridge, Bridge: -1, Text: fmt.Sprintf("%d more thin bridges are marked on the preview.", thin-8)})
	}
	return kept, best, nil
}

// openings splits the merged template into one piece per opening and names
// each after the shapes inside it ("left-eye", "nose-and-mouth"), falling
// back to "opening-N". Pieces are ordered by their first shape, so the file
// lists them in the drawing's order.
func openings(shapes []shape, tmpl geom.Polys, shift func(geom.Polys) geom.Polys) []svgout.Opening {
	pieces := geom.Split(tmpl)
	members := make([][]shape, len(pieces))
	for _, s := range shapes {
		orig := geom.ToRings(shift(s.orig))
		if len(orig) == 0 || len(orig[0]) == 0 {
			continue
		}
		pt := orig[0][0]
		best, bestArea := -1, math.Inf(1)
		for i, pc := range pieces {
			// A shape's vertex is strictly inside the opening grown from it,
			// since every mode offsets outward by at least the bushing wall.
			if a := geom.Area(pc); a < bestArea && geom.Contains(pc, pt) {
				best, bestArea = i, a
			}
		}
		if best >= 0 {
			members[best] = append(members[best], s)
		}
	}
	order := make([]int, len(pieces))
	for i := range order {
		order[i] = i
	}
	first := func(i int) int {
		if len(members[i]) == 0 {
			return math.MaxInt
		}
		return members[i][0].index
	}
	sort.SliceStable(order, func(a, b int) bool { return first(order[a]) < first(order[b]) })

	used := map[string]bool{}
	out := make([]svgout.Opening, 0, len(pieces))
	for n, i := range order {
		// Name after its shapes when they're few and all named.
		var names []string
		for _, s := range members[i] {
			if id := xmlID(s.name); id != "" {
				names = append(names, id)
			}
		}
		id := fmt.Sprintf("opening-%d", n+1)
		if len(names) > 0 && len(names) == len(members[i]) && len(names) <= 3 {
			id = strings.Join(names, "-and-")
		}
		out = append(out, svgout.Opening{ID: unique(used, id), Polys: pieces[i]})
	}
	return out
}

// drawingPieces is the original drawing, one piece per shape, with ids like
// "drawing-left-eye" or "drawing-shape-3".
func drawingPieces(shapes []shape, shift func(geom.Polys) geom.Polys) []svgout.Opening {
	used := map[string]bool{}
	out := make([]svgout.Opening, 0, len(shapes))
	for _, s := range shapes {
		name := xmlID(s.name)
		if name == "" {
			name = fmt.Sprintf("shape-%d", s.index)
		}
		out = append(out, svgout.Opening{ID: unique(used, "drawing-"+name), Polys: shift(s.orig)})
	}
	return out
}

func prefixed(prefix string, in []svgout.Opening) []svgout.Opening {
	out := make([]svgout.Opening, len(in))
	for i, o := range in {
		out[i] = svgout.Opening{ID: prefix + o.ID, Polys: o.Polys}
	}
	return out
}

// cutPieces splits the simulated cut into its separate cut-outs and names
// each after the template opening it sits in ("cut-left-eye").
func cutPieces(opens []svgout.Opening, cut geom.Polys) []svgout.Opening {
	type piece struct {
		polys geom.Polys
		open  int // index into opens, or len(opens) if none
	}
	var pieces []piece
	for _, pc := range geom.Split(cut) {
		p := piece{pc, len(opens)}
		if rings := geom.ToRings(pc); len(rings) > 0 && len(rings[0]) > 0 {
			for i, o := range opens {
				if geom.Contains(o.Polys, rings[0][0]) {
					p.open = i
					break
				}
			}
		}
		pieces = append(pieces, p)
	}
	// Same order as the template, so the file reads in the drawing's order.
	sort.SliceStable(pieces, func(i, j int) bool { return pieces[i].open < pieces[j].open })
	used := map[string]bool{}
	out := make([]svgout.Opening, len(pieces))
	for n, p := range pieces {
		name := fmt.Sprintf("cut-%d", n+1)
		if p.open < len(opens) {
			name = "cut-" + opens[p.open].ID
		}
		out[n] = svgout.Opening{ID: unique(used, name), Polys: p.polys}
	}
	return out
}

func unique(used map[string]bool, id string) string {
	base := id
	for k := 2; used[id]; k++ {
		id = fmt.Sprintf("%s-%d", base, k)
	}
	used[id] = true
	return id
}

var genericID = regexp.MustCompile(`^(path|rect|circle|ellipse|polygon|polyline|layer|g)-?\d*$`)

// xmlID turns a shape name into a safe, readable id, or "" for names that
// carry no meaning (path123, rect7…).
func xmlID(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" || id[0] < 'a' || id[0] > 'z' || len(id) > 40 {
		return ""
	}
	if genericID.MatchString(id) {
		return ""
	}
	return id
}

func limitMessage(err error) string {
	if errors.Is(err, svgin.ErrUnsafe) {
		return "This SVG has a DOCTYPE or ENTITY declaration, which Bushwhack doesn't accept. Re-export it as plain SVG."
	}
	return "This drawing has too many elements to process (" + strings.TrimPrefix(err.Error(), "limit exceeded: ") + "). Simplify it and try again."
}
