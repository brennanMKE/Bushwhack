# Bushwhack SEO Preparation

Site: https://bushwhack.sstools.co/
Stack: Go web server
Goal: get the landing page and tool indexed and ranking for router template and guide bushing queries.

## Current state

Already present in `<head>`:

- `<title>`: "Bushwhack: templates that cut where you drew"
- `meta description`
- Open Graph: `og:title`, `og:description`, `og:image`, `og:url`
- `twitter:card` = `summary_large_image`
- `viewport`, `theme-color`

Problem observed: an external fetch of the page returned the metadata but no body text (no headings or paragraphs). Confirm whether content is rendered client-side.

---

## Priority 1: Crawlability (blockers)

### 1.1 Server-render the body content

Verify:

```sh
curl -s https://bushwhack.sstools.co/ | grep -iE '<h1|<h2|<p'
```

If empty, render the landing page content in the Go HTML template so it exists in the initial response. The interactive tool can still hydrate with JS, but the explanatory text, headings and links must be in the raw HTML.

Acceptance:
- `curl` output contains exactly one `<h1>` and readable paragraph text.
- Page is usable (content readable) with JS disabled.

### 1.2 robots.txt

Serve at `/robots.txt`:

```
User-agent: *
Allow: /

Sitemap: https://bushwhack.sstools.co/sitemap.xml
```

Disallow any upload/result/API endpoints that produce per-user output (for example `/api/`), so crawlers don't index generated templates.

### 1.3 sitemap.xml

Serve at `/sitemap.xml`. Generate from a list of public routes in Go so new pages are added automatically. Include `<lastmod>`.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://bushwhack.sstools.co/</loc>
    <lastmod>2026-10-03</lastmod>
  </url>
</urlset>
```

### 1.4 Canonical URLs

Every page gets a self-referencing canonical:

```html
<link rel="canonical" href="https://bushwhack.sstools.co/">
```

Also:
- Redirect `http://` to `https://` with a 301.
- Pick one form for trailing slashes and 301 the other.
- Tool result URLs with query strings should either canonicalize to the base page or be `noindex`.

---

## Priority 2: Head and markup fixes

Add to the base template:

```html
<html lang="en">
<meta property="og:type" content="website">
<meta property="og:site_name" content="Bushwhack">
<meta name="twitter:title" content="Bushwhack: templates that cut where you drew">
<meta name="twitter:description" content="Router templates sized for your guide bushing and bit.">
<meta name="twitter:image" content="https://bushwhack.sstools.co/og-image.png">
<link rel="icon" href="/favicon.svg" type="image/svg+xml">
<link rel="icon" href="/favicon.ico" sizes="any">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">
```

Page title and description should be per-page template variables, not hardcoded.

### OG image

- Size: 1200×630 PNG, under ~300 KB.
- Check with LinkedIn Post Inspector and opengraph.xyz.

### Semantic structure

- One `<h1>` per page.
- Use `<main>`, `<header>`, `<footer>`, `<nav>`.
- Form inputs (bushing OD, bit diameter, SVG upload) have `<label>` elements.
- Example images and diagrams have descriptive `alt` text.

---

## Priority 3: Structured data

Add JSON-LD to the landing page:

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "WebApplication",
  "name": "Bushwhack",
  "url": "https://bushwhack.sstools.co/",
  "description": "Upload an SVG, enter your guide bushing and bit, and get a router template offset so the cut lands on your line.",
  "applicationCategory": "DesignApplication",
  "operatingSystem": "Any",
  "offers": { "@type": "Offer", "price": "0", "priceCurrency": "USD" },
  "author": { "@type": "Person", "name": "Brennan Stehling", "url": "https://brennan.sstools.co/" }
}
</script>
```

For tutorial pages, use `HowTo` or `Article` schema. Validate with Google's Rich Results Test and validator.schema.org.

---

## Priority 4: Content

A tool page with little text ranks poorly. Search traffic comes from explanatory content around the tool.

### Target queries

- router template maker / generator
- guide bushing offset calculator
- template guide bushing offset
- router inlay template
- SVG to router template
- how to make a router template

### Landing page content (below the tool)

Add a section explaining:

1. What a guide bushing offset is.
2. The formula: `offset = (bushing OD - bit diameter) / 2`
3. A worked example: 5/16" OD bushing with a 1/8" bit gives an offset of 3/32" (0.09375"). The template is offset inward by that amount for an outside cut.
4. Inside vs outside cuts and which direction the template offsets.
5. Common bushing and bit combinations (table).
6. Short FAQ: template material, minimum inside corner radius (limited by bushing radius), file formats accepted.

### Additional pages (each its own URL, title, description, h1)

- `/guide-bushing-offset` : explainer plus a standalone offset calculator (high search intent, easy win)
- `/inlays` : making inlay templates
- `/signs` : lettering and sign templates
- `/holiday` : seasonal and special-event builds (ornaments, signs, gifts)
- `/starter-files` : free starter SVGs with preview images; each file can have its own page

Each page links back to the tool and to related pages.

---

## Priority 5: Performance

- Run Lighthouse (mobile) and target 90+ on Performance, Accessibility, Best Practices, SEO.
- Core Web Vitals: LCP < 2.5s, CLS < 0.1, INP < 200ms.
- Enable gzip or brotli compression in the Go server.
- Set `Cache-Control` headers on static assets with content-hashed filenames.
- Defer non-critical JS. Set explicit `width`/`height` on images to avoid layout shift.

---

## Priority 6: Indexing and measurement

1. Google Search Console: add a **Domain property** for `sstools.co` (DNS TXT verification) so all subdomains are covered. Submit the sitemap. Use URL Inspection to request indexing of the landing page.
2. Bing Webmaster Tools: import from Search Console. Bing also feeds DuckDuckGo and several AI search tools.
3. Analytics: add a privacy-friendly option (Plausible, Umami, or GoatCounter) to see which pages and referrers bring traffic.

---

## Priority 7: Links and promotion

A new subdomain starts with no authority. Get initial links from:

- brennan.sstools.co (projects section)
- Sequoia Fabrica site (member project or resource page)
- GitHub README if the repo is public
- r/woodworking, r/router, r/CNC (show a build made with a template, not just a link)
- Lumberjocks and Sawmill Creek forums
- YouTube or Instagram build videos linking to the tool

---

## Verification checklist

- [x] `curl` shows h1 and body text in raw HTML
- [x] `/robots.txt` returns 200 and references the sitemap
- [x] `/sitemap.xml` returns 200 and validates
- [x] Canonical tag on every page
- [x] HTTP and trailing-slash variants 301 to canonical
- [x] `lang`, `og:type`, `og:site_name`, favicons added
- [ ] OG image is 1200×630 and previews correctly
- [ ] JSON-LD passes Rich Results Test
- [x] Offset explainer and FAQ on landing page
- [x] `/guide-bushing-offset` page live
- [ ] Lighthouse mobile scores 90+
- [ ] Search Console domain property verified, sitemap submitted
- [ ] Bing Webmaster Tools set up
- [ ] At least 3 inbound links from external sites
