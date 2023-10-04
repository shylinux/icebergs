package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CASE = "case"

func init() {
	Index.MergeCommands(ice.Commands{
		CASE: {Name: "case dev zone id auto insert", Help: "用例", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone*=demo name=hi cmd=GET,POST api*=/chat/cmd/web.chat.favor arg:textarea res:textarea"},
			cli.CHECK: {Help: "检查", Hand: func(m *ice.Message, arg ...string) {
				if m.ProcessInner(); len(arg) > 1 {
					success := 0
					m.Cmd("", arg[0], arg[1], func(value ice.Maps) {
						m.Push(mdb.TIME, m.Time()).Push(mdb.ID, value[mdb.ID])
						if err := m.Cmdx("", cli.CHECK, value); err == ice.OK {
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
				if res := kit.UnMarshal(m.Cmdx("", ctx.RUN)); m.Option(ice.RES) != "" {
					for k, v := range kit.KeyValue(nil, "", kit.UnMarshal(m.Option(ice.RES))) {
						if v != kit.Value(res, k) {
							m.Echo(kit.Formats(res))
							return
						}
					}
				}
				m.Echo(ice.OK)
			}},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(web.SPIDE_HEADER, web.ContentType, web.ApplicationJSON, web.UserAgent, "Mozilla/5.0")
				m.Cmdy(web.SPIDE, m.Option(ice.DEV), web.SPIDE_RAW, m.Option(ice.CMD), m.Option(cli.API), web.SPIDE_DATA, m.Option(ice.ARG)).ProcessInner()
				m.StatusTime(nfs.SCRIPT, `curl "`+kit.MergeURL2(m.Cmd(web.SPIDE, m.Option(ice.DEV)).Append(web.CLIENT_ORIGIN), m.Option(cli.API))+`" -H "Content-Type: application/json"`+` -d '`+m.Option(ice.ARG)+`'`)
			}},
		}, mdb.ZoneAction(mdb.FIELDS, "time,id,name,cmd,api,arg,res")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(web.SPIDE).RenameAppend(web.CLIENT_NAME, ice.DEV, web.CLIENT_URL, "address")
			} else if mdb.ZoneSelect(m, arg[1:]...); len(arg) > 1 {
				m.PushAction(ctx.RUN, cli.CHECK).Action(cli.CHECK)
			}
		}},
	})
}
