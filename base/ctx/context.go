package ctx

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _context_list(m *ice.Message, sub *ice.Context, name string) {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if !strings.HasPrefix(s.Cap(ice.CTX_FOLLOW), name+ice.PT) {
			return
		}
		m.Push(kit.MDB_NAME, strings.TrimPrefix(s.Cap(ice.CTX_FOLLOW), name+ice.PT))
		m.Push(kit.MDB_STATUS, s.Cap(ice.CTX_STATUS))
		m.Push(kit.MDB_STREAM, s.Cap(ice.CTX_STREAM))
		m.Push(kit.MDB_HELP, s.Help)
	})
}

const CONTEXT = "context"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CONTEXT: {Name: "context name=web.chat action=context,command,config key auto spide", Help: "模块", Action: ice.MergeAction(map[string]*ice.Action{
			"spide": {Name: "spide", Help: "架构图", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[1] == CONTEXT { // 模块列表
					m.Cmdy(CONTEXT, kit.Select(ice.ICE, arg, 0), CONTEXT)
					m.Display("/plugin/story/spide.js", "root", kit.Select(ice.ICE, arg, 0),
						"field", "name", "split", ice.PT, "prefix", "spide")
					return
				}
				if index := kit.Keys(arg[0], arg[1]); strings.HasSuffix(index, arg[2]) { // 命令列表
					m.Cmdy(CONTEXT, index, COMMAND).Table(func(i int, value map[string]string, head []string) {
						m.Push("file", arg[1])
					})
				} else { // 命令详情
					m.Cmdy(COMMAND, kit.Keys(index, strings.Split(arg[2], " ")[0]))
				}
			}},
		}, CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Search(kit.Select(ice.ICE, arg, 0)+ice.PT, func(p *ice.Context, s *ice.Context, key string) {
				msg := m.Spawn(s)
				defer m.Copy(msg)

				switch kit.Select(CONTEXT, arg, 1) {
				case CONTEXT:
					_context_list(msg, s, kit.Select("", arg, 0))
				case COMMAND:
					msg.Cmdy(COMMAND, arg[2:])
				case CONFIG:
					msg.Cmdy(CONFIG, arg[2:])
				}
			})
		}},
	}})
}
