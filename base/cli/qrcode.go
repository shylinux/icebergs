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
	kit "shylinux.com/x/toolkits"
)

var _trans_web = map[string]color.Color{
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
		list := []string{}
		for k := range _trans_web {
			list = append(list, k)
		}
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
	return _trans_web[str]
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
	m.Echo(text)
}
func _qrcode_web(m *ice.Message, text string) {
	qr, _ := qrcode.New(text, qrcode.Medium)
	qr.ForegroundColor = _parse_color(m.Option(FG))
	qr.BackgroundColor = _parse_color(m.Option(BG))

	if data, err := qr.PNG(kit.Int(m.Option(SIZE))); m.Assert(err) {
		m.Echo(`<img src="data:image/png;base64,%s" title='%s'>`, base64.StdEncoding.EncodeToString(data), text)
	}
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

const (
	FG   = "fg"
	BG   = "bg"
	SIZE = "size"

	DARK  = 255
	LIGHT = 127
)
const (
	COLOR  = "color"
	BLACK  = "black"
	RED    = "red"
	GREEN  = "green"
	YELLOW = "yellow"
	BLUE   = "blue"
	PURPLE = "purple"
	CYAN   = "cyan"
	WHITE  = "white"
	RANDOM = "random"
	GLASS  = "#0000"
	GRAY   = "gray"
)
const QRCODE = "qrcode"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		QRCODE: {Name: "qrcode text@key fg@key bg@key size auto", Help: "二维码", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.AddRender(ice.RENDER_QRCODE, func(m *ice.Message, cmd string, args ...ice.Any) string {
					return m.Cmd(QRCODE, kit.Simple(args...)).Result()
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case "text":
					m.Push("text", "hi")
					m.Push("text", "hello")
					m.Push("text", "world")
				case "fg", "bg":
					m.Push("color", "red")
					m.Push("color", "green")
					m.Push("color", "blue")
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Option(SIZE, kit.Select("240", arg, 3))
			m.Option(BG, kit.Select(WHITE, arg, 2))
			m.Option(FG, kit.Select(BLUE, arg, 1))

			if m.IsCliUA() {
				_qrcode_cli(m, kit.Select(ice.Info.Domain, arg, 0))
			} else {
				_qrcode_web(m, kit.Select(m.Option(ice.MSG_USERWEB), arg, 0))
			}
		}},
	}})
}
