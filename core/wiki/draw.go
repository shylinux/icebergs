package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
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

const DRAW = "draw"
const (
	DrawPlugin = "/plugin/local/wiki/draw.js"
)

func init() {
	sub := Index.Register(&ice.Context{Name: "draw", Help: "思维导图",
		Configs: map[string]*ice.Config{
			DRAW: {Name: "draw", Help: "思维导图", Value: kit.Data(
				"path", "", "regs", ".*\\.svg",
				"prefix", `<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg" width="%v" height="%v">`, "suffix", `</svg>`,
			)},
		},
		Commands: map[string]*ice.Command{
			DRAW: {Name: "draw path=src/main.svg auto", Help: "思维导图", Meta: kit.Dict(mdb.PLUGIN, DrawPlugin), Action: map[string]*ice.Action{
				nfs.SAVE: {Name: "save path text", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_wiki_save(m, DRAW, arg[0], kit.Select(m.Option("content"), arg, 1))
				}},
				"run": {Name: "show zone type name text", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					_draw_show(m, arg[0], arg[1], arg[2], arg[3], arg[4:]...)
				}},
				mdb.PLUGIN: {Name: "plugin", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
					_draw_plugin(m, arg...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if !_wiki_list(m, DRAW, kit.Select("./", arg, 0)) {
					_wiki_show(m, DRAW, arg[0])
				}
			}},
		},
	}, nil)

	sub.Register(&ice.Context{Name: "工作", Help: "工作",
		Commands: map[string]*ice.Command{
			"项目开发": {Name: "项目开发", Help: "项目开发", Action: map[string]*ice.Action{
				"run": {Hand: func(m *ice.Message, arg ...string) {
					m.Echo("hello world")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				m.Echo("hello world")
			}},
			"项目测试": {Name: "项目测试", Help: "项目测试", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {

			}},
		},
	}, &web.Frame{})

}
