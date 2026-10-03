// Package patterns is Bushwhack's embedded starter pattern library: original
// CC0 SVGs grouped by occasion, with metadata and computed safe sizes.
package patterns

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/brennanMKE/Bushwhack/internal/template"
)

//go:embed */*.svg */*.json
var files embed.FS

// Meta is {slug}.json.
type Meta struct {
	Slug              string  `json:"slug"`
	Name              string  `json:"name"`
	Occasion          string  `json:"occasion"`
	RecommendedSizeIn float64 `json:"recommendedSizeIn"`
	Fit               string  `json:"fit"`
	Note              string  `json:"note"`
	License           string  `json:"license"`
}

// Pattern is a loaded pattern.
type Pattern struct {
	Meta
	SVG []byte `json:"-"`
	// Drawing outline at the recommended size, for thumbnails.
	Drawing       string   `json:"drawing"`
	WidthIn       float64  `json:"widthIn"`
	HeightIn      float64  `json:"heightIn"`
	MinSafeSizeIn *float64 `json:"minSafeSizeIn"` // nil until computed
}

// Library holds every pattern.
type Library struct {
	mu       sync.RWMutex
	patterns []*Pattern
	bySlug   map[string]*Pattern
}

var slugRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Load reads and validates the embedded library. It fails on any bad pattern.
func Load() (*Library, error) { return load(files) }

func load(fsys fs.FS) (*Library, error) {
	lib := &Library{bySlug: map[string]*Pattern{}}
	metas, err := fs.Glob(fsys, "*/*.json")
	if err != nil {
		return nil, err
	}
	for _, mp := range metas {
		raw, err := fs.ReadFile(fsys, mp)
		if err != nil {
			return nil, err
		}
		var m Meta
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("%s: %w", mp, err)
		}
		dir, base := path.Split(mp)
		dir = strings.TrimSuffix(dir, "/")
		switch {
		case !slugRE.MatchString(m.Slug) || m.Slug+".json" != base:
			return nil, fmt.Errorf("%s: slug %q must match the file name", mp, m.Slug)
		case m.Name == "":
			return nil, fmt.Errorf("%s: missing name", mp)
		case m.Occasion != dir:
			return nil, fmt.Errorf("%s: occasion %q must match its directory", mp, m.Occasion)
		case m.RecommendedSizeIn <= 0 || m.RecommendedSizeIn > 48:
			return nil, fmt.Errorf("%s: bad recommendedSizeIn", mp)
		case m.Fit != template.FitArtwork && m.Fit != template.FitCanvas:
			return nil, fmt.Errorf("%s: fit must be artwork or canvas", mp)
		case m.License != "CC0-1.0":
			return nil, fmt.Errorf("%s: license must be CC0-1.0", mp)
		}
		if _, ok := Find(m.Occasion); !ok {
			return nil, fmt.Errorf("%s: unknown occasion %q", mp, m.Occasion)
		}
		if _, dup := lib.bySlug[m.Slug]; dup {
			return nil, fmt.Errorf("%s: duplicate slug %q", mp, m.Slug)
		}
		svg, err := fs.ReadFile(fsys, path.Join(dir, m.Slug+".svg"))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", mp, err)
		}
		p := &Pattern{Meta: m, SVG: svg}
		res, err := template.Process(context.Background(), svg, p.Options(template.DefaultOptions()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", mp, err)
		}
		p.Drawing = res.Layers.Drawing
		p.WidthIn, p.HeightIn = res.PageWidthIn, res.PageHeightIn
		lib.patterns = append(lib.patterns, p)
		lib.bySlug[m.Slug] = p
	}
	sort.Slice(lib.patterns, func(i, j int) bool {
		a, b := lib.patterns[i], lib.patterns[j]
		if a.Occasion != b.Occasion {
			return occasionIndex(a.Occasion) < occasionIndex(b.Occasion)
		}
		return a.Name < b.Name
	})
	return lib, nil
}

func occasionIndex(slug string) int {
	for i, o := range Occasions {
		if o.Slug == slug {
			return i
		}
	}
	return len(Occasions)
}

// Options returns base with this pattern's fit and recommended size applied.
func (p *Pattern) Options(base template.Options) template.Options {
	base.Fit = p.Fit
	base.Size = p.RecommendedSizeIn
	base.Scale = 0
	return base
}

// All returns every pattern in occasion order.
func (l *Library) All() []Pattern {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Pattern, len(l.patterns))
	for i, p := range l.patterns {
		out[i] = *p
	}
	return out
}

// Get returns one pattern.
func (l *Library) Get(slug string) (Pattern, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	p, ok := l.bySlug[slug]
	if !ok {
		return Pattern{}, false
	}
	return *p, true
}

// Check is the pattern design rule: at size with the default 5/16 bushing and
// 1/8 bit, no merged openings, every bridge at least 1/4 in, and under 3%
// lost area per shape. It returns the first rule broken, or "".
func Check(res *template.Result, minBridge float64) string {
	for _, n := range res.Notes {
		if n.Kind == template.NoteMerged {
			return n.Text
		}
	}
	if !math.IsInf(res.MinBridgeIn, 1) && res.MinBridgeIn < minBridge-1e-9 {
		return fmt.Sprintf("thinnest bridge %.3f in is under %.3f in", res.MinBridgeIn, minBridge)
	}
	for _, s := range res.Shapes {
		if s.LostAreaPct >= 3 {
			return fmt.Sprintf("shape %d loses %.1f%% of its area", s.Index, s.LostAreaPct)
		}
	}
	return ""
}

// MinSafeSize finds, to 1/16 in, the smallest longest-side size at which p
// still passes Check with the default bushing and bit.
func MinSafeSize(ctx context.Context, p *Pattern) (float64, error) {
	opts := p.Options(template.DefaultOptions())
	ok := func(size float64) (bool, error) {
		o := opts
		o.Size = size
		res, err := template.Process(ctx, p.SVG, o)
		if err != nil {
			return false, err
		}
		return Check(res, o.MinBridge) == "", nil
	}
	hi := p.RecommendedSizeIn
	if good, err := ok(hi); err != nil || !good {
		return 0, fmt.Errorf("%s fails at its recommended size", p.Slug)
	}
	lo := 0.5
	for hi-lo > 1.0/16 {
		mid := (lo + hi) / 2
		good, err := ok(mid)
		if err != nil {
			return 0, err
		}
		if good {
			hi = mid
		} else {
			lo = mid
		}
	}
	return math.Ceil(hi*16) / 16, nil
}

// ComputeSafeSizes fills MinSafeSizeIn for every pattern. Run it in the
// background at startup; readers see nil until each value lands.
func (l *Library) ComputeSafeSizes(ctx context.Context) {
	l.mu.RLock()
	ps := append([]*Pattern(nil), l.patterns...)
	l.mu.RUnlock()
	for _, p := range ps {
		v, err := MinSafeSize(ctx, p)
		if err != nil {
			continue
		}
		l.mu.Lock()
		p.MinSafeSizeIn = &v
		l.mu.Unlock()
	}
}
