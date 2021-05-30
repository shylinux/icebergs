package cli

import (
	"encoding/base64"
	"github.com/skip2/go-qrcode"
	"image/color"
	"math/rand"
	"strconv"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
)

var _trans_web = map[string]color.Color{
	"black":   color.RGBA{0, 0, 0, 255},
	"red":     color.RGBA{255, 0, 0, 255},
	"green":   color.RGBA{0, 255, 0, 255},
	"yellow":  color.RGBA{255, 255, 0, 255},
	"blue":    color.RGBA{0, 0, 255, 255},
	"magenta": color.RGBA{255, 0, 255, 255},
	"cyan":    color.RGBA{0, 255, 255, 255},
	"white":   color.RGBA{255, 255, 255, 255},
}

func _parse_color(str string) color.Color {
	if str == "random" {
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
	if r > 128 {
		res += 1
	}
	if g > 128 {
		res += 2
	}
	if b > 128 {
		res += 4
	}
	return kit.Format(res)
}

func _qrcode_cli(m *ice.Message, text string, arg ...string) {
	qr, _ := qrcode.New(text, qrcode.Medium)
	fg := _trans_cli(m.Option("fg"))
	bg := _trans_cli(m.Option("bg"))

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
	qr.ForegroundColor = _parse_color(m.Option("fg"))
	qr.BackgroundColor = _parse_color(m.Option("bg"))

	if data, err := qr.PNG(kit.Int(m.Option("size"))); m.Assert(err) {
		m.Echo(`<img src="data:image/png;base64,%s" title='%s'>`, base64.StdEncoding.EncodeToString(data), text)
	}
}

const QRCODE = "qrcode"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			QRCODE: {Name: "qrcode", Help: "二维码", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			QRCODE: {Name: "qrcode text fg bg size auto", Help: "二维码", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option("fg", kit.Select("blue", arg, 1))
				m.Option("bg", kit.Select("white", arg, 2))
				m.Option("size", kit.Select("240", arg, 3))

				if aaa.SessIsCli(m) {
					_qrcode_cli(m, arg[0])
				} else {
					_qrcode_web(m, arg[0])
				}
			}},
		},
	})
}
