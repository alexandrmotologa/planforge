package planforge

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var embeddedDist embed.FS

// GetStaticFS returns the sub-filesystem containing the built frontend assets.
func GetStaticFS() (fs.FS, error) {
	return fs.Sub(embeddedDist, "web/dist")
}
