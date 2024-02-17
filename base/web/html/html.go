package html

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

const (
	Mozilla        = "Mozilla"
	Firefox        = "Firefox"
	Safari         = "Safari"
	Chrome         = "Chrome"
	Edg            = "Edg"
	Mobile         = "Mobile"
	Alipay         = "Alipay"
	MicroMessenger = "MicroMessenger"
	Android        = "Android"
	IPhone         = "iPhone"
	Mac            = "Mac"
	Linux          = "Linux"
	Windows        = "Windows"

	UserAgent       = "User-Agent"
	XForwardedFor   = "X-Forwarded-For"
	XHost           = "X-Host"
	Referer         = "Referer"
	Authorization   = "Authorization"
	Bearer          = "Bearer"
	Basic           = "Basic"
	Accept          = "Accept"
	AcceptLanguage  = "Accept-Language"
	ContentEncoding = "Content-Encoding"
	ContentLength   = "Content-Length"
	ContentType     = "Content-Type"

	ApplicationForm  = "application/x-www-form-urlencoded"
	ApplicationOctet = "application/octet-stream"
	ApplicationJSON  = "application/json"
)
const (
	H1       = "h1"
	H2       = "h2"
	H3       = "h3"
	SPAN     = "span"
	CHECKBOX = "checkbox"

	STYLE  = "style"
	WIDTH  = "width"
	HEIGHT = "height"

	BG_COLOR = "background-color"
	FG_COLOR = "color"
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

	TEXT     = "text"
	TEXTAREA = "textarea"
	PASSWORD = "password"
	SELECT   = "select"
	BUTTON   = "button"

	VIEW    = "view"
	INPUT   = "input"
	VALUE   = "value"
	OUTPUT  = "output"
	LAYOUT  = "layout"
	RESIZE  = "resize"
	REFRESH = "refresh"
	FILTER  = "filter"
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
func FormatA(inner string, arg ...string) string {
	return kit.Format(`<a href="%s">%s</a>`, kit.Select(inner, arg, 0), inner)
}
func FormatDanger(value string) string {
	return Format(SPAN, value, STYLE, kit.JoinCSS(BG_COLOR, "var(--danger-bg-color)", FG_COLOR, "var(--danger-fg-color)"))
}

var SystemList = []string{
	Android,
	IPhone,
	Mac,
	Linux,
	Windows,
}
var AgentList = []string{
	MicroMessenger,
	Alipay,
	Edg,
	Chrome,
	Safari,
	Firefox,
	"Go-http-client",
}
