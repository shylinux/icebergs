package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"net/http"
	_ "net/http/pprof"
	"strings"
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
				PPROF, []string{"go", "tool", "pprof"},
			)},
		},
		Commands: map[string]*ice.Command{
			PPROF: {Name: "pprof zone id auto create", Help: "性能分析", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone binnary service seconds", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, PPROF, "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert zone type name text", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, PPROF, _sub_key(m, m.Option(kit.MDB_ZONE)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_ID) != "" {
						m.Cmdy(mdb.MODIFY, PPROF, _sub_key(m, m.Option(kit.MDB_ZONE)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID), arg)
					} else {
						m.Cmdy(mdb.MODIFY, PPROF, "", mdb.HASH, kit.MDB_ZONE, m.Option(kit.MDB_ZONE), arg)
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, PPROF, "", mdb.HASH, kit.MDB_ZONE, m.Option(kit.MDB_ZONE))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case "binnary":
						m.Cmd(nfs.DIR, "bin/", "path,size,time").Table(func(index int, value map[string]string, head []string) {
							m.Push("binnary", value["path"])
							m.Push("", value, []string{"size,time"})
						})
					case "service":
						m.Cmd(web.SPIDE).Table(func(index int, value map[string]string, head []string) {
							m.Push("service", kit.MergeURL2(value["client.url"], "/debug/pprof/profile"))
						})
					}
				}},

				cli.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, m.Option(SERVICE), SECONDS, m.Option(SECONDS))

					cmd := kit.Simple(m.Confv(PPROF, "meta.pprof"), "-text", m.Option(BINNARY), msg.Append(kit.MDB_FILE))
					res := strings.Split(m.Cmdx(cli.SYSTEM, cmd), "\n")
					if len(res) > 20 {
						res = res[:20]
					}

					m.Cmd(mdb.INSERT, PPROF, _sub_key(m, m.Option(kit.MDB_ZONE)), mdb.LIST, kit.MDB_TEXT, strings.Join(res, "\n"), kit.MDB_FILE, msg.Append(kit.MDB_FILE))
					m.Echo(strings.Join(res, "\n"))
					m.ProcessInner()
				}},
				web.SERVE: {Name: "serve", Help: "展示", Hand: func(m *ice.Message, arg ...string) {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					p := u.Hostname() + ":" + m.Cmdx(tcp.PORT, aaa.RIGHT)

					m.Cmd(cli.DAEMON, m.Confv(PPROF, "meta.pprof"), "-http="+p, m.Option(BINNARY), m.Option(kit.MDB_FILE))
					m.Echo("http://%s/ui/top", p)
					m.ProcessInner()
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,count,zone,binnary,service,seconds", kit.Select("time,id,text,binnary,file", mdb.DETAIL, len(arg) > 1), len(arg) > 0))
				m.Cmdy(mdb.SELECT, PPROF, "", mdb.ZONE, arg)
				if len(arg) == 0 {
					m.PushAction(cli.RUN, mdb.REMOVE)

				} else {
					m.Table(func(index int, value map[string]string, head []string) {
						m.PushDownload(kit.MDB_LINK, "pprof.pd.gz", value[kit.MDB_FILE])
						m.PushButton(web.SERVE)
					})
				}
			}},

			"/pprof/": {Name: "/pprof/", Help: "性能分析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.R.URL.Path = strings.Replace("/code"+m.R.URL.Path, "code", "debug", 1)
				http.DefaultServeMux.ServeHTTP(m.W, m.R)
				m.Render(ice.RENDER_VOID)
			}},
		},
	})
}
