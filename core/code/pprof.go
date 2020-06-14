package code

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"net/http"
	_ "net/http/pprof"
	"strings"
)

const (
	PPROF = "pprof"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			"pprof": {Name: "pprof", Help: "性能分析", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
				"stop", "ps aux|grep pprof|grep -v grep|cut -d' ' -f2|xargs -n1 kill",
			)},
		},
		Commands: map[string]*ice.Command{
			"pprof": {Name: "pprof run name time", Help: "性能分析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Show(cmd, arg...) {
					return
				}

				switch arg[0] {
				case "run":
					m.Richs(cmd, nil, arg[1], func(key string, value map[string]interface{}) {
						m.Gos(m.Spawn(), func(msg *ice.Message) {
							m.Sleep("1s").Grows(cmd, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
								// 压测命令
								m.Cmd(ice.WEB_FAVOR, "pprof", "shell", value[kit.MDB_TEXT], m.Cmdx(kit.Split(kit.Format(value[kit.MDB_TEXT]))))
							})
						})

						// 启动监控
						name := arg[1] + ".pd.gz"
						value = value["meta"].(map[string]interface{})
						msg := m.Cmd(ice.WEB_SPIDE, "self", "cache", "GET", kit.Select("/code/pprof/profile", value["remote"]), "seconds", kit.Select("5", arg, 2))
						m.Cmd(ice.WEB_FAVOR, "pprof", "shell", "text", m.Cmdx(ice.CLI_SYSTEM, "go", "tool", "pprof", "-text", msg.Append("text")))
						m.Cmd(ice.WEB_FAVOR, "pprof", "pprof", name, msg.Append("data"))

						arg = kit.Simple("web", value[kit.MDB_TEXT], msg.Append("text"))
					})

					fallthrough
				case "web":
					// 展示结果
					p := kit.Format("%s:%s", m.Conf(ice.WEB_SHARE, "meta.host"), m.Cmdx("tcp.getport"))
					m.Cmd(ice.CLI_DAEMON, "go", "tool", "pprof", "-http="+p, arg[1:])
					m.Cmd(ice.WEB_FAVOR, "pprof", "bin", arg[1], m.Cmd(ice.WEB_CACHE, "catch", "bin", arg[1]).Append("data"))
					m.Cmd(ice.WEB_FAVOR, "pprof", "spide", arg[2], "http://"+p)
					m.Echo(p)

				case "stop":
					m.Cmd(ice.CLI_SYSTEM, "sh", "-c", m.Conf(cmd, "meta.stop"))

				case "add":
					key := m.Rich(cmd, nil, kit.Data(
						kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2], "remote", arg[3],
					))

					for i := 4; i < len(arg)-1; i += 2 {
						m.Grow(cmd, kit.Keys(kit.MDB_HASH, key), kit.Dict(
							kit.MDB_NAME, arg[i], kit.MDB_TEXT, arg[i+1],
						))
					}
				}
			}},
			"/pprof/": {Name: "/pprof/", Help: "性能分析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.R.URL.Path = strings.Replace("/code"+m.R.URL.Path, "code", "debug", 1)
				http.DefaultServeMux.ServeHTTP(m.W, m.R)
				m.Render(ice.RENDER_VOID)
			}},
		},
	}, nil)
}
