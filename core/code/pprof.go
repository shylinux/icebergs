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
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PPROF: {Name: PPROF, Help: "性能分析", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE,
				PPROF, []string{GO, "tool", PPROF},
			)},
		},
		Commands: map[string]*ice.Command{
			"/pprof/": {Name: "/pprof/", Help: "性能分析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.R.URL.Path = strings.Replace("/code"+m.R.URL.Path, "code", "debug", 1)
				http.DefaultServeMux.ServeHTTP(m.W, m.R)
				m.Render(ice.RENDER_VOID)
			}},
			PPROF: {Name: "pprof zone id auto create", Help: "性能分析", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone=some binnary service seconds=3", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, PPROF, "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert zone type name text", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, PPROF, "", mdb.HASH, kit.MDB_ZONE, arg[1])
					m.Cmdy(mdb.INSERT, PPROF, "", mdb.ZONE, m.Option(kit.MDB_ZONE), arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, PPROF, "", mdb.ZONE, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, PPROF, "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case BINNARY:
						m.Cmd(nfs.DIR, "bin/", "path,size,time").Table(func(index int, value map[string]string, head []string) {
							m.Push(BINNARY, value["path"])
							m.Push("", value, []string{"size,time"})
						})
					case SERVICE:
						m.Cmd(web.SPIDE).Table(func(index int, value map[string]string, head []string) {
							m.Push(SERVICE, kit.MergeURL2(value["client.url"], "/debug/pprof/profile"))
						})
					}
				}},

				cli.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option(SERVICE), SECONDS, m.Option(SECONDS))

					cmd := kit.Simple(m.Confv(PPROF, kit.Keym(PPROF)), "-text", m.Option(BINNARY), msg.Append(kit.MDB_FILE))
					res := strings.Split(m.Cmdx(cli.SYSTEM, cmd), ice.NL)
					if len(res) > 20 {
						res = res[:20]
					}

					m.Cmd(mdb.INSERT, PPROF, "", mdb.ZONE, m.Option(kit.MDB_ZONE),
						kit.MDB_TEXT, strings.Join(res, ice.NL), kit.MDB_FILE, msg.Append(kit.MDB_FILE))
					m.Echo(strings.Join(res, ice.NL))
					m.ProcessInner()
				}},
				web.SERVE: {Name: "serve", Help: "展示", Hand: func(m *ice.Message, arg ...string) {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					p := u.Hostname() + ":" + m.Cmdx(tcp.PORT, aaa.RIGHT)

					m.Cmd(cli.DAEMON, m.Confv(PPROF, kit.Keym(PPROF)), "-http="+p, m.Option(BINNARY), m.Option(kit.MDB_FILE))
					m.Echo("http://%s/ui/top", p)
					m.ProcessInner()
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,zone,count,binnary,service,seconds", "time,id,text,file")
				if m.Cmdy(mdb.SELECT, PPROF, "", mdb.ZONE, arg); len(arg) == 0 {
					m.PushAction(cli.RUN, mdb.REMOVE)
					return
				}

				m.Table(func(index int, value map[string]string, head []string) {
					m.PushDownload(kit.MDB_LINK, "pprof.pd.gz", value[kit.MDB_FILE])
					m.PushButton(web.SERVE)
				})
			}},
		},
	})
}
