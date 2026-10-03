package server

import (
	"bytes"
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
		"index.html":          {Data: []byte(`<html><script>boot()</script></html>`)},
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
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || len(out.Occasions) != 6 {
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
