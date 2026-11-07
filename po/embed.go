// Package po embeds the translation catalogues that ship with incus.
//
// The .po files stay in po/ as their source of truth, so translators and the
// gettext toolchain keep working unchanged; this package only gives the Go
// binary a way to carry them. The pattern matches the catalogues and not
// incus.pot, which is a template and never read at runtime.
package po

import (
	"embed"
)

// Files holds every catalogue embedded at build time.
//
//go:embed *.po
var Files embed.FS
