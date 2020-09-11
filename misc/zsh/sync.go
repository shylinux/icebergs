package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"strings"
)

const SYNC = "sync"
const SHELL = "shell"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(
				kit.MDB_FIELD, "time,id,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			web.LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if f, _, e := m.R.FormFile("sub"); e == nil {
					defer f.Close()
					// 文件参数
					if b, e := ioutil.ReadAll(f); e == nil {
						m.Option("sub", string(b))
					}
				}

				m.Option("you", m.Conf("zsh", "meta.proxy"))
				m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
					// 查找空间
					m.Option("you", kit.Select(m.Conf("zsh", "meta.proxy"), value["you"]))
				})

				m.Option("arg", strings.ReplaceAll(m.Option("arg"), "%20", " "))
				m.Logs(ice.LOG_AUTH, "you", m.Option("you"), "url", m.Option(ice.MSG_USERURL), "cmd", m.Optionv("cmds"), "sub", m.Optionv("sub"))
				m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
			}},

			SYNC: {Name: "sync id=auto auto 导出 导入", Help: "同步流", Action: map[string]*ice.Action{
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select(m.Conf(SYNC, kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_ID, arg)
			}},
			"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "history":
					vs := strings.SplitN(strings.TrimSpace(m.Option("arg")), " ", 4)
					if strings.Contains(m.Option("SHELL"), "zsh") {
						vs = []string{vs[0], m.Time("2006-01-02"), m.Time("15:04:05"), strings.Join(vs[1:], " ")}
					}
					m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TYPE, SHELL, kit.MDB_NAME, vs[0],
						kit.MDB_TEXT, kit.Select("", vs, 3), "pwd", m.Option("pwd"), kit.MDB_TIME, vs[1]+" "+vs[2])

				default:
					m.Richs("login", nil, m.Option("sid"), func(key string, value map[string]interface{}) {
						kit.Value(value, kit.Keys("sync", arg[0]), kit.Dict(
							"time", m.Time(), "text", m.Option("sub"),
							"pwd", m.Option("pwd"), "cmd", arg[1:],
						))
					})
				}
			}},
		},
	}, nil)
}
