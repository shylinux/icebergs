package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const (
	PASTE = "paste"
)

func init() {
	Index.Register(&ice.Context{Name: PASTE, Help: "粘贴板",
		Configs: map[string]*ice.Config{
			PASTE: {Name: "paste", Help: "粘贴板", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Load(PASTE)
				m.Cmd(mdb.SEARCH, mdb.CREATE, PASTE, m.Prefix(PASTE))
				m.Cmd(mdb.RENDER, mdb.CREATE, PASTE, m.Prefix(PASTE))
			}},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Save(PASTE)
			}},

			PASTE: {Name: "paste text auto 添加:button", Help: "粘贴板", Action: map[string]*ice.Action{
				mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
					m.Conf(PASTE, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), kit.MDB_META, kit.MDB_SHORT), kit.MDB_TEXT)

					h := m.Rich(PASTE, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), kit.Dict(
						kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, strings.TrimSpace(arg[2]),
					))
					m.Log_INSERT(PASTE, arg[2])
					m.Echo(h)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(PASTE, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), m.Option(kit.MDB_TEXT), func(key string, value map[string]interface{}) {
						if arg[0] == kit.MDB_TEXT {
							return
						}
						m.Log_MODIFY(PASTE, m.Option(kit.MDB_TEXT))
						value[arg[0]] = arg[1]
					})
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(PASTE, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), m.Option(kit.MDB_TEXT), func(key string, value map[string]interface{}) {
						m.Log_REMOVE(PASTE, m.Option(kit.MDB_TEXT))
						m.Conf(PASTE, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), kit.MDB_HASH, key), "")
					})
				}},
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(PASTE, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						if strings.Contains(kit.Format(value[kit.MDB_NAME]), arg[1]) ||
							strings.Contains(kit.Format(value[kit.MDB_TEXT]), arg[1]) {

							m.Push("pod", m.Option("pod"))
							m.Push("ctx", m.Cap(ice.CTX_FOLLOW))
							m.Push("cmd", PASTE)
							m.Push(kit.MDB_TIME, value["time"])
							m.Push(kit.MDB_SIZE, value["size"])
							m.Push(kit.MDB_TYPE, PASTE)
							m.Push(kit.MDB_NAME, value["name"])
							m.Push(kit.MDB_TEXT, value["text"])
						}
					})
				}},
				mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(arg[2])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(PASTE, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
					if len(arg) == 0 {
						m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT})
						if strings.Contains(m.Option(ice.MSG_USERUA), "MicroMessenger") {
							return
						}
						m.Push("action", m.Cmdx(mdb.RENDER, web.RENDER.Button, "删除"))
						return
					}
					m.Push("detail", value)
				})
				m.Sort("time", "time_r")
			}},
		},
	}, nil)
}
