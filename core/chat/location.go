package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"net/url"
	"strings"
)

const (
	LOCATION  = "location"
	LONGITUDE = "longitude"
	LATITUDE  = "latitude"
)

func init() {
	Index.Register(&ice.Context{Name: LOCATION, Help: "地理位置",
		Configs: map[string]*ice.Config{
			LOCATION: {Name: "location", Help: "地理位置", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Load(LOCATION)
				m.Cmd(mdb.SEARCH, mdb.CREATE, LOCATION, m.Prefix(LOCATION))
				m.Cmd(mdb.RENDER, mdb.CREATE, LOCATION, m.Prefix(LOCATION))
			}},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save(LOCATION) }},

			LOCATION: {Name: "location text auto 添加:button", Help: "地理位置", Action: map[string]*ice.Action{
				mdb.INSERT: {Hand: func(m *ice.Message, arg ...string) {
					m.Conf(LOCATION, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), kit.MDB_META, kit.MDB_SHORT), kit.MDB_TEXT)

					h := m.Rich(LOCATION, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), kit.Dict(
						kit.MDB_TYPE, arg[0], kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2],
						LONGITUDE, arg[3], LATITUDE, arg[4],
					))
					m.Log_INSERT(LOCATION, arg[2])
					m.Echo(h)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(LOCATION, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), m.Option(kit.MDB_TEXT), func(key string, value map[string]interface{}) {
						if arg[0] == kit.MDB_TEXT {
							return
						}
						m.Log_MODIFY(PASTE, m.Option(kit.MDB_TEXT))
						value[arg[0]] = arg[1]
					})
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(LOCATION, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), m.Option(kit.MDB_TEXT), func(key string, value map[string]interface{}) {
						m.Log_REMOVE(LOCATION, m.Option(kit.MDB_TEXT))
						m.Conf(LOCATION, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM), kit.MDB_HASH, key), "")
					})
				}},
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					m.Richs(LOCATION, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
						if strings.Contains(kit.Format(value[kit.MDB_NAME]), arg[1]) ||
							strings.Contains(kit.Format(value[kit.MDB_TEXT]), arg[1]) {

							m.Push("pod", m.Option("pod"))
							m.Push("ctx", m.Cap(ice.CTX_FOLLOW))
							m.Push("cmd", LOCATION)
							m.Push(kit.MDB_TIME, value["time"])
							m.Push(kit.MDB_SIZE, value["size"])
							m.Push(kit.MDB_TYPE, LOCATION)
							m.Push(kit.MDB_NAME, value["name"])
							m.Push(kit.MDB_TEXT, value["text"])
						}
					})
				}},
				mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.RENDER, web.RENDER.Frame, kit.Format(
						"https://map.baidu.com/search/%s/@12958750.085,4825785.55,16z?querytype=s&da_src=shareurl&wd=%s",
						arg[2], arg[2]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(LOCATION, kit.Keys(kit.MDB_META, m.Option(ice.MSG_RIVER), m.Option(ice.MSG_STORM)), kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
					if len(arg) == 0 {
						m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_TEXT,
							LONGITUDE, LATITUDE})
						if strings.Contains(m.Option(ice.MSG_USERUA), "MicroMessenger") {
							return
						}
						m.Push("action", m.Cmdx(mdb.RENDER, web.RENDER.Button, "删除"))
						loc := m.Cmdx(mdb.RENDER, web.RENDER.A, "百度地图", kit.Format(
							"https://map.baidu.com/search/%s/@12958750.085,4825785.55,16z?querytype=s&da_src=shareurl&wd=%s",
							url.QueryEscape(kit.Format(value[kit.MDB_TEXT])),
							url.QueryEscape(kit.Format(value[kit.MDB_TEXT])),
						))
						m.Push("location", loc)
						return
					}
					m.Push("detail", value)
				})
				m.Sort("time", "time_r")
			}},
		},
	}, nil)
}
