package cli

import (
	"encoding/base64"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/misc/qrcode"
	kit "shylinux.com/x/toolkits"
)

func _qrcode_cli(m *ice.Message, text string) {
	sc := qrcode.New(text)
	fg := ParseCliColor(m.Option(FG))
	bg := ParseCliColor(m.Option(BG))
	data := sc.Bitmap()
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
		m.Echo(lex.NL)
	}
	m.Echo(text).Echo(lex.NL)
}
func _qrcode_web(m *ice.Message, text string) string {
	sc := qrcode.New(text)
	sc.ForegroundColor = ParseColor(m.Option(FG))
	sc.BackgroundColor = ParseColor(m.Option(BG))
	if data, err := sc.PNG(kit.Int(m.Option(SIZE))); m.Assert(err) {
		m.Echo(`<img src="data:image/png;base64,%s" title='%s'>`, base64.StdEncoding.EncodeToString(data), text)
	}
	return text
}

const (
	SIZE = "size"
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
					m.Push(arg[0], BLACK, WHITE)
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			switch m.Option(ice.MSG_THEME) {
			case LIGHT, WHITE:
				m.Option(FG, kit.Select(BLACK, arg, 1))
				m.Option(BG, kit.Select(WHITE, arg, 2))
			default:
				m.Option(FG, kit.Select(kit.Select(BLACK, m.Option("--plugin-fg-color")), arg, 1))
				m.Option(BG, kit.Select(kit.Select(WHITE, m.Option("--plugin-bg-color")), arg, 2))
			}
			if m.IsCliUA() {
				_qrcode_cli(m, kit.Select(kit.Select(ice.Info.Make.Domain, ice.Info.Domain), arg, 0))
			} else {
				m.OptionDefault(SIZE, "360")
				m.StatusTime(mdb.LINK, _qrcode_web(m, tcp.PublishLocalhost(m, kit.Select(m.Option(ice.MSG_USERWEB), arg, 0))))
			}
		}},
	})
}
