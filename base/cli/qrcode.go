package cli

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"strings"

	"shylinux.com/x/go-qrcode"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

var _color_map = map[string]color.Color{
	BLACK:  color.RGBA{0, 0, 0, DARK},
	RED:    color.RGBA{DARK, 0, 0, DARK},
	GREEN:  color.RGBA{0, DARK, 0, DARK},
	YELLOW: color.RGBA{DARK, DARK, 0, DARK},
	BLUE:   color.RGBA{0, 0, DARK, DARK},
	PURPLE: color.RGBA{DARK, 0, DARK, DARK},
	CYAN:   color.RGBA{0, DARK, DARK, DARK},
	WHITE:  color.RGBA{DARK, DARK, DARK, DARK},
}

func _parse_color(str string) color.Color {
	if str == RANDOM {
		list := kit.SortedKey(_color_map)
		str = list[rand.Intn(len(list))]
	}
	if strings.HasPrefix(str, "#") {
		if len(str) == 7 {
			str += "ff"
		}
		if u, e := strconv.ParseUint(str[1:], 16, 64); e == nil {
			return color.RGBA{
				uint8((u & 0xFF000000) >> 24),
				uint8((u & 0x00FF0000) >> 16),
				uint8((u & 0x0000FF00) >> 8),
				uint8((u & 0x000000FF) >> 0),
			}
		}
	}
	return _color_map[str]
}
func _parse_cli_color(str string) string {
	res := 0
	r, g, b, _ := _parse_color(str).RGBA()
	if r > LIGHT {
		res += 1
	}
	if g > LIGHT {
		res += 2
	}
	if b > LIGHT {
		res += 4
	}
	return kit.Format(res)
}
func _qrcode_cli(m *ice.Message, text string) {
	qr, _ := qrcode.New(text, qrcode.Medium)
	fg := _parse_cli_color(m.Option(FG))
	bg := _parse_cli_color(m.Option(BG))
	data := qr.Bitmap()
	for i, row := range data {
		if n := len(data); i < 3 || i >= n-3 {
			continue
		}
		for i, col := range row {
			if n := len(row); i < 3 || i >= n-3 {
				continue
			}
			m.Echo("\033[4%sm  \033[0m", kit.Select(bg, fg, col))
		}
		m.Echo(ice.NL)
	}
	m.Echo(text).Echo(ice.NL)
}
func _qrcode_web(m *ice.Message, text string) {
	qr, _ := qrcode.New(text, qrcode.Medium)
	qr.ForegroundColor = _parse_color(m.Option(FG))
	qr.BackgroundColor = _parse_color(m.Option(BG))
	if data, err := qr.PNG(kit.Int(m.Option(SIZE))); m.Assert(err) {
		m.Echo(`<img src="data:image/png;base64,%s" title='%s'>`, base64.StdEncoding.EncodeToString(data), text)
	}
}

const (
	BG    = "bg"
	FG    = "fg"
	DARK  = 255
	LIGHT = 127
	SIZE  = "size"
)
const (
	COLOR  = "color"
	BLACK  = "black"
	WHITE  = "white"
	BLUE   = "blue"
	RED    = "red"
	GRAY   = "gray"
	CYAN   = "cyan"
	GREEN  = "green"
	PURPLE = "purple"
	YELLOW = "yellow"
	RANDOM = "random"
	GLASS  = "#0000"
)
const QRCODE = "qrcode"

func init() {
	Index.MergeCommands(ice.Commands{
		QRCODE: {Name: "qrcode text fg@key bg@key size auto", Help: "二维码", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.AddRender(ice.RENDER_QRCODE, func(m *ice.Message, args ...ice.Any) string {
					return m.Cmd(QRCODE, kit.Simple(args...)).Result()
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case FG, BG:
					m.Push(arg[0], BLACK, WHITE, BLUE, RED, CYAN, GREEN, PURPLE, YELLOW)
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Option(FG, kit.Select(kit.Select(BLACK, WHITE, m.Option(ice.THEME) == BLACK || m.Option(ice.THEME) == "dark"), arg, 1))
			m.Option(BG, kit.Select(kit.Select(WHITE, BLACK, m.Option(ice.THEME) == BLACK || m.Option(ice.THEME) == "dark"), arg, 2))
			if m.IsCliUA() {
				_qrcode_cli(m, kit.Select(kit.Select(ice.Info.Make.Domain, ice.Info.Domain), arg, 0))
			} else {
				m.Option(SIZE, kit.Select(kit.Format(kit.Max(240, kit.Min(480, kit.Int(m.Option(ice.MSG_HEIGHT)), kit.Int(m.Option(ice.MSG_WIDTH))))), arg, 3))
				_qrcode_web(m, kit.Select(m.Option(ice.MSG_USERWEB), arg, 0))
				m.StatusTime(mdb.LINK, kit.Select(m.Option(ice.MSG_USERWEB), arg, 0))
			}
		}},
	})
}

func Color(m *ice.Message, c string, str ice.Any) string {
	wrap, color := `<span style="color:%s">%v</span>`, c
	if m.IsCliUA() {
		wrap, color = "\033[3%sm%v\033[0m", _parse_cli_color(c)
	}
	return fmt.Sprintf(wrap, color, str)
}
func ColorRed(m *ice.Message, str ice.Any) string    { return Color(m, RED, str) }
func ColorGreen(m *ice.Message, str ice.Any) string  { return Color(m, GREEN, str) }
func ColorYellow(m *ice.Message, str ice.Any) string { return Color(m, YELLOW, str) }

func PushText(m *ice.Message, text string) {
	m.OptionFields(ice.MSG_DETAIL)
	if m.PushScript(nfs.SCRIPT, text); strings.HasPrefix(text, ice.HTTP) {
		m.PushQRCode(QRCODE, text)
		m.PushAnchor(text)
	}
	m.Echo(text)
}
