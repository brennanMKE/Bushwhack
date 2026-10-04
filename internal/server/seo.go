package server

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"html"
	"io"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/brennanMKE/Bushwhack/patterns"
)

// DefaultSiteURL is the public origin used for canonical links and the sitemap.
const DefaultSiteURL = "https://bushwhack.sstools.co"

// loadPages reads the prerendered pages (index.html, guide.html…) keyed by
// route ("/", "/guide"), and the SPA shell that every other route loads.
func loadPages(fsys fs.FS) (pages map[string][]byte, shell []byte) {
	pages = map[string][]byte{}
	if fsys == nil {
		return pages, nil
	}
	names, _ := fs.Glob(fsys, "*.html")
	for _, name := range names {
		b, err := fs.ReadFile(fsys, name)
		if err != nil {
			continue
		}
		if name == "spa.html" {
			shell = b
			continue
		}
		pages[pageRoute(name)] = b
	}
	if shell == nil {
		// An SPA-only build has just index.html.
		shell = pages["/"]
	}
	return pages, shell
}

func pageRoute(name string) string {
	r := "/" + strings.TrimSuffix(name, ".html")
	if r == "/index" {
		return "/"
	}
	return r
}

// head is the per-route metadata injected into the SPA shell, so routes
// that aren't prerendered still have a title, description and canonical URL
// before any script runs.
type head struct {
	title, description, path string
	noindex                  bool
}

// shellHead describes a client-rendered route, or reports false when the
// route doesn't exist.
func (s *server) shellHead(route string) (head, bool) {
	if route == "/make" {
		return head{
			title:       "Make a router template · Bushwhack",
			description: "Upload an SVG, enter your guide bushing and bit, and download a true-size router template with every opening offset so the cut lands on your line.",
			path:        "/make",
		}, true
	}
	if slug, ok := strings.CutPrefix(route, "/patterns/"); ok && s.cfg.Patterns != nil {
		if p, ok := s.cfg.Patterns.Get(slug); ok {
			occ := ""
			if o, ok := patterns.Find(p.Occasion); ok {
				occ = " for " + o.Name
			}
			return head{
				title:       p.Name + " pattern · Bushwhack",
				description: "Free " + p.Name + " router template pattern" + occ + ". " + p.Note,
				path:        route,
			}, true
		}
	}
	return head{title: "Not found · Bushwhack", noindex: true}, false
}

func (s *server) injectHead(shell []byte, h head) []byte {
	var b strings.Builder
	e := html.EscapeString
	b.WriteString("<title>" + e(h.title) + "</title>\n")
	if h.noindex {
		b.WriteString(`<meta name="robots" content="noindex" />` + "\n")
	} else {
		url := s.cfg.SiteURL + h.path
		b.WriteString(`<meta name="description" content="` + e(h.description) + `" />` + "\n")
		b.WriteString(`<link rel="canonical" href="` + e(url) + `" />` + "\n")
		b.WriteString(`<meta property="og:title" content="` + e(h.title) + `" />` + "\n")
		b.WriteString(`<meta property="og:description" content="` + e(h.description) + `" />` + "\n")
		b.WriteString(`<meta property="og:url" content="` + e(url) + `" />` + "\n")
		b.WriteString(`<meta name="twitter:title" content="` + e(h.title) + `" />` + "\n")
		b.WriteString(`<meta name="twitter:description" content="` + e(h.description) + `" />` + "\n")
	}
	i := bytes.Index(shell, []byte("</head>"))
	if i < 0 {
		return shell
	}
	out := make([]byte, 0, len(shell)+b.Len())
	out = append(out, shell[:i]...)
	out = append(out, b.String()...)
	return append(out, shell[i:]...)
}

// sitemap lists the prerendered pages, the workbench and every pattern page.
func (s *server) sitemap(w http.ResponseWriter, r *http.Request) {
	routes := make([]string, 0, len(s.pages)+1)
	for route := range s.pages {
		routes = append(routes, route)
	}
	routes = append(routes, "/make")
	sort.Strings(routes)
	if s.cfg.Patterns != nil {
		for _, p := range s.cfg.Patterns.All() {
			routes = append(routes, "/patterns/"+p.Slug)
		}
	}
	type url struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	}
	set := struct {
		XMLName xml.Name `xml:"urlset"`
		NS      string   `xml:"xmlns,attr"`
		URLs    []url    `xml:"url"`
	}{NS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, route := range routes {
		set.URLs = append(set.URLs, url{Loc: s.cfg.SiteURL + route, LastMod: s.cfg.Updated})
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	io.WriteString(w, xml.Header)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	enc.Encode(set)
}

// canonicalRedirect sends /guide/ and /guide.html to /guide, and
// /index.html to /, keeping the query. It reports whether it redirected.
func (s *server) canonicalRedirect(w http.ResponseWriter, r *http.Request) bool {
	p := r.URL.Path
	target := ""
	switch {
	case p != "/" && strings.HasSuffix(p, "/"):
		target = strings.TrimRight(p, "/")
		if target == "" {
			target = "/"
		}
	case path.Ext(p) == ".html" && path.Dir(p) == "/":
		if _, ok := s.pages[pageRoute(path.Base(p))]; ok || path.Base(p) == "spa.html" {
			target = pageRoute(path.Base(p))
			if path.Base(p) == "spa.html" {
				target = "/"
			}
		}
	}
	if target == "" {
		return false
	}
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
	return true
}

// gzipped compresses text responses for clients that accept gzip.
func gzipped(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead || r.Header.Get("Range") != "" ||
			!strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gw := &gzipWriter{ResponseWriter: w}
		defer gw.close()
		next.ServeHTTP(gw, r)
	})
}

var gzPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

type gzipWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	wroteHeader bool
}

func compressible(contentType string) bool {
	ct, _, _ := strings.Cut(contentType, ";")
	switch strings.TrimSpace(ct) {
	case "text/html", "text/css", "text/plain", "text/javascript", "application/javascript",
		"application/json", "application/xml", "image/svg+xml":
		return true
	}
	return false
}

func (g *gzipWriter) WriteHeader(code int) {
	if g.wroteHeader {
		return
	}
	g.wroteHeader = true
	h := g.Header()
	h.Add("Vary", "Accept-Encoding")
	if code != http.StatusNoContent && code != http.StatusNotModified &&
		h.Get("Content-Encoding") == "" && compressible(h.Get("Content-Type")) {
		h.Del("Content-Length")
		h.Set("Content-Encoding", "gzip")
		g.gz = gzPool.Get().(*gzip.Writer)
		g.gz.Reset(g.ResponseWriter)
	}
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	if !g.wroteHeader {
		if g.Header().Get("Content-Type") == "" {
			g.Header().Set("Content-Type", http.DetectContentType(b))
		}
		g.WriteHeader(http.StatusOK)
	}
	if g.gz != nil {
		return g.gz.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipWriter) close() {
	if g.gz != nil {
		g.gz.Close()
		gzPool.Put(g.gz)
	}
}
