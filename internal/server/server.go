// Package server is Bushwhack's HTTP layer: the embedded UI, the processing
// API and a health check. It is stateless; nothing is stored.
package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/brennanMKE/Bushwhack/internal/template"
	"github.com/brennanMKE/Bushwhack/internal/units"
	"github.com/brennanMKE/Bushwhack/patterns"
)

// Config tunes the server.
type Config struct {
	Static         fs.FS // built SPA: index.html plus assets
	Patterns       *patterns.Library
	Now            func() time.Time // for "upcoming" occasion; default time.Now
	MaxUpload      int64            // bytes
	RequestTimeout time.Duration    // per /api/process call
	MaxConcurrent  int              // simultaneous /api/process calls
	Version        string
	Log            *slog.Logger
}

type server struct {
	cfg   Config
	sem   chan struct{}
	index []byte // SPA shell
	csp   string
}

// New returns the HTTP handler.
func New(cfg Config) http.Handler {
	if cfg.MaxUpload == 0 {
		cfg.MaxUpload = 2 << 20
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = 4
	}
	if cfg.Log == nil {
		cfg.Log = slog.Default()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	s := &server{cfg: cfg, sem: make(chan struct{}, cfg.MaxConcurrent)}
	s.index, _ = fs.ReadFile(cfg.Static, "index.html")
	s.csp = contentSecurityPolicy(s.index, styleAttrs(cfg.Static))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/process", s.process)
	mux.HandleFunc("GET /api/patterns", s.patternList)
	mux.HandleFunc("GET /api/patterns/{slug}", s.patternSVG)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, "ok "+cfg.Version+"\n")
	})
	mux.Handle("GET /", s.static())
	return s.headers(mux)
}

var inlineScript = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

var styleAttr = regexp.MustCompile(`style="([^"]*)"`)

// styleAttrs finds style="…" literals in the built JS (SvelteKit's route
// announcer has one) so the CSP can allow exactly those, by hash.
func styleAttrs(fsys fs.FS) [][]byte {
	var out [][]byte
	if fsys == nil {
		return nil
	}
	seen := map[string]bool{}
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".js" {
			return nil
		}
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil
		}
		for _, m := range styleAttr.FindAllSubmatch(b, -1) {
			if !seen[string(m[1])] {
				seen[string(m[1])] = true
				out = append(out, m[1])
			}
		}
		return nil
	})
	return out
}

func hash(b []byte) string {
	sum := sha256.Sum256(b)
	return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
}

// contentSecurityPolicy allows the SPA shell's inline bootstrap script and
// the app's few static style attributes by hash, so neither script-src nor
// style-src needs 'unsafe-inline'.
func contentSecurityPolicy(index []byte, styles [][]byte) string {
	scripts := "'self'"
	for _, m := range inlineScript.FindAllSubmatch(index, -1) {
		scripts += " " + hash(m[1])
	}
	style := "'self'"
	if len(styles) > 0 {
		style += " 'unsafe-hashes'"
		for _, s := range styles {
			style += " " + hash(s)
		}
	}
	return "default-src 'self'; img-src 'self' data: blob:; style-src " + style + "; font-src 'self'; script-src " + scripts +
		"; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'"
}

func (s *server) headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", s.csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// static serves built assets, and the SPA shell for every other path so
// client routes like /make and /patterns/bat load directly.
func (s *server) static() http.Handler {
	files := http.FileServerFS(s.cfg.Static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p != "" && p != "index.html" {
			if st, err := fs.Stat(s.cfg.Static, p); err == nil && !st.IsDir() {
				if strings.HasPrefix(p, "_app/immutable/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "public, max-age=3600")
				}
				files.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(p, "api/") || path.Ext(p) != "" {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(s.index))
	})
}

type occasionJSON struct {
	patterns.Occasion
	Upcoming bool               `json:"upcoming"`
	Patterns []patterns.Pattern `json:"patterns"`
}

func (s *server) patternList(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Patterns == nil {
		writeJSON(w, http.StatusOK, map[string]any{"occasions": []any{}})
		return
	}
	now := s.cfg.Now()
	up := patterns.Upcoming(now)
	all := s.cfg.Patterns.All()
	var out []occasionJSON
	for _, o := range patterns.Ordered(now) {
		oj := occasionJSON{Occasion: o, Upcoming: o.Slug == up.Slug, Patterns: []patterns.Pattern{}}
		for _, p := range all {
			if p.Occasion == o.Slug {
				oj.Patterns = append(oj.Patterns, p)
			}
		}
		out = append(out, oj)
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, map[string]any{"occasions": out})
}

// patternSVG serves /api/patterns/{slug}.svg as a download.
func (s *server) patternSVG(w http.ResponseWriter, r *http.Request) {
	slug, ok := strings.CutSuffix(r.PathValue("slug"), ".svg")
	if !ok || s.cfg.Patterns == nil {
		http.NotFound(w, r)
		return
	}
	p, ok := s.cfg.Patterns.Get(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "image/svg+xml")
	h.Set("Content-Disposition", `attachment; filename="`+p.Slug+`.svg"`)
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	h.Set("Cache-Control", "public, max-age=3600")
	w.Write(p.SVG)
}

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiError{msg})
}

type shapeJSON struct {
	Index            int     `json:"index"`
	Element          string  `json:"element"`
	Name             string  `json:"name"`
	Path             string  `json:"path"`
	WidthIn          float64 `json:"widthIn"`
	HeightIn         float64 `json:"heightIn"`
	TemplateWidthIn  float64 `json:"templateWidthIn"`
	TemplateHeightIn float64 `json:"templateHeightIn"`
	LostAreaPct      float64 `json:"lostAreaPct"`
	OvercutPct       float64 `json:"overcutPct"`
}

type pointJSON struct {
	XIn float64 `json:"xIn"`
	YIn float64 `json:"yIn"`
}

type bridgeJSON struct {
	A      int       `json:"a"`
	B      int       `json:"b"`
	DistIn float64   `json:"distIn"`
	At     pointJSON `json:"at"` // midpoint of the narrowest gap
	P      pointJSON `json:"p"`
	Q      pointJSON `json:"q"`
}

type pageJSON struct {
	WidthIn  float64 `json:"widthIn"`
	HeightIn float64 `json:"heightIn"`
}

type optionsJSON struct {
	BushingOD float64 `json:"bushingOD"`
	BitDia    float64 `json:"bitDia"`
	Mode      string  `json:"mode"`
	Fit       string  `json:"fit"`
	Size      float64 `json:"size"`
	Scale     float64 `json:"scale,omitempty"`
	MinBridge float64 `json:"minBridge"`
}

type response struct {
	TemplateSVG    string          `json:"templateSvg"`
	PreviewSVG     string          `json:"previewSvg"`
	Page           pageJSON        `json:"page"`
	ScaleInPerUnit float64         `json:"scaleInPerUnit"`
	OffsetIn       float64         `json:"offsetIn"`
	Shapes         []shapeJSON     `json:"shapes"`
	MinBridgeIn    *float64        `json:"minBridgeIn"` // null with fewer than two openings
	MinBridgeAt    *pointJSON      `json:"minBridgeAt"`
	Bridges        []bridgeJSON    `json:"bridges"`
	Layers         template.Layers `json:"layers"`
	LostRegionsSVG string          `json:"lostRegionsSvg"`
	Pattern        string          `json:"pattern,omitempty"`
	Warnings       []string        `json:"warnings"`
	Notes          []template.Note `json:"notes"`
	Options        optionsJSON     `json:"options"`
}

func round(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}

// ParseOptions reads and validates form fields. Exposed for tests.
func ParseOptions(get func(string) string) (template.Options, error) {
	o := template.DefaultOptions()
	length := func(field, label string, dst *float64) error {
		v := strings.TrimSpace(get(field))
		if v == "" {
			return nil
		}
		in, err := units.ParseLength(v)
		if err != nil {
			return errors.New(label + ": " + err.Error())
		}
		*dst = in
		return nil
	}
	if err := length("bushing", "Bushing OD", &o.BushingOD); err != nil {
		return o, err
	}
	if err := length("bit", "Bit diameter", &o.BitDia); err != nil {
		return o, err
	}
	if err := length("size", "Size", &o.Size); err != nil {
		return o, err
	}
	if err := length("minBridge", "Minimum bridge", &o.MinBridge); err != nil {
		return o, err
	}
	if v := strings.TrimSpace(get("mode")); v != "" {
		o.Mode = v
	}
	if v := strings.TrimSpace(get("fit")); v != "" {
		o.Fit = v
	}
	if v := strings.TrimSpace(get("scale")); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 || math.IsInf(f, 0) || math.IsNaN(f) {
			return o, errors.New("Scale must be a positive number of inches per SVG unit")
		}
		o.Scale = f
	}
	return o, nil
}

func (s *server) process(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxUpload+64<<10)
	if err := r.ParseMultipartForm(s.cfg.MaxUpload + 64<<10); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			fail(w, http.StatusRequestEntityTooLarge, "The file is too large (max 2 MB).")
			return
		}
		fail(w, http.StatusBadRequest, "Expected a multipart form with an SVG file.")
		return
	}
	defer r.MultipartForm.RemoveAll()
	opts, err := ParseOptions(r.FormValue)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	var data []byte
	slug := strings.TrimSpace(r.FormValue("pattern"))
	if slug != "" {
		var p patterns.Pattern
		ok := false
		if s.cfg.Patterns != nil {
			p, ok = s.cfg.Patterns.Get(slug)
		}
		if !ok {
			fail(w, http.StatusNotFound, "That pattern doesn't exist.")
			return
		}
		data = p.SVG
		if strings.TrimSpace(r.FormValue("size")) == "" {
			opts.Size = p.RecommendedSizeIn
		}
		if strings.TrimSpace(r.FormValue("fit")) == "" {
			opts.Fit = p.Fit
		}
	} else {
		file, hdr, err := r.FormFile("file")
		if err != nil {
			fail(w, http.StatusBadRequest, "Choose an SVG file to upload.")
			return
		}
		defer file.Close()
		if hdr.Size > s.cfg.MaxUpload {
			fail(w, http.StatusRequestEntityTooLarge, "The file is too large (max 2 MB).")
			return
		}
		data, err = io.ReadAll(io.LimitReader(file, s.cfg.MaxUpload+1))
		if err != nil || int64(len(data)) > s.cfg.MaxUpload {
			fail(w, http.StatusRequestEntityTooLarge, "The file is too large (max 2 MB).")
			return
		}
	}

	select {
	case s.sem <- struct{}{}:
	case <-time.After(3 * time.Second):
		fail(w, http.StatusServiceUnavailable, "Bushwhack is busy. Try again in a moment.")
		return
	case <-r.Context().Done():
		return
	}

	// Process checks ctx between steps, but a single Clipper call can't be
	// interrupted. Answer at the deadline regardless; the semaphore slot is
	// released only when the work really ends, so overruns can't pile up.
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.RequestTimeout)
	type outcome struct {
		res *template.Result
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		defer func() { <-s.sem }()
		defer cancel()
		res, err := template.Process(ctx, data, opts)
		done <- outcome{res, err}
	}()
	var res *template.Result
	select {
	case o := <-done:
		res, err = o.res, o.err
	case <-time.After(s.cfg.RequestTimeout + time.Second):
		s.cfg.Log.Warn("process overran its deadline", "bytes", len(data))
		fail(w, http.StatusUnprocessableEntity, "This drawing took too long to process. Simplify it and try again.")
		return
	case <-r.Context().Done():
		cancel()
		return
	}
	if err != nil {
		var ie *template.InputError
		if errors.As(err, &ie) {
			s.cfg.Log.Info("process rejected", "bytes", len(data), "err", ie.Msg, "ms", time.Since(start).Milliseconds())
			fail(w, http.StatusUnprocessableEntity, ie.Msg)
			return
		}
		s.cfg.Log.Error("process failed", "bytes", len(data), "err", err)
		fail(w, http.StatusInternalServerError, "Something went wrong processing that file.")
		return
	}

	out := response{
		TemplateSVG:    string(res.TemplateSVG),
		PreviewSVG:     string(res.PreviewSVG),
		Page:           pageJSON{round(res.PageWidthIn, 4), round(res.PageHeightIn, 4)},
		ScaleInPerUnit: res.ScaleInPerUnit,
		OffsetIn:       res.Offset,
		Shapes:         make([]shapeJSON, len(res.Shapes)),
		Bridges:        make([]bridgeJSON, len(res.Bridges)),
		Warnings:       res.Warnings,
		Notes:          res.Notes,
		Layers:         res.Layers,
		LostRegionsSVG: res.Layers.Lost,
		Pattern:        slug,
		Options: optionsJSON{
			BushingOD: opts.BushingOD, BitDia: opts.BitDia, Mode: opts.Mode, Fit: opts.Fit,
			Size: opts.Size, Scale: opts.Scale, MinBridge: opts.MinBridge,
		},
	}
	if out.Warnings == nil {
		out.Warnings = []string{}
	}
	if out.Notes == nil {
		out.Notes = []template.Note{}
	}
	if !math.IsInf(res.MinBridgeIn, 1) {
		v := round(res.MinBridgeIn, 4)
		out.MinBridgeIn = &v
	}
	for i, sh := range res.Shapes {
		out.Shapes[i] = shapeJSON{
			Index: sh.Index, Element: sh.Element, Name: sh.Name, Path: sh.Path,
			WidthIn: round(sh.WidthIn, 4), HeightIn: round(sh.HeightIn, 4),
			TemplateWidthIn: round(sh.TemplateWidthIn, 4), TemplateHeightIn: round(sh.TemplateHeightIn, 4),
			LostAreaPct: round(sh.LostAreaPct, 2), OvercutPct: round(sh.OvercutPct, 2),
		}
	}
	pt := func(x, y float64) pointJSON { return pointJSON{round(x, 4), round(y, 4)} }
	for i, b := range res.Bridges {
		out.Bridges[i] = bridgeJSON{
			A: b.A, B: b.B, DistIn: round(b.DistIn, 4),
			At: pt((b.P.X+b.Q.X)/2, (b.P.Y+b.Q.Y)/2), P: pt(b.P.X, b.P.Y), Q: pt(b.Q.X, b.Q.Y),
		}
	}
	if len(out.Bridges) > 0 && out.MinBridgeIn != nil {
		at := out.Bridges[0].At
		out.MinBridgeAt = &at
	}
	s.cfg.Log.Info("processed", "bytes", len(data), "shapes", len(res.Shapes), "ms", time.Since(start).Milliseconds())
	writeJSON(w, http.StatusOK, out)
}
