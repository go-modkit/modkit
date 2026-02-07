// Package templates provides embedded template files.
package templates

import "embed"

//go:embed *.tpl
var content embed.FS

// FS returns the embedded filesystem containing templates.
func FS() embed.FS {
	return content
}
