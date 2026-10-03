// Package web embeds the built Svelte SPA (web/build, produced by
// `npm run build`). Run `make web` before `go build`.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var build embed.FS

// FS returns the built SPA rooted at its index.html.
func FS() fs.FS {
	sub, err := fs.Sub(build, "build")
	if err != nil {
		panic(err)
	}
	return sub
}
