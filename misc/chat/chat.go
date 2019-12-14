package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/web"
	_ "github.com/shylinux/icebergs/misc"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "chat", Help: "聊天模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"group": {Name: "group", Value: map[string]interface{}{
			"meta": map[string]interface{}{},
			"list": map[string]interface{}{},
			"hash": map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"_init": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"/ocean": {Name: "/ocean", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"/river": {Name: "/river", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				arg = kit.Simple(m.Optionv("cmds"))
			}

			if len(arg) == 0 {
				m.Confm("group", "hash", func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("create_time", value["create_time"])
					m.Push("name", value["name"])
				})
				return
			}

			switch arg[0] {
			case "create":
				// h := kit.Hashs("uniq")
				h := kit.ShortKey(m.Confm("group", "hash"), 6)
				m.Conf("group", "hash."+h, map[string]interface{}{
					"create_time": m.Time(),
					"create_name": arg[1],
				})
				m.Log("info", "river create %v %v", h, kit.Formats(m.Confv("group", "hash."+h)))
				m.Echo(h)
			}
		}},
		"/action": {Name: "/action", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if cmds, ok := m.Optionv("cmds").([]string); ok {
				m.Cmdy("web.space", cmds)
				return
			}
		}},
		"/storm": {Name: "/storm", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"/steam": {Name: "/steam", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"_exit": {Name: "_init", Help: "hello", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	},
}

func init() { web.Index.Register(Index, &web.WEB{}) }
