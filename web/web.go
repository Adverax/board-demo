// Package web is the board page (COMP-003): plain HTML, CSS and JavaScript
// with no build step (CON-005), embedded into the binary with go:embed.
//
// Layout: index.html is the page served at GET /; static/ holds the assets it
// links, served at /static/<name>. static/tokens.css is a derived copy of
// meta/design/tokens.css (go:embed cannot reach parent directories); a test
// keeps the two identical.
package web

import (
	"embed"
	"io/fs"
)

// The patterns name every file explicitly: a missing file is a compile error,
// not a 404 at runtime.
//
//go:embed index.html static/tokens.css static/app.css static/app.js
var files embed.FS

// Page returns the embedded page: index.html at the root, assets under static/.
func Page() fs.FS { return files }
