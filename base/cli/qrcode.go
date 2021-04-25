package cli

import (
	"bytes"
	"encoding/base64"
	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/skip2/go-qrcode"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
)

func _qrcode_cli(m *ice.Message, arg ...string) {
	fg := qrcodeTerminal.ConsoleColors.BrightBlue
	bg := qrcodeTerminal.ConsoleColors.BrightWhite
	switch m.Option("fg") {
	case "black":
		fg = qrcodeTerminal.ConsoleColors.BrightBlack
	case "red":
		fg = qrcodeTerminal.ConsoleColors.BrightRed
	case "green":
		fg = qrcodeTerminal.ConsoleColors.BrightGreen
	case "yellow":
		fg = qrcodeTerminal.ConsoleColors.BrightYellow
	case "blue":
		fg = qrcodeTerminal.ConsoleColors.BrightBlue
	case "cyan":
		fg = qrcodeTerminal.ConsoleColors.BrightCyan
	case "magenta":
		fg = qrcodeTerminal.ConsoleColors.BrightMagenta
	case "white":
		fg = qrcodeTerminal.ConsoleColors.BrightWhite
	}
	obj := qrcodeTerminal.New2(fg, bg, qrcodeTerminal.QRCodeRecoveryLevels.Medium)
	m.Echo("%s", *obj.Get(arg[0]))
}
func _qrcode_web(m *ice.Message, arg ...string) {
	buf := bytes.NewBuffer(make([]byte, 0, ice.MOD_BUFS))
	if qr, e := qrcode.New(arg[0], qrcode.Medium); m.Assert(e) {
		m.Assert(qr.Write(kit.Int(kit.Select("240", arg, 1)), buf))
	}
	src := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	m.Echo(`<img src="%s" title='%s' height=%s>`, src, arg[0], kit.Select("240", arg, 1))
}

const QRCODE = "qrcode"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			QRCODE: {Name: "qrcode", Help: "二维码", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			QRCODE: {Name: "qrcode text auto", Help: "二维码", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if aaa.SessIsCli(m) {
					_qrcode_cli(m, arg...)
					return
				}
				_qrcode_web(m, arg...)
			}},
		},
	})
}
