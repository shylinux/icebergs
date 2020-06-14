package code

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	qrs "github.com/skip2/go-qrcode"
	"github.com/tuotoo/qrcode"
	"os"
)

func init() {
	Index.Register(&ice.Context{Name: "qrc", Help: "二维码",
		Configs: map[string]*ice.Config{
			QRCODE: {Name: "qrcode", Help: "二维码", Value: kit.Data(
				"plug", `{"display": {"height": "400px"}}`,
			)},
		},
		Commands: map[string]*ice.Command{
			"list": {Name: "list name", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if f, e := os.Open(arg[0]); e == nil {
					defer f.Close()
					if q, e := qrcode.Decode(f); e == nil {
						m.Echo(q.Content)
						return
					}
				}
				m.Echo("hello world")
			}},
			"save": {Name: "save name text", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if qr, e := qrs.New(kit.Select(m.Option("content", arg, 1)), qrs.Medium); m.Assert(e) {
					if f, e := os.Create(arg[0]); m.Assert(e) {
						defer f.Close()
						m.Assert(qr.Write(kit.Int(kit.Select("256", arg, 2)), f))
						m.Echo(arg[0])
					}
				}
			}},
			"plug": {Name: "plug name text", Help: "插件", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(m.Conf(QRCODE, "meta.plug"))
			}},
			"show": {Name: "show name", Help: "渲染", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(`<img src="/share/local/%s">`, arg[0])
			}},
		},
	}, nil)
}
