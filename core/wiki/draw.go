package wiki

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/toolkits"

	"path"
)

const (
	DRAW = "draw"
)

func _draw_show(m *ice.Message, zone, kind, name, text string, arg ...string) {
	m.Cmdy(kit.Keys(zone, kind), name, text, arg)
}
func _draw_plugin(m *ice.Message, arg ...string) {
	m.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if m.Target() == s {
			return
		}
		m.Push(s.Name, cmd.Name)
	})
}

func init() {
	sub := &ice.Context{Name: "draw", Help: "思维导图",
		Configs: map[string]*ice.Config{
			DRAW: {Name: "draw", Help: "思维导图", Value: kit.Data(kit.MDB_SHORT, "name", "path", "", "regs", ".*\\.svg",
				"prefix", `<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg" width="%v" height="%v">`, "suffix", `</svg>`,
			)},
		},
		Commands: map[string]*ice.Command{
			DRAW: {Name: "draw path=自然/编程/hi.svg auto", Help: "思维导图", Meta: kit.Dict(
				"display", "/plugin/local/wiki/draw.js",
			), Action: map[string]*ice.Action{
				"save": {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					nfs.FileSave(m, path.Join(m.Conf(DRAW, "meta.path"), kit.Select("hi.svg", arg, 0)), kit.Select(m.Option("content"), arg, 1))
				}},
				"run": {Name: "show path text", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					_draw_show(m, arg[0], arg[1], arg[2], arg[3], arg[4:]...)
				}},
				"plugin": {Name: "plugin", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_draw_plugin(m, arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				reply(m, cmd, arg...)
			}},
		},
	}

	sub.Register(&ice.Context{Name: "创业", Help: "创业",
		Commands: map[string]*ice.Command{
			"项目开发": {Name: "项目开发", Help: "项目开发", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				m.Echo("hello world")
			}},
			"项目测试": {Name: "项目测试", Help: "项目测试", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {

			}},
			"改变世界": {Name: "改变世界", Help: "改变世界", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				m.Echo("hello world")
			}},
			"认识世界": {Name: "认识世界", Help: "认识世界", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				m.Echo("hello world")
			}},
		},
	}, nil)

	Index.Register(sub, nil)
}
