package code

import (
	"net/http"
	_ "net/http/pprof"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	BINNARY = "binnary"
	SERVICE = "service"
	SECONDS = "seconds"
)

const PPROF = "pprof"

func init() {
	Index.MergeCommands(ice.Commands{
		PPROF: {Name: "pprof zone id auto", Help: "性能分析", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				web.AddRewrite(func(w http.ResponseWriter, r *http.Request) bool {
					if p := r.URL.Path; strings.HasPrefix(p, "/debug/") {
						r.URL.Path = strings.Replace(r.URL.Path, "/debug/", "/code/", -1)
						m.Debug("rewrite %v -> %v", p, r.URL.Path)
					}
					return false
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case BINNARY:
					m.Cmdy(nfs.DIR, ice.BIN, nfs.DIR_CLI_FIELDS).RenameAppend(nfs.PATH, BINNARY)
				case SERVICE:
					m.Cmd(web.SPIDE, func(value ice.Maps) {
						m.Push(SERVICE, kit.MergeURL2(value[web.CLIENT_URL], "/debug/pprof/profile"))
					})
				}
			}},
			mdb.CREATE: {Name: "create zone=some binnary=bin/ice.bin service='http://localhost:9020/debug/pprof/profile' seconds=3", Help: "创建"},

			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				msg := m.Cmd(web.SPIDE, ice.DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option(SERVICE), SECONDS, m.Option(SECONDS))
				cmd := kit.Simple(m.Configv(PPROF), "-text", m.Option(BINNARY), msg.Append(nfs.FILE))
				m.Option(mdb.TEXT, strings.Join(kit.Slice(strings.Split(m.Cmdx(cli.SYSTEM, cmd), ice.NL), 0, 20), ice.NL))
				m.Option(nfs.FILE, msg.Append(nfs.FILE))
				mdb.ZoneInsert(m, m.OptionSimple("zone,text,file"))
				m.Echo(m.Option(mdb.TEXT)).ProcessInner()
			}},
			web.SERVE: {Name: "serve", Help: "展示", Hand: func(m *ice.Message, arg ...string) {
				u := web.OptionUserWeb(m)
				p := u.Hostname() + ice.DF + m.Cmdx(tcp.PORT, aaa.RIGHT)
				m.Cmd(cli.DAEMON, m.Configv(PPROF), "-http="+p, m.Option(BINNARY), m.Option(nfs.FILE))
				m.Echo("http://%s/ui/top", p).ProcessInner()
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,text,file", PPROF, kit.List(GO, "tool", PPROF))), Hand: func(m *ice.Message, arg ...string) {
			m.Fields(len(arg), "time,zone,count,binnary,service,seconds", m.Config(mdb.FIELD))
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.EchoAnchor(web.MergeLink(m, "/code/pprof/"))
				m.PushAction(ice.RUN, mdb.REMOVE)
				m.Action(mdb.CREATE)
			} else {
				m.Tables(func(value ice.Maps) {
					m.PushDownload(mdb.LINK, "pprof.pd.gz", value[nfs.FILE])
					m.PushButton(web.SERVE)
				})
			}
		}},
		web.PP(PPROF): {Name: "/pprof/", Help: "性能分析", Hand: func(m *ice.Message, arg ...string) {
			defer m.Render(ice.RENDER_VOID)
			m.R.URL.Path = "/debug" + m.R.URL.Path
			http.DefaultServeMux.ServeHTTP(m.W, m.R)
		}},
	})
}
