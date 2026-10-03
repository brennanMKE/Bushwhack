# Router Template Offset Service

Working title: **TemplateOffset** (rename freely).

A Go web service that takes an SVG of shapes and produces a router template sized for a specific guide bushing and bit. The user uploads an SVG, enters bushing OD, bit diameter, target size, and cut mode, then downloads a true-size SVG template plus a preview showing what the router will actually cut.

A working Python prototype exists (`router_template.py`, using shapely + svgpathtools). Treat it as the reference implementation and match its output numerically (see Golden Test Values).

---

## 1. Why this exists

When routing with a guide bushing, the bushing rides against the template edge and the bit is centered inside the bushing. The cut therefore lands a fixed distance away from the template edge. If you draw the template at the exact size of the final shape, the result comes out wrong and tight features are hard or impossible for the bushing to enter. The fix is to offset every shape in the template by the right amount.

## 2. Geometry

Definitions (all in inches internally):

```
bushingR = bushingOD / 2
bitR     = bitDiameter / 2
```

The bushing wall touches the template edge, so the bushing (and bit) center sits `bushingR` from the edge. The bit sweeps a band from `bushingR - bitR` to `bushingR + bitR` away from the template edge.

```
template edge
     |
     |<-- bushingR - bitR -->|  near edge of bit (wall of the hole)
     |<-- bushingR -------------->|  bit center
     |<-- bushingR + bitR ---------------->|  far edge of bit (edge of a freed piece)
```

### Cut modes

| Mode | What matches the drawing | Template offset (outward, per opening) |
|------|--------------------------|----------------------------------------|
| `hole` (default) | The opening cut in the workpiece (jack-o'-lantern faces, pockets) | `bushingR - bitR` |
| `piece` | The part that falls out from inside the opening (inlay pieces) | `bushingR + bitR` |

Example: 5/16" bushing, 1/8" bit: `hole` offset = 0.09375" (3/32"), `piece` offset = 0.21875" (7/32").

v1 assumes a **female template** (openings in a sheet, bushing rides inside each opening). Male templates (bushing rides outside a positive shape) are the same math with the offset sign flipped; list as a v2 option.

### Corner behavior (explain in UI, not a bug)

- Offset with **round joins**. Rounding the template corners to radius `offset` is safe because `offset < bushingR`, so the bushing still reaches as far as it physically can. Miter joins add nothing and create spikes on acute corners.
- Sharp inside corners of the final cut always round to `bitR`. Unavoidable.
- Any feature of the drawing narrower than the bit diameter cannot be cut. After offsetting, those same features are narrower than the bushing, so the two conditions agree.

### Router simulation (used for preview and warnings)

```
centerPath = inset(template, bushingR)          // where the bushing center can travel
hole mode:  cut = outset(centerPath, bitR)
piece mode: cut = inset(centerPath, bitR)
```

Compare `cut` to the original shapes to report lost area per shape.

## 3. Scaling

SVGs arrive in arbitrary user units, and the physical size must be explicit. Support three options:

- `fit=canvas` (default): scale so the SVG canvas (viewBox, else width/height) longest side equals `size` inches.
- `fit=artwork`: scale so the bounding box of the drawn shapes' longest side equals `size` inches.
- `scale`: explicit inches per user unit. This overrides the other two.

Output page: same size as the scaled canvas so shape positions match the original. Grow the page (and shift shapes) only if an offset shape would cross the edge, and report that in the warnings.

Output SVG must declare physical units: `width="6in" height="4.676in" viewBox="0 0 6 4.676"`. Tell the user to import at 100% in their laser/CNC software and never rescale.

## 4. Processing pipeline

1. **Parse SVG** (`encoding/xml`). Collect `<path>`, `<polygon>`, `<polyline>`, `<rect>`, `<circle>`, `<ellipse>`. Apply nested `transform` attributes (matrix, translate, scale, rotate, skewX, skewY). Ignore `<text>` and `<image>` and add a warning that they were ignored.
2. **Parse path data.** Write a small parser for M/L/H/V/C/S/Q/T/A/Z, absolute and relative, including implicit repeated commands and number formats like `1e-3`, `.5.5`, `-1-2`. Arcs need the SVG spec's endpoint-to-center conversion (SVG 1.1, Appendix F.6).
3. **Flatten curves** to polylines. Tolerance should be about 0.002" at final physical size, so compute it from the scale. Adaptive subdivision for beziers is preferable to fixed steps.
4. **Build polygons.** Each closed subpath becomes a ring. Use even-odd fill within a single `<path>` so nested subpaths become holes. Treat unclosed subpaths as closed and add a warning.
5. **Transform to inches** using the chosen scale.
6. **Offset** each shape by the mode's offset with round joins.
7. **Union** the offset shapes. If the count of resulting polygons is lower than the input count, openings merged: warn.
8. **Bridge check.** Minimum distance between distinct offset openings. Warn if below `minBridge` (default 0.25"). In the Classic Jack test file this catches a 0.055" bridge between nose and mouth.
9. **Simulate** the cut (section 2) and compute lost area % per shape, ignoring slivers under about 0.005".
10. **Emit** `template.svg` (black hairline outlines, stroke 0.01") and `preview.svg` (original blue, template black, simulated cut red).

## 5. Go implementation

### Geometry library

Use a **pure Go Clipper2 port** for offsetting and boolean ops. Avoid cgo so the binary stays a single static file. Candidates found (evaluate both, pick one, record the decision in `DECISIONS.md`):

- `github.com/bolom009/go-clipper2`: pure Go, integer-precision, has `ClipperOffset` with `NewClipperOffset(miterLimit, arcTolerance, ...)` and `AddPaths(paths, joinType, endType)`.
- `github.com/cwbudde/go-clipper2/port`: pure Go port with offsetting examples.
- `github.com/epit3d/goclipper2`: linked from upstream Clipper2 README, but cgo-based wrapper. Use only if both pure ports fail tests.

Evaluation criteria: correct round-join offsets on the golden file, handles holes/polytree, license compatible, maintained, no panics in fuzzing.

Clipper works in integer coordinates. Either use its double-precision API or scale inches by 1e5 to int64. Set `arcTolerance` so round joins are accurate to about 0.001".

Distance between polygons (bridge check) is not part of Clipper. Implement as minimum segment-to-segment distance across rings. O(n*m) is fine at these sizes. Add a bounding-box prefilter.

### Package layout

```
cmd/server/main.go           HTTP server, flags, embeds web/
internal/svgin/              XML walk, transforms, shape elements -> path commands
internal/pathdata/           path "d" tokenizer/parser, arc conversion
internal/flatten/            bezier/arc flattening with tolerance
internal/geom/               polygon types, even-odd assembly, clipper wrapper, distance
internal/template/           pipeline: scale, offset, union, simulate, warnings (pure, no I/O)
internal/svgout/             template + preview SVG writers
web/                         index.html, app.js, style.css (embedded with embed.FS)
testdata/                    Classic_Jack.svg + expected outputs
```

Keep `internal/template` a pure function: `Process(input []byte, opts Options) (Result, error)`. Everything else is plumbing. This makes it trivial to also ship a CLI (`cmd/templateoffset`) later.

```go
type Options struct {
    BushingOD  float64 // inches
    BitDia     float64 // inches
    Mode       string  // "hole" | "piece"
    Fit        string  // "canvas" | "artwork"
    Size       float64 // inches, longest side
    Scale      float64 // inches per unit; 0 = unset
    MinBridge  float64 // inches
}

type ShapeReport struct {
    Index                     int
    WidthIn, HeightIn         float64
    TemplateWidthIn, TemplateHeightIn float64
    LostAreaPct               float64
}

type Result struct {
    TemplateSVG []byte
    PreviewSVG  []byte
    PageWidthIn, PageHeightIn float64
    ScaleInPerUnit float64
    Offset      float64
    Shapes      []ShapeReport
    MinBridgeIn float64
    Warnings    []string
}
```

### HTTP API

- `GET /` serves the UI.
- `POST /api/process`: multipart with `file` (SVG) plus option fields. Returns JSON:
  ```json
  {
    "templateSvg": "<svg ...>",
    "previewSvg": "<svg ...>",
    "page": {"widthIn": 6.0, "heightIn": 4.676},
    "scaleInPerUnit": 0.02069,
    "offsetIn": 0.09375,
    "shapes": [{"index":1,"widthIn":0.560,"heightIn":0.748,"templateWidthIn":0.747,"templateHeightIn":0.936,"lostAreaPct":1.8}],
    "minBridgeIn": 0.055,
    "warnings": ["Thin bridge: 0.055 in between openings (min 0.25 in)"]
  }
  ```
- `GET /healthz`.

Stateless; nothing stored server-side in v1.

### UI

Single page, no framework required (vanilla JS is enough). Units entered as fractions or decimals (`5/16`, `0.3125`, `8mm`). Parse on the server too, never trust the client.

- Upload (drag and drop).
- Inputs: bushing OD, bit diameter, mode, fit, size, min bridge.
- Live preview re-runs on input change (debounced 300 ms).
- Shape table and warnings list.
- Download buttons for template SVG and preview SVG.
- Short explainer of the geometry with the diagram from section 2.

### Input safety

- Max upload 2 MB; `http.MaxBytesReader`.
- `encoding/xml` does not expand custom entities. Still reject any `<!DOCTYPE` / `<!ENTITY` to be safe.
- Cap element count (e.g. 20k), path command count, and flattened vertex count (e.g. 500k). Return 422 with a clear message.
- Per-request timeout (e.g. 10 s) via context; check it inside flattening and offset loops.
- Output SVG is generated by us, never echo input markup. Serve with `Content-Type: image/svg+xml` and a strict CSP if served as files.

## 6. Testing

### Golden test values (from the Python reference, `Classic_Jack.svg`, 290x226 canvas, 4 closed paths)

Default options (bushing 0.3125, bit 0.125, `hole`, `fit=canvas`, size 6):

| shape | width | height | tmpl w | tmpl h | lost area |
|-------|-------|--------|--------|--------|-----------|
| 1 | 0.560 | 0.748 | 0.747 | 0.936 | 1.8% |
| 2 | 3.192 | 1.546 | 3.379 | 1.733 | 0.6% |
| 3 | 0.568 | 0.658 | 0.755 | 0.845 | 1.9% |
| 4 | 0.554 | 0.678 | 0.741 | 0.865 | 3.0% |

- Scale 0.02069 in/unit, offset 0.0938 in, page 6.000 x 4.676 in.
- Min bridge 0.055 in (must produce the thin bridge warning).

With `fit=artwork`: scale 0.03889 in/unit, mouth (shape 2) 6.000 x 2.906, template 6.187 x 3.093, min bridge 0.269 in (no warning).

Shape order follows document order of `<path>` elements. Tolerances: dimensions within 0.003", lost area within 0.5 percentage points.

### Unit tests

- Path parser: every command, relative and absolute, implicit repeats, compact number syntax, arcs with flags written as `1 0 1 10 10` and `10010 10`.
- Transforms: nested groups, all transform functions.
- Offset: a 1" square in hole mode with 5/16 and 1/8 produces a template with bounding box 1.1875" and corner radius 0.09375".
- Simulation: square from above round-trips to a 1" square with 0.0625" corner radii; lost area equals `(4 - pi) * 0.0625^2` per corner set.
- Even-odd: a ring inside a ring becomes a hole, and the hole shrinks when offset outward.
- Bridge: two squares 0.3" apart in hole mode at default sizes report 0.3 - 2 * 0.09375 = 0.1125".

### Fuzz

`go test -fuzz` on the path parser and on `Process` with random small SVGs. Must never panic.

## 7. Milestones

1. `internal/pathdata` + `internal/svgin` + flatten, with tests. Dump flattened polygons for `Classic_Jack.svg` and compare against the Python prototype's vertex count and bounds.
2. Clipper evaluation and `internal/geom`. Record choice in `DECISIONS.md`.
3. `internal/template` pipeline passing golden tests.
4. `cmd/templateoffset` CLI (same flags as the Python script) for quick manual checks.
5. HTTP server + embedded UI.
6. Safety limits, fuzzing, Dockerfile (distroless static), deploy.

## 8. Later ideas

- Male template mode.
- DXF output for CNC users.
- mm units throughout.
- Bushing/bit presets the user can save in the browser.
- Auto-suggest a scale that clears `minBridge`.
- Optional registration holes or a border frame for clamping the template.
- Show the bushing circle at the tightest spot so users see why a feature is lost.

## 9. Open questions

- Hosting target and domain (could live under sstools.co).
- Whether to keep it free and stateless or add accounts for saved templates.
- License for the code (must be compatible with the chosen Clipper2 port).
