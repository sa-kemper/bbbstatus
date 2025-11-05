package web

import "embed"

//go:embed static
var StaticFS embed.FS

//go:embed views/*.gohtml
var Views embed.FS
