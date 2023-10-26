package html

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

const (
	H1 = "h1"
	H2 = "h2"
	H3 = "h3"
)
const (
	DARK   = "dark"
	LIGHT  = "light"
	WHITE  = "white"
	BLACK  = "black"
	SILVER = "silver"

	PROJECT = "project"
	CONTENT = "content"
	PROFILE = "profile"
	DISPLAY = "display"

	VIEW   = "view"
	INPUT  = "input"
	OUTPUT = "output"
	LAYOUT = "layout"
	RESIZE = "resize"
	FILTER = "filter"

	COLOR            = "color"
	BACKGROUND_COLOR = "background-color"
)

const (
	FLOAT  = "float"
	CHROME = "chrome"

	TEXT_PLAIN = "text/plain"
)

func IsImage(name, mime string) bool {
	return strings.HasPrefix(mime, "image/") || kit.ExtIsImage(name)
}
func IsVideo(name, mime string) bool {
	return strings.HasPrefix(mime, "video/") || kit.ExtIsVideo(name)
}
func IsAudio(name, mime string) bool {
	return strings.HasPrefix(mime, "audio/")
}
