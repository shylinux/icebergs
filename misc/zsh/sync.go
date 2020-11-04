package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"strings"
)

const (
	SHELL = "shell"
)
const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,id,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			SYNC: {Name: "sync id auto export import", Help: "同步流", Action: map[string]*ice.Action{
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_TOPIC:
						m.Cmdy(m.Prefix(FAVOR)).Appendv(ice.MSG_APPEND, kit.MDB_TOPIC, kit.MDB_COUNT, kit.MDB_TIME)
					}
				}},
				FAVOR: {Name: "favor topic type name text", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(m.Prefix(FAVOR), mdb.INSERT, kit.MDB_TOPIC, m.Option(kit.MDB_TOPIC),
						kit.MDB_TYPE, m.Option(kit.MDB_TYPE), kit.MDB_NAME, m.Option(kit.MDB_NAME), kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
					m.Option(mdb.CACHE_FILED, kit.MDB_ID)
					m.Option(mdb.CACHE_VALUE, arg[0])
				} else {
					m.Option(mdb.FIELDS, m.Conf(SYNC, kit.META_FIELD))
					m.Option(ice.MSG_CONTROL, ice.CONTROL_PAGE)
					defer m.PushAction(FAVOR)
				}

				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, m.Option(mdb.CACHE_FILED), m.Option(mdb.CACHE_VALUE))
			}},
			"/sync": {Name: "/sync", Help: "同步", Action: map[string]*ice.Action{
				"history": {Name: "history", Help: "历史", Hand: func(m *ice.Message, arg ...string) {
					ls := strings.SplitN(strings.TrimSpace(m.Option(ARG)), " ", 4)
					if text := strings.TrimSpace(strings.Join(ls[3:], " ")); text != "" {
						m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TYPE, SHELL, kit.MDB_NAME, ls[0],
							tcp.HOSTNAME, m.Option(tcp.HOSTNAME), aaa.USERNAME, m.Option(aaa.USERNAME),
							kit.MDB_TEXT, text, PWD, m.Option(PWD), kit.MDB_TIME, ls[1]+" "+ls[2])

					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.HASH, kit.MDB_TYPE, SHELL, kit.MDB_NAME, arg[0],
					kit.MDB_TEXT, m.Option(SUB), PWD, m.Option(PWD))
			}},
		},
	})
}
