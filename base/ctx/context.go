package ctx

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _context_list(m *ice.Message, sub *ice.Context, name string) {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if name != "" && name != ice.ICE && !strings.HasPrefix(s.Cap(ice.CTX_FOLLOW), name+ice.PT) {
			return
		}
		m.Push(mdb.NAME, s.Cap(ice.CTX_FOLLOW))
		m.Push(mdb.STATUS, s.Cap(ice.CTX_STATUS))
		m.Push(mdb.STREAM, s.Cap(ice.CTX_STREAM))
		m.Push(mdb.HELP, s.Help)
	})
}
func Inputs(m *ice.Message, field string) bool {
	switch strings.TrimPrefix(field, "extra.") {
	case ice.POD:
		m.Cmdy("route")
	case ice.CTX:
		m.Cmdy(CONTEXT)
	case ice.CMD:
		m.Cmdy(CONTEXT, kit.Select(m.Option(ice.CTX), m.Option(kit.Keys(mdb.EXTRA, ice.CTX))), COMMAND)
	case ice.ARG:

	default:
		return false
	}
	return true
}

const CONTEXT = "context"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CONTEXT: {Name: "context name=web action=context,command,config key auto spide", Help: "模块", Action: ice.MergeAction(map[string]*ice.Action{
			"spide": {Name: "spide", Help: "架构图", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[1] == CONTEXT { // 模块列表
					m.Cmdy(CONTEXT, kit.Select(ice.ICE, arg, 0), CONTEXT)
					m.Display("/plugin/story/spide.js?prefix=spide", "root", kit.Select(ice.ICE, arg, 0), "split", ice.PT)

				} else if index := kit.Keys(arg[1]); strings.HasSuffix(index, arg[2]) { // 命令列表
					m.Cmdy(CONTEXT, index, COMMAND).Table(func(i int, value map[string]string, head []string) {
						m.Push("file", arg[1])
					})

				} else { // 命令详情
					m.Cmdy(COMMAND, kit.Keys(index, strings.Split(arg[2], ice.SP)[0]))
				}
			}},
		}, CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, m.Source().Cap(ice.CTX_FOLLOW))
			}
			m.Search(arg[0]+ice.PT, func(p *ice.Context, s *ice.Context, key string) {
				msg := m.Spawn(s)
				defer m.Copy(msg)

				switch kit.Select(CONTEXT, arg, 1) {
				case CONTEXT:
					_context_list(msg, s, arg[0])
				case COMMAND:
					msg.Cmdy(COMMAND, arg[2:])
				case CONFIG:
					msg.Cmdy(CONFIG, arg[2:])
				}
			})
		}},
	}})
}
