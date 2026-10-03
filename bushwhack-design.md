# Bushwhack: Design Proposal

Companion to `router-template-service.md` (geometry, pipeline, Go packages, API). This file covers brand, visual design, pages, components, copy, and the starter pattern library. Where the two documents touch (API fields, routes), this file lists the additions needed.

---

## 1. Brief

**Product:** Bushwhack turns an SVG into a router template sized for your guide bushing and bit, so the router cuts the shape you drew.

**Audience:** Hobby woodworkers and makerspace members with a plunge or trim router, a set of guide bushings, and a laser cutter, CNC, or scroll saw to make the template. They know what a bushing is. Many have never thought about the offset math.

**Primary job:** Get someone from "I have a shape" to "I have a template file at true size" in under a minute, and make them trust the result.

**Secondary job:** Inspire builds for holidays and special events through a seasonal library of starter patterns.

## 2. Design concept: the workbench

The site is a workbench. The working surface is a green self-healing cutting mat with an inch grid. Templates appear as hardboard sheets with openings cut through them. The original drawing is a chalk line, the same blue a carpenter snaps on a board. The router's cut is red. The bushing is brass, because most guide bushings are.

These are real objects from the shop, so the visual language explains the product without extra words: chalk line in, hardboard template out, red cut lands on the chalk line.

**The one bold element** is the hero animation (section 6). Everything else stays quiet.

## 3. Tokens

### Color

| Name | Hex | Role |
|------|-----|------|
| Cutting Mat | `#1F4A3F` | Preview canvas, workbench panels, footer |
| Mat Grid | `#5E8E80` | Grid lines on the mat (decorative only) |
| Shop Wall | `#E4E9E3` | Page background (light mode) |
| Graphite | `#23262A` | Body text, icons |
| Brass | `#C4943A` | Bushing in diagrams, primary buttons, focus rings |
| Hardboard | `#8A5E36` | Template sheet fill in previews |
| Chalk Blue | `#2E6FD8` on light, `#7AA7F0` on mat | Original drawing line |
| Cut Red | `#C93A2E` on light, `#F07A6E` on mat | Simulated router cut line |

Rules:
- Chalk Blue and Cut Red are semantic. Use them only for drawing and cut lines, never for UI chrome, links, or errors.
- Brass is the only action color.
- Errors use Graphite text with a Brass left bar plus an icon. They do not use red, which is reserved for the cut line.

Measured contrast:
- Graphite on Shop Wall: 12.4:1.
- Graphite on Brass (button text): 5.5:1.
- Shop Wall text on Cutting Mat: 9.0:1.
- `#7AA7F0` on mat: 4.1:1. `#F07A6E` on mat: 3.7:1. Both pass 3:1 for graphical objects.
- Mat Grid is 2.7:1. That is intentional, because the grid is decoration and carries no information.

**Dark mode:** do not swap to near-black.
- Page background becomes deep mat `#16302A`, and text becomes Shop Wall (11.5:1).
- Brass lightens to `#D9A84A` (6.5:1 on `#16302A`).
- Previews already sit on the mat, so they don't change.

### Type

- **Display: Big Shoulders Stencil** (Google Fonts, OFL, variable `wght` 100 to 900 and `opsz`). A stencil face is a template you can read, so the type tells the story before the copy does. Use it only for:
  - the wordmark
  - H1 and H2
  - the big dimension readout in the workbench
  Weight 800 for H1, 700 for H2. Tracking 0.01em. The face is condensed, so keep sizes large (H1 clamp 3rem to 5.5rem).
- **Text: Atkinson Hyperlegible Next** (Google Fonts, OFL, variable `wght` 200 to 800 with italics). Use it for everything else: body, labels, inputs, tables. Its distinct `0/O` and `1/l/I` forms matter when people read fractions and decimals off a screen in a dusty shop. Body is 1.0625rem with 1.55 line height. Use tabular figures in tables (`font-variant-numeric: tabular-nums`), and verify the font supports `tnum`. If it doesn't, right-align the numbers and accept proportional figures.
- **Self-host** both as woff2 in `web/fonts/` and serve them with `embed.FS`. No requests to Google at runtime, which keeps the single-binary promise.

Type scale (1.25 ratio): 0.85, 1.0625 (body), 1.33, 1.66, 2.07, 2.6, then the display clamp. Line length max 68ch for prose.

Avoid:
- all-caps labels
- eyebrow labels above headings
- monospace for numbers
- single accented words inside headlines

### Shape and space

- Spacing on an 8px base.
- Two radii only: 2px for inputs and buttons (machined edge feel) and the full-round pill for the mode toggle. Cards don't get a uniform radius. Pattern tiles are square-cornered hardboard swatches.
- No drop shadows. Depth comes from color (hardboard on mat).

### Logo mark

Two concentric circles: a brass ring (bushing) around a graphite dot (bit), with a thin chalk-blue arc just outside the ring. It reads as a bushing in plan view, and it works at favicon size. The wordmark "Bushwhack" is set in Big Shoulders Stencil 800, sentence case.

## 4. Voice and copy

Plain shop talk, sentence case, active verbs. Specific over clever. Use the user's words: "bushing," "bit," "template," "cut," "hole," "piece." Never use "offset geometry" in UI copy. That term belongs in the Guide.

Core copy:
- **Hero headline:** "Templates that cut where you drew."
- **Hero sub:** "Upload an SVG, tell us your guide bushing and bit, and get a template sized so your router lands right on the line."
- **Primary button:** "Make a template"
- **Secondary button:** "Browse patterns"
- **Download buttons:** "Download template" and "Download preview" (the result toast says "Template downloaded")
- **Import reminder** (always shown next to the download): "Import at 100%. Don't resize it, or the bushing math is off."

Error and warning copy states what happened and what to do:
- "This file has no closed shapes. Bushwhack needs outlines it can cut around, like paths or polygons."
- "The wood between the nose and mouth is 0.055 in thick. Your bushing can break through anything under 1/4 in. Make the design bigger or move those shapes apart."
- "2 shapes merged into one opening. Space them farther apart or make the design bigger."
- "Some detail is too small for a 1/8 in bit and won't show up in the cut. It's marked on the preview."

Empty workbench: "Drop an SVG here, or start with a pattern." Add a row of three seasonal pattern tiles.

## 5. Information architecture

| Route | Page | Purpose |
|-------|------|---------|
| `/` | Home | Hero demo, how it works, seasonal patterns, footer |
| `/make` | Workbench | Upload, settings, preview, results, download |
| `/make?pattern=classic-jack` | Workbench | Opens with a pattern preloaded |
| `/patterns` | Patterns | Library grouped by occasion, upcoming holiday first |
| `/patterns/{slug}` | Pattern detail | Preview, recommended size, notes, "Open in workbench", SVG download |
| `/guide` | Guide | How bushing offsets work, corners, holes vs pieces, template tips, safety |
| `/api/process` | API | See service spec |
| `/api/patterns` | API | Pattern metadata JSON (new, see section 9) |

Header: wordmark on the left, then "Make", "Patterns", and "Guide" as plain text links on the right. On mobile the links stay visible as text (there are only three, so no hamburger).

## 6. Home page

```
+---------------------------------------------------------------+
| Bushwhack                              Make  Patterns  Guide  |
+---------------------------------------------------------------+
|                                  |                            |
|  Templates that cut              |   [ HERO DEMO ON MAT ]     |
|  where you drew.                 |   chalk outline            |
|                                  |   hardboard grows out      |
|  Upload an SVG, tell us your     |   brass bushing travels    |
|  guide bushing and bit...        |   red cut lands on chalk   |
|                                  |                            |
|  [Make a template] Browse patterns                            |
+---------------------------------------------------------------+
|  How it works   1 Upload   2 Set bushing + bit   3 Download   |
+---------------------------------------------------------------+
|  Coming up: Halloween          [tile] [tile] [tile] [tile]    |
|  See all patterns                                             |
+---------------------------------------------------------------+
|  footer on mat green: Guide, GitHub, made at Sequoia Fabrica  |
+---------------------------------------------------------------+
```

- **Alignment:** left-aligned text throughout. The hero is 5/12 copy and 7/12 demo on desktop. On narrow screens the demo stacks above the copy at full width.
- **Hero demo** (the single orchestrated motion). This is an inline SVG of the Classic Jack mouth, about 6 seconds, played once on load:
  1. The chalk-blue outline draws on (stroke-dashoffset, 1.2s).
  2. The hardboard sheet fades in and its opening grows outward from the chalk line to the template edge (0.8s). The interpolation is precomputed: two paths at the same point count, with the in-between path tweened by JS or SMIL.
  3. A brass bushing circle with a graphite bit dot travels the template edge once (3s). A red cut line trails the bit and lands on the chalk line.
  4. It holds on the final frame, with a small caption: "3/32 in offset for a 5/16 in bushing and 1/8 in bit."

  A "Replay" text button sits under the demo. With `prefers-reduced-motion`, show the final frame and no replay autoplay.
- **How it works:** this content is a real sequence, so numbering is justified. Each step has a one-line description and no icons.
- **Seasonal row:** shows the next upcoming occasion based on today's date (Halloween, then winter holidays, Valentine's, spring, July 4, plus evergreen "Celebrations"). The tiles are hardboard swatches with the pattern's opening cut through them, showing mat green behind.

## 7. Workbench (`/make`)

```
+-------------------+-------------------------------------------+
| Your file         |                                           |
|  [drop zone]      |        PREVIEW ON CUTTING MAT             |
|                   |        1 in grid, rulers top/left         |
| Bushing outside   |        hardboard sheet with openings      |
|  [5/16] in        |        chalk dashed, red cut solid        |
| Bit               |                                           |
|  [1/8 ] in        |   legend: Drawing / Template / Cut        |
|                   |   (each toggles a layer)                  |
| What should match |                                           |
|  (Cut a hole)     +-------------------------------------------+
|  (Cut out a piece)|  6.000 x 4.676 in     (big stencil readout)|
|                   |  Shape  Size        Template     Lost      |
| Size              |  Mouth  3.19 x 1.55 3.38 x 1.73  0.6%      |
|  Longest side [6] |  ...                                       |
|  Fit (Whole page) |  Warnings (brass bar notes)                |
|      (Just shapes)|                                            |
|                   |  [Download template] Download preview      |
| More: min bridge  |  Import at 100%. Don't resize it...        |
+-------------------+-------------------------------------------+
```

- **Layout:** controls column 320px fixed, preview fluid. Below 900px the preview goes first, then the controls, then the results.
- **Measurement inputs:**
  - Accept `5/16`, `0.3125`, `5/16"`, `8mm`, and `1 1/4`.
  - Under each field, show the parsed value in both forms: "5/16 in, 0.3125 in" (or "8 mm, 0.315 in").
  - Invalid input shows the message under the field: "Enter a size like 5/16 or 0.3125."
- **Mode toggle** (pill segmented control):
  - "Cut a hole" means the opening matches your drawing.
  - "Cut out a piece" means the part that falls out matches your drawing.
  - A small diagram under the toggle swaps to show which side of the red line is kept.
- **Fit toggle:** "Whole page" and "Just the shapes," with helper text "Longest side becomes 6 in."
- **Live update:** re-run 300ms after the last change. While running, dim the preview to 70% and keep the old result visible. No spinner overlay.
- **Preview canvas:**
  - SVG on the mat with a 1 in grid and rulers labeled in inches.
  - Zoom with buttons and pinch/scroll. Pan by drag.
  - Draw layers so color isn't the only signal: drawing in dashed Chalk Blue, template as a Hardboard fill with a Brass 1px edge, cut in solid Cut Red.
  - Hovering or focusing a row in the results table highlights that shape on the canvas.
- **Thin bridge marker:** at the thinnest bridge, draw a brass bushing circle at true size with a short leader line and the measurement. Clicking the matching warning scrolls to and pulses that marker once. This is user-triggered motion, so it's fine.
- **Lost detail marker:** shade the areas the bit can't reach with a fine Cut Red hatch.
- **Results:**
  - The page size goes in the big stencil readout.
  - The table shows shape name (from the SVG `id` if present, else "Shape 1"), size, template size, and lost %.
  - Warnings appear as brass-bar notes.
- **Settings persistence:** remember bushing, bit, and mode in `localStorage` (wrapped in try/catch). Do not store the uploaded files.

## 8. Patterns

- **`/patterns`:**
  - Sections per occasion. The upcoming occasion comes first, then the rest in calendar order, then "Celebrations" (birthdays, weddings, anniversaries).
  - Each section heading is in stencil H2 with a one-line description.
  - Tiles are square hardboard swatches with the openings showing mat, plus the name below in Atkinson. No cards and no shadows.
- **Pattern detail:**
  - A large preview at the recommended size.
  - The recommended size and the smallest safe size for the default 5/16 bushing and 1/8 bit, computed at build time by the pipeline.
  - A short build note (wood, finish, display idea).
  - Actions: "Open in workbench" (primary) and "Download SVG".
- **License line:** every pattern is original to Bushwhack and released as CC0. Say so on each page: "Free to use, sell what you make."

## 9. Starter pattern library

### Launch set (start with Halloween; it's October)

| Occasion | Patterns |
|----------|----------|
| Halloween | Classic Jack (existing), Grinning Jack, Bat, Black Cat, Ghost, Crescent Moon |
| Winter holidays | Snowflake, Pine Tree, Star, Ornament, Mitten |
| Valentine's | Heart, Two Hearts, Arrow Heart |
| Spring | Egg, Tulip, Bunny |
| July 4 | Star, Burst |
| Celebrations | Balloon, Cake, Ring Pair, numerals 0 to 9 (original chunky shapes, not from a font) |

Draw the patterns in-house. Do not trace commercial clip art or characters.

### Pattern design rules

These are enforced by a `go test` that runs every pattern through the pipeline:
- Closed paths only. No strokes, text, images, or transforms beyond translate and scale.
- At the recommended size with a 5/16 bushing and 1/8 bit:
  - no "openings merged" warning
  - minimum bridge 1/4 in or more
  - lost area under 3% per shape
- Minimum feature width 3/16 in at the recommended size.
- A 0.25 in margin from the canvas edge.
- Canvas is square or close to it, viewBox in inches, `width`/`height` in inches.

### Storage

```
patterns/
  halloween/
    classic-jack.svg
    classic-jack.json
  winter/
    snowflake.svg
    snowflake.json
```

`{slug}.json`:
```json
{
  "slug": "classic-jack",
  "name": "Classic Jack",
  "occasion": "halloween",
  "recommendedSizeIn": 6,
  "fit": "artwork",
  "note": "Cut through 1/2 in plywood, back it with orange acrylic, and light it from behind.",
  "license": "CC0-1.0"
}
```

Embed `patterns/` with `embed.FS`. At startup, load the metadata, validate it, and fail fast on bad patterns. `GET /api/patterns` returns all of it, plus computed `minSafeSizeIn` per pattern.

The occasion order and date windows live in one Go table, so "upcoming" is deterministic and testable:
- Halloween: Sep 1 to Oct 31
- Winter: Nov 1 to Dec 31
- Valentine's: Jan 1 to Feb 14
- Spring: Feb 15 to Apr 30
- July 4: May 1 to Jul 4
- Celebrations: the rest, and always listed last

## 10. API additions to the service spec

- `/api/process` accepts `pattern={slug}` as an alternative to file upload.
- Add `minBridgeAt: {"xIn": 2.81, "yIn": 2.40}` to the response for the bridge marker, and `lostRegionsSvg` (path data) for the hatch overlay.
- Add `name` per shape, from the SVG `id` or `inkscape:label` if present.
- The preview SVG from the API is the downloadable one (blue/black/red on white, for printing). The workbench draws its own mat-style preview client-side from three path layers. Return those as `layers: {drawing, template, cut}` path data in inches.

## 11. Implementation notes for Claude Code

- Use Go `html/template` server-rendered pages and vanilla JS modules. No framework and no build step. CSS goes in one file with custom properties for every token above.
- `web/` layout:
  ```
  web/
    templates/  base.html, home.html, make.html, patterns.html, pattern.html, guide.html
    static/
      css/bushwhack.css
      js/make.js, demo.js, units.js
      fonts/  big-shoulders-stencil.woff2, atkinson-hyperlegible-next.woff2 (+ italic)
      img/    logo.svg, og-image.png
  ```
- `units.js` mirrors the Go measurement parser. Share test cases between them via `testdata/units.json`.
- The hero demo paths are generated at build time by a small Go command (`cmd/gen-demo`) that runs the pipeline on the Classic Jack mouth and writes resampled paths with matching point counts. Don't hand-author them.
- Open Graph image: the final hero frame rendered to a 1200x630 PNG.

Accessibility and quality floor:
- visible focus (2px Brass outline, 2px offset)
- full keyboard operation of the toggles and zoom
- the preview gets an `aria-label` summarizing size and warnings
- each warning is a live region
- respect reduced motion
- test at 320px width

Screenshot each page in light and dark mode at 1440px and 390px and review them before calling a milestone done.

## 12. Plan review: defaults avoided

This plan was checked against common generated-site patterns:
- **Cream background with a clay accent:** replaced by a cool Shop Wall background and brass, taken from real bushing material.
- **Near-black dark mode:** replaced by deep mat green.
- **Uniform rounded cards with soft shadows:** replaced by square hardboard swatches with real openings, and no shadows.
- **Hero showing a big stat with a gradient:** replaced by a live demonstration of the actual product behavior.
- **Monospace numbers and all-caps eyebrows:** replaced by Atkinson Hyperlegible Next for numbers and plain headings.
- **Numbered steps:** kept, because "How it works" is a true sequence.
