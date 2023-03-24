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
		m.Push(mdb.HELP, s.Help)
	})
}

const CONTEXT = "context"

func init() {
	Index.MergeCommands(ice.Commands{
		CONTEXT: {Name: "context name=web action=context,command,config key auto", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, m.Source().Cap(ice.CTX_FOLLOW))
			}
			m.Search(arg[0]+ice.PT, func(p *ice.Context, s *ice.Context) {
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
	})
}
func Inputs(m *ice.Message, field string) bool {
	switch strings.TrimPrefix(field, "extra.") {
	case ice.POD:
		m.Cmdy(ice.SPACE)
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
