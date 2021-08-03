package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const CASE = "case"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CASE: {Name: CASE, Help: "用例", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,name,cmd,api,arg,res",
			)},
		},
		Commands: map[string]*ice.Command{
			CASE: {Name: "case zone id auto", Help: "用例", Action: ice.MergeAction(map[string]*ice.Action{
				mdb.INSERT: {Name: "create zone name=hi cmd=POST,GET api arg:textarea res:textarea", Help: "添加"},

				cli.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Option(web.SPIDE_HEADER, web.ContentType, web.ContentJSON)
					m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(web.SPIDE, web.SPIDE_DEV, web.SPIDE_RAW,
						m.Option(cli.CMD), m.Option(cli.API), web.SPIDE_DATA, m.Option(cli.ARG)))))
					m.Info(`curl "` + m.Option(cli.API) + `" -H "Content-Type: application/json"` + ` -d '` + m.Option(cli.ARG) + `'`)
					m.ProcessInner()
				}},
				cli.CHECK: {Name: "check", Help: "检查", Hand: func(m *ice.Message, arg ...string) {
					if m.ProcessInner(); len(arg) > 0 {
						success := 0
						m.Cmd(m.PrefixKey(), arg[0]).Table(func(index int, value map[string]string, head []string) {
							m.Push(kit.MDB_TIME, m.Time())
							m.Push(kit.MDB_ID, value[kit.MDB_ID])
							if err := m.Cmdx(m.PrefixKey(), cli.CHECK, value); err == ice.OK {
								m.Push(cli.ERR, cli.Color(m, cli.GREEN, err))
								success++
							} else {
								m.Push(cli.ERR, cli.Color(m, cli.RED, err))
							}
							m.Push(cli.API, value[cli.API])
							m.Push(cli.ARG, value[cli.ARG])
							m.Push(cli.RES, value[cli.RES])
						})
						m.StatusTimeCount(ice.SUCCESS, success)
						return
					}

					res := kit.UnMarshal(m.Cmdx(m.PrefixKey(), cli.RUN))
					if m.Option(cli.RES) != "" {
						for k, v := range kit.KeyValue(nil, "", kit.UnMarshal(m.Option(cli.RES))) {
							if v != kit.Value(res, k) {
								m.Echo(kit.Formats(res))
								return
							}
						}
					}
					m.Echo(ice.OK)
				}},
			}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,zone,count", m.Conf(m.PrefixKey(), kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.ZONE, arg); len(arg) == 0 {
					m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
					m.PushAction(mdb.INSERT, cli.CHECK, mdb.REMOVE)
				} else {
					m.Action(mdb.INSERT, cli.CHECK)
					m.PushAction(cli.RUN, cli.CHECK)
				}
				m.StatusTimeCount()
			}},
		},
	})

}
