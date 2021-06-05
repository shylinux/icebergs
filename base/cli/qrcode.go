package cli

import (
	"encoding/base64"
	"image/color"
	"math/rand"
	"strconv"
	"strings"

	"github.com/skip2/go-qrcode"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
)

var _trans_web = map[string]color.Color{
	BLACK:   color.RGBA{0, 0, 0, DARK},
	RED:     color.RGBA{DARK, 0, 0, DARK},
	YELLOW:  color.RGBA{DARK, DARK, 0, DARK},
	GREEN:   color.RGBA{0, DARK, 0, DARK},
	CYAN:    color.RGBA{0, DARK, DARK, DARK},
	BLUE:    color.RGBA{0, 0, DARK, DARK},
	MAGENTA: color.RGBA{DARK, 0, DARK, DARK},
	WHITE:   color.RGBA{DARK, DARK, DARK, DARK},
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
func _trans_cli(str string) string {
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

func _qrcode_cli(m *ice.Message, text string, arg ...string) {
	qr, _ := qrcode.New(text, qrcode.Medium)
	fg := _trans_cli(m.Option(FG))
	bg := _trans_cli(m.Option(BG))

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
		m.Echo("\n")
	}
}
func _qrcode_web(m *ice.Message, text string, arg ...string) {
	qr, _ := qrcode.New(text, qrcode.Medium)
	qr.ForegroundColor = _parse_color(m.Option(FG))
	qr.BackgroundColor = _parse_color(m.Option(BG))

	if data, err := qr.PNG(kit.Int(m.Option(SIZE))); m.Assert(err) {
		m.Echo(`<img src="data:image/png;base64,%s" title='%s'>`, base64.StdEncoding.EncodeToString(data), text)
	}
}

const (
	FG   = "fg"
	BG   = "bg"
	SIZE = "size"

	DARK  = 255
	LIGHT = 127
)
const (
	BLACK   = "black"
	RED     = "red"
	YELLOW  = "yellow"
	GREEN   = "green"
	CYAN    = "cyan"
	BLUE    = "blue"
	MAGENTA = "magenta"
	WHITE   = "white"
	RANDOM  = "random"
)
const QRCODE = "qrcode"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			QRCODE: {Name: QRCODE, Help: "二维码", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			QRCODE: {Name: "qrcode text fg bg size auto", Help: "二维码", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(SIZE, kit.Select("240", arg, 3))
				m.Option(BG, kit.Select(WHITE, arg, 2))
				m.Option(FG, kit.Select(BLUE, arg, 1))

				if aaa.SessIsCli(m) {
					_qrcode_cli(m, kit.Select(m.Conf("web.share", kit.Keym(kit.MDB_DOMAIN)), arg))
				} else {
					_qrcode_web(m, kit.Select(m.Option(ice.MSG_USERWEB), arg, 0))
				}
			}},
		},
	})
}
