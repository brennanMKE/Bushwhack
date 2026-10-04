package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/brennanMKE/Bushwhack/patterns"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	lib, err := patterns.Load()
	if err != nil {
		t.Fatal(err)
	}
	static := fstest.MapFS{
		"index.html":          {Data: []byte(`<html><head></head><h1>Home</h1><script>home()</script></html>`)},
		"guide.html":          {Data: []byte(`<html><head></head><h1>Guide</h1><script>guide()</script></html>`)},
		"spa.html":            {Data: []byte(`<html><head></head><script>boot()</script></html>`)},
		"robots.txt":          {Data: []byte("User-agent: *\n")},
		"_app/immutable/a.js": {Data: []byte(`x='<div style="position: absolute">'`)},
		"fonts/f.woff2":       {Data: []byte("font")},
	}
	return New(Config{Static: static, Patterns: lib})
}

func post(t *testing.T, h http.Handler, fields map[string]string, file string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		w.WriteField(k, v)
	}
	if file != "" {
		fw, _ := w.CreateFormFile("file", "x.svg")
		fw.Write([]byte(file))
	}
	w.Close()
	req := httptest.NewRequest("POST", "/api/process", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestProcessFile(t *testing.T) {
	h := newTestServer(t)
	rec := post(t, h, map[string]string{"bushing": "5/16", "bit": "1/8", "size": "6"},
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"><rect id="a" x="1" y="1" width="2" height="2"/></svg>`)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var out response
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.OffsetIn != 0.09375 || len(out.Shapes) != 1 || out.Shapes[0].Name != "a" || out.Layers.Template == "" || out.MinBridgeIn != nil {
		t.Errorf("unexpected response %+v", out)
	}
}

func TestProcessPattern(t *testing.T) {
	rec := post(t, newTestServer(t), map[string]string{"pattern": "classic-jack"}, "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"pattern":"classic-jack"`) {
		t.Fatalf("status %d: %.200s", rec.Code, rec.Body)
	}
	rec = post(t, newTestServer(t), map[string]string{"pattern": "nope"}, "")
	if rec.Code != 404 {
		t.Errorf("unknown pattern status %d", rec.Code)
	}
}

func TestProcessErrors(t *testing.T) {
	h := newTestServer(t)
	cases := []struct {
		fields map[string]string
		file   string
		code   int
	}{
		{map[string]string{"bushing": "abc"}, `<svg/>`, 400},
		{nil, "", 400},
		{map[string]string{"bit": "1/2"}, `<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`, 422},
		{nil, `<svg xmlns="http://www.w3.org/2000/svg"><text>hi</text></svg>`, 422},
		{nil, `not xml`, 422},
		{nil, strings.Repeat("x", 3<<20), 413},
	}
	for i, c := range cases {
		if rec := post(t, h, c.fields, c.file); rec.Code != c.code {
			t.Errorf("case %d: status %d, want %d: %.200s", i, rec.Code, c.code, rec.Body)
		}
	}
}

func TestSPAAndHeaders(t *testing.T) {
	h := newTestServer(t)
	for path, want := range map[string]int{"/": 200, "/make": 200, "/patterns/bat": 200, "/fonts/f.woff2": 200, "/missing.js": 404, "/api/nope": 404, "/healthz": 200} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != want {
			t.Errorf("%s: %d, want %d", path, rec.Code, want)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/make", nil))
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self' 'sha256-") || !strings.Contains(csp, "'unsafe-hashes' 'sha256-") || strings.Contains(csp, "unsafe-inline") {
		t.Errorf("csp %q", csp)
	}
	if !strings.Contains(rec.Body.String(), "boot()") {
		t.Error("SPA shell not served for client route")
	}
	for _, src := range []string{"home()", "guide()", "boot()"} {
		if !strings.Contains(csp, hash([]byte(src))) {
			t.Errorf("csp missing hash for %s", src)
		}
	}
}

func get(h http.Handler, path string, hdr ...string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", path, nil)
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	h.ServeHTTP(rec, req)
	return rec
}

func TestPagesAndRedirects(t *testing.T) {
	h := newTestServer(t)
	if b := get(h, "/").Body.String(); !strings.Contains(b, "<h1>Home</h1>") {
		t.Errorf("/ served %q", b)
	}
	if b := get(h, "/guide").Body.String(); !strings.Contains(b, "<h1>Guide</h1>") {
		t.Errorf("/guide served %q", b)
	}
	for from, to := range map[string]string{
		"/guide/":        "/guide",
		"/guide.html":    "/guide",
		"/index.html":    "/",
		"/make/?bit=1/8": "/make?bit=1/8",
		"/patterns/bat/": "/patterns/bat",
	} {
		rec := get(h, from)
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != to {
			t.Errorf("%s: %d to %q, want 301 to %q", from, rec.Code, rec.Header().Get("Location"), to)
		}
	}
}

func TestShellHead(t *testing.T) {
	h := newTestServer(t)
	b := get(h, "/patterns/bat").Body.String()
	for _, want := range []string{"<title>Bat pattern · Bushwhack</title>", `<link rel="canonical" href="https://bushwhack.sstools.co/patterns/bat" />`, "for Halloween"} {
		if !strings.Contains(b, want) {
			t.Errorf("/patterns/bat missing %q", want)
		}
	}
	if b := get(h, "/make?bushing=3/8").Body.String(); !strings.Contains(b, `href="https://bushwhack.sstools.co/make"`) {
		t.Error("/make canonical should drop the query")
	}
	for _, p := range []string{"/nope", "/patterns/nope"} {
		rec := get(h, p)
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "noindex") || !strings.Contains(rec.Body.String(), "boot()") {
			t.Errorf("%s: %d, want 404 shell with noindex", p, rec.Code)
		}
	}
}

func TestSitemapAndRobots(t *testing.T) {
	lib, _ := patterns.Load()
	h := New(Config{Static: fstest.MapFS{"index.html": {Data: []byte("<html></html>")}, "guide.html": {Data: []byte("<html></html>")}}, Patterns: lib, Updated: "2026-10-03"})
	rec := get(h, "/sitemap.xml")
	b := rec.Body.String()
	for _, want := range []string{"<loc>https://bushwhack.sstools.co/</loc>", "<loc>https://bushwhack.sstools.co/guide</loc>", "<loc>https://bushwhack.sstools.co/make</loc>", "<loc>https://bushwhack.sstools.co/patterns/bat</loc>", "<lastmod>2026-10-03</lastmod>"} {
		if !strings.Contains(b, want) {
			t.Errorf("sitemap missing %s", want)
		}
	}
	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/xml") {
		t.Errorf("sitemap type %q", rec.Header().Get("Content-Type"))
	}
	if get(newTestServer(t), "/robots.txt").Code != 200 {
		t.Error("robots.txt not served")
	}
}

func TestGzip(t *testing.T) {
	h := newTestServer(t)
	rec := get(h, "/api/patterns", "Accept-Encoding", "gzip, br")
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("JSON not gzipped")
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.NewDecoder(zr).Decode(&v); err != nil {
		t.Fatal(err)
	}
	if get(h, "/fonts/f.woff2", "Accept-Encoding", "gzip").Header().Get("Content-Encoding") != "" {
		t.Error("fonts should not be gzipped")
	}
	if get(h, "/api/patterns").Header().Get("Content-Encoding") != "" {
		t.Error("gzipped without Accept-Encoding")
	}
}

func TestPatternsAPI(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/patterns", nil))
	var out struct {
		Occasions []struct {
			Slug     string            `json:"slug"`
			Patterns []json.RawMessage `json:"patterns"`
		} `json:"occasions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || len(out.Occasions) != 8 {
		t.Fatalf("bad patterns response: %v %.200s", err, rec.Body)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/patterns/bat.svg", nil))
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/svg+xml" || !strings.Contains(rec.Header().Get("Content-Disposition"), "bat.svg") {
		t.Errorf("svg download: %d %v", rec.Code, rec.Header())
	}
}

func TestParseOptions(t *testing.T) {
	o, err := ParseOptions(fields{"bushing": "8mm", "bit": "1/4", "mode": "piece"}.get)
	if err != nil || o.Mode != "piece" || o.BitDia != 0.25 || o.BushingOD < 0.3149 || o.BushingOD > 0.315 {
		t.Errorf("%+v %v", o, err)
	}
}

type fields map[string]string

func (f fields) get(k string) string { return f[k] }
