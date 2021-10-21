package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CASE = "case"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CASE: {Name: CASE, Help: "用例", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,name,cmd,api,arg,res",
		)},
	}, Commands: map[string]*ice.Command{
		CASE: {Name: "case dev zone id auto", Help: "用例", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create name address", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.SPIDE, mdb.CREATE, arg)
			}},
			mdb.INSERT: {Name: "insert zone name=hi cmd=POST,GET api arg:textarea res:textarea", Help: "添加"},

			cli.CHECK: {Name: "check", Help: "检查", Hand: func(m *ice.Message, arg ...string) {
				if m.ProcessInner(); len(arg) > 0 {
					success := 0
					m.Cmd(m.PrefixKey(), arg[0]).Table(func(index int, value map[string]string, head []string) {
						m.Push(kit.MDB_TIME, m.Time())
						m.Push(kit.MDB_ID, value[kit.MDB_ID])
						if err := m.Cmdx(m.PrefixKey(), cli.CHECK, value); err == ice.OK {
							m.Push(ice.ERR, cli.Color(m, cli.GREEN, err))
							success++
						} else {
							m.Push(ice.ERR, cli.Color(m, cli.RED, err))
						}
						m.Push(cli.API, value[cli.API])
						m.Push(ice.ARG, value[ice.ARG])
						m.Push(ice.RES, value[ice.RES])
					})
					m.StatusTimeCount(ice.SUCCESS, success)
					return
				}

				res := kit.UnMarshal(m.Cmdx(m.PrefixKey(), ice.RUN))
				if m.Option(ice.RES) != "" {
					for k, v := range kit.KeyValue(nil, "", kit.UnMarshal(m.Option(ice.RES))) {
						if v != kit.Value(res, k) {
							m.Echo(kit.Formats(res))
							return
						}
					}
				}
				m.Echo(ice.OK)
			}},
			ice.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.ContentType, web.ContentJSON)
				m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(web.SPIDE, m.Option(ice.DEV), web.SPIDE_RAW,
					m.Option(ice.CMD), m.Option(cli.API), web.SPIDE_DATA, m.Option(ice.ARG)))))
				m.Info(`curl "` + m.Option(cli.API) + `" -H "Content-Type: application/json"` + ` -d '` + m.Option(ice.ARG) + `'`)
				m.ProcessDisplay("/plugin/local/wiki/json.js")
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(web.SPIDE)
				m.Action(mdb.CREATE)
				m.RenameAppend("client.name", "dev")
				m.RenameAppend("client.url", "address")
				return
			}

			if mdb.ZoneSelect(m, arg[1:]...); len(arg) == 1 {
				m.Action(mdb.INSERT, mdb.EXPORT, mdb.IMPORT)
				m.PushAction(mdb.INSERT, cli.CHECK, mdb.REMOVE)
			} else {
				m.Action(mdb.INSERT, cli.CHECK)
				m.PushAction(ice.RUN, cli.CHECK)
			}
			m.StatusTimeCount()
		}},
	}})
}
