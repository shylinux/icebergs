package code

import (
	"net/http"
	_ "net/http/pprof"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const PPROF = "pprof"

func init() {
	const (
		BINNARY = "binnary"
		SERVICE = "service"
		SECONDS = "seconds"
	)
	web.Index.MergeCommands(ice.Commands{"/debug/": {Actions: aaa.WhiteAction(), Hand: func(m *ice.Message, arg ...string) {
		defer m.Render(ice.RENDER_VOID)
		http.DefaultServeMux.ServeHTTP(m.W, m.R)
	}}})
	Index.MergeCommands(ice.Commands{
		PPROF: {Name: "pprof zone id auto", Help: "性能分析", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case BINNARY:
					m.Cmdy(nfs.DIR, ice.BIN, nfs.DIR_CLI_FIELDS).RenameAppend(nfs.PATH, BINNARY)
				case SERVICE:
					m.Cmd(web.SPIDE, func(value ice.Maps) { m.Push(SERVICE, kit.MergeURL2(value[web.CLIENT_URL], "/debug/pprof/profile")) })
				}
			}},
			mdb.CREATE: {Name: "create zone*=some binnary*=bin/ice.bin service*='http://localhost:9020/debug/pprof/profile' seconds*=30"},
			cli.START: {Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(BINNARY) == "" {
					return
				}
				msg := m.Cmd(web.SPIDE, ice.DEV, web.SPIDE_CACHE, http.MethodGet, m.Option(SERVICE), m.OptionSimple(SECONDS))
				cmd := kit.Simple(mdb.Configv(m, PPROF), "-text", m.Option(BINNARY), msg.Append(nfs.FILE))
				m.Option(mdb.TEXT, strings.Join(kit.Slice(strings.Split(m.Cmdx(cli.SYSTEM, cmd), lex.NL), 0, 20), lex.NL))
				mdb.ZoneInsert(m, m.OptionSimple("zone,text"), msg.AppendSimple(nfs.FILE))
				m.Echo(m.Option(mdb.TEXT)).ProcessInner()
			}},
			web.SERVE: {Help: "展示", Hand: func(m *ice.Message, arg ...string) {
				u := web.UserWeb(m)
				p := u.Hostname() + nfs.DF + m.Cmdx(tcp.PORT, aaa.RIGHT)
				m.Cmd(cli.DAEMON, mdb.Configv(m, PPROF), "-http="+p, m.Option(BINNARY), m.Option(nfs.FILE))
				m.Sleep3s().ProcessOpen(kit.Format("http://%s/ui/top", p))
			}},
		}, mdb.ZoneAction(mdb.FIELDS, "time,zone,count,binnary,service,seconds", mdb.FIELD, "time,id,text,file", PPROF, kit.List(GO, "tool", PPROF))), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.PushAction(cli.START, mdb.REMOVE).Action(mdb.CREATE)
				m.EchoAnchor(web.MergeLink(m, "/debug/pprof/"))
			} else {
				m.Table(func(value ice.Maps) { m.PushDownload(mdb.LINK, "pprof.pd.gz", value[nfs.FILE]).PushButton(web.SERVE) })
			}
		}},
	})
}
