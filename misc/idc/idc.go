package idc

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "idc", Help: "idc",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"conf": {Name: "conf", Help: "conf", Value: kit.Data(kit.MDB_SHORT, "name")},
		"show": {Name: "show", Help: "show", Value: kit.Data(kit.MDB_SHORT, "show")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("conf", "show")
		}},

		"show": {Name: "show key type name text", Help: "show", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Show(cmd, arg...) {
				return
			}

			m.Optionv("header", "Content-Type", "application/json")
			m.Cmdy(ice.WEB_SPIDE, "dev", "msg", "POST", "/code/idc/show", "data", kit.Format(kit.Dict("cmds", append([]string{}, arg...))))
		}},

		"/show": {Name: "/show key type name text", Help: "show", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Richs("show", nil, arg[0], nil) == nil {
				m.Rich("show", nil, kit.Data("show", arg[0]))
			}

			m.Richs("show", nil, arg[0], func(key string, val map[string]interface{}) {
				m.Grow("show", kit.Keys(kit.MDB_HASH, key), kit.Dict(
					kit.MDB_TYPE, arg[1], kit.MDB_NAME, arg[2], kit.MDB_TEXT, arg[3],
					kit.MDB_EXTRA, kit.Dict(arg[4:]),
				))
			})
			m.Echo("hello world")
		}},
		"/conf": {Name: "conf key field", Help: "conf", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo(m.Conf("conf", arg[0], arg[1]))
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
