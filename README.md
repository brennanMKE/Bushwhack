# Bushwhack

Turns an SVG into a router template sized for your guide bushing and bit, so the router cuts the shape you drew. Live at <https://bushwhack.sstools.co>.

- `router-template-service.md`: the geometry, pipeline and API spec.
- `bushwhack-design.md`: the brand, pages and pattern library design.
- `DECISIONS.md`: choices made while building it.
- `DEPLOY.md`: how it runs on photon.

## Layout

```
cmd/server        HTTP server; embeds the built SPA and the pattern library
cmd/bushwhack     CLI: bushwhack -bushing 5/16 -bit 1/8 face.svg
cmd/gen-demo      regenerates the home page animation (make demo)
internal/pathdata SVG path "d" parser, arcs to cubics
internal/svgin    XML walk, transforms, basic shapes
internal/flatten  adaptive bezier flattening
internal/geom     Clipper2 wrapper, distances
internal/template the pipeline: Process(ctx, svg, opts) → Result (pure, no I/O)
internal/svgout   template and printable preview writers
internal/units    "5/16", "1 1/4", "8mm" parsing (mirrored in web/src/lib/units.ts)
internal/server   API, pages and SPA fallback, sitemap, CSP, gzip
patterns/         CC0 starter patterns (SVG + JSON) and the occasion calendar
web/              SvelteKit (adapter-static): content pages prerendered, the rest an SPA, built into web/build
testdata/         shared cases for the Go and TypeScript unit parsers
deploy/           systemd unit, Apache vhosts, deploy script
```

## Develop

Needs Go 1.25+ and Node 22+.

```sh
make build          # npm build + Go binaries into bin/
make run            # serves http://localhost:8080
make dev            # Vite on :5173 with hot reload (run `make run` alongside)
make test           # go vet, go test, vitest, svelte-check
make demo           # rebuild the hero animation after editing classic-jack.svg
```

`go build` needs `web/build` to exist. Run `make web` once after cloning.

Every pattern in `patterns/` must pass `go test ./patterns/`. At its recommended size, with a 5/16 in bushing and 1/8 in bit, that means:

- no merged openings
- every bridge at least 1/4 in
- under 3% lost area per shape
- features at least 3/16 in wide
- a 0.25 in margin
- an inch-based canvas
- a recommended size of at most 5 in, so the template fits a 6 × 6 in jig with a 3/8 in border

## License

The code is MIT (`LICENSE`). The starter patterns in `patterns/` are CC0 (`patterns/LICENSE`). Fonts in `web/static/fonts` are under the SIL Open Font License: Big Shoulders Stencil and Atkinson Hyperlegible Next. Polygon offsetting uses [go-clipper2](https://github.com/bolom009/go-clipper2) (Boost).
