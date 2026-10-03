# Decisions

## Clipper2 port: `github.com/bolom009/go-clipper2` (v1.3.0)

The spec named three candidates:

- **bolom009/go-clipper2**: pure Go, Boost license, tagged releases, tests, and a functional-options `InflatePaths64`. A 1 in square offset by 3/32 in with round joins had area 1.40243 sq in, against an exact 1.40261 at an arc tolerance of 0.0005 in. The simulated cut round-tripped to 0.99652, against an exact 0.99665. Holes shrink correctly under outward offsets.
- **cwbudde/go-clipper2**: its module path is `github.com/go-clipper/clipper2`, so `go get` of the advertised path fails. It has no tagged release. Not used.
- **epit3d/goclipper2**: cgo. Not needed.

Working fixed-point at 1e5 units per inch. Known quirk: bolom009's `GetBounds64` returns max-int for the right and bottom bounds, so `geom.Bounds` computes bounds itself.

## Front end: SvelteKit SPA, embedded

`bushwhack-design.md` §11 asked for server-rendered `html/template` with no build step. The owner prefers Svelte for SPAs, so the front end is SvelteKit 3 with `adapter-static` in SPA mode (`fallback: index.html`, `ssr = false`). The routes are real paths (`/make`, `/patterns/bat`, `/guide`). The Go server serves `index.html` for any path that isn't a file or `/api/*`. `vite build` output is embedded with `embed.FS`, so deployment is still a single static binary. Fonts are self-hosted from `web/static/fonts`, with no runtime requests to Google.

The CSP stays strict without `'unsafe-inline'`. At startup the server hashes the SPA shell's inline bootstrap `<script>`, and the one `style="…"` literal in the built JS (SvelteKit's route announcer). It allows exactly those hashes.

## Preview colours

The original request said blue for the original, red for the template and green for the cut. The design doc, which came later, wins:

- **Workbench (client-drawn on the mat):** chalk-blue dashed drawing, hardboard template with a brass edge, solid red cut, and red hatch where the bit can't reach.
- **Downloadable preview (printable, white):** blue drawing, black template, red cut, and brass circles at thin bridges.

Red is reserved for the cut. Warnings use a brass bar, never red.

## Warnings are structured

`template.Result.Notes` carries `{kind, text, shapes, bridge}`, so the UI can pulse the right bridge marker or highlight a shape when a warning is clicked. Text follows the design voice. Shape names come from the SVG `id` or `inkscape:label` ("the left eye"). Generic ids like `path123` fall back to "shape 3".

## Spec deviations

- **Piece mode, convex corners:** the spec's square test assumed rounded corners in both modes. In piece mode, outside corners of the piece stay sharp, because the bit's far edge wraps them. The tests assert that.
- **Overcut:** besides lost area, each shape reports *overcut*: narrow notches into a shape that the template fills in and the bit routes away. It's flagged at 5% or more.
- **Floating islands:** a shape with a counter (an O) leaves a loose piece of template. It's flagged.
- **Fit `document`:** besides `canvas` and `artwork`, there's a `document` fit. It uses the SVG's own physical width/height (in, mm or cm), which suits Inkscape files.
- **Lost-area warning:** a shape is flagged at 5% or more lost area, or as "can't be cut" at 99.5% or more. The golden file's worst shape loses 3%.

## Not done yet

- The golden test values in the spec come from `Classic_Jack.svg` and the Python prototype, neither of which is in this repo. The `patterns/halloween/classic-jack.svg` here is a new original drawing, not that file.
- Male templates, DXF output and mm throughout are still in "later ideas".
