package wiki

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"path"
)

const (
	DRAW = "draw"
)

func init() {
	Index.Merge(&ice.Context{
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
					m.Cmd("nfs.save", path.Join(m.Conf(DRAW, "meta.path"), kit.Select("hi.svg", arg[0])), arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				reply(m, cmd, arg...)
			}},
		},
	}, nil)
}
