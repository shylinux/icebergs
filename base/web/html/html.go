package html

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

const (
	H1   = "h1"
	H2   = "h2"
	H3   = "h3"
	SPAN = "span"

	STYLE  = "style"
	WIDTH  = "width"
	HEIGHT = "height"

	BACKGROUND_COLOR = "background-color"

	COLOR = "color"
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
	VALUE  = "value"
	OUTPUT = "output"
	LAYOUT = "layout"
	RESIZE = "resize"
	FILTER = "filter"
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

const (
	GetLocation      = "getLocation"
	ConnectWifi      = "ConnectWifi"
	GetClipboardData = "getClipboardData"
	ScanQRCode       = "scanQRCode"
	ChooseImage      = "chooseImage"
	Record1          = "record1"
	Record2          = "record2"
)

func Format(tag string, inner string, arg ...string) string {
	return kit.Format("<%s %s>%s</%s>", tag, kit.JoinProperty(arg...), inner, tag)
}
func FormatDanger(value string) string {
	return Format(SPAN, value, STYLE, kit.JoinCSS(
		BACKGROUND_COLOR, "var(--danger-bg-color)",
		COLOR, "var(--danger-fg-color)",
	))
}
