package cli

import (
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func NodeInfo(m *ice.Message, kind, name string) {
	name = strings.ReplaceAll(name, ".", "_")
	m.Conf(RUNTIME, "node.type", kind)
	m.Conf(RUNTIME, "node.name", name)
	ice.Info.NodeName = name
	ice.Info.NodeType = kind
}

var Index = &ice.Context{Name: "cli", Help: "命令模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			// 启动配置
			for _, k := range []string{"ctx_self", "ctx_dev", "ctx_shy", "ctx_pid", "ctx_user", "ctx_share", "ctx_river"} {
				m.Conf(RUNTIME, kit.Keys("conf", k), os.Getenv(k))
			}

			// 主机信息
			m.Conf(RUNTIME, "host.GOARCH", runtime.GOARCH)
			m.Conf(RUNTIME, "host.GOOS", runtime.GOOS)
			m.Conf(RUNTIME, "host.pid", os.Getpid())

			// 启动信息
			if name, e := os.Hostname(); e == nil {
				m.Conf(RUNTIME, "boot.hostname", kit.Select(name, os.Getenv("HOSTNAME")))
			}
			if name, e := os.Getwd(); e == nil {
				name = path.Base(kit.Select(name, os.Getenv("PWD")))
				ls := strings.Split(name, "/")
				name = ls[len(ls)-1]
				ls = strings.Split(name, "\\")
				name = ls[len(ls)-1]
				m.Conf(RUNTIME, "boot.pathname", name)
			}
			if m.Conf(RUNTIME, "boot.username", kit.Select(os.Getenv("USER"), os.Getenv("ctx_user"))) == "" {
				if user, e := user.Current(); e == nil && user.Name != "" {
					m.Conf(RUNTIME, "boot.username", kit.Select(user.Name, os.Getenv("ctx_user")))
				}
			}
			ice.Info.HostName = m.Conf(RUNTIME, "boot.hostname")
			ice.Info.PathName = m.Conf(RUNTIME, "boot.pathname")
			ice.Info.UserName = m.Conf(RUNTIME, "boot.username")

			ice.Info.CtxShare = m.Conf(RUNTIME, "conf.ctx_share")
			ice.Info.CtxRiver = m.Conf(RUNTIME, "conf.ctx_river")

			// 启动次数
			count := kit.Int(m.Conf(RUNTIME, "boot.count")) + 1
			m.Conf(RUNTIME, "boot.count", count)

			// 节点信息
			m.Conf(RUNTIME, "node.time", m.Time())
			NodeInfo(m, "worker", m.Conf(RUNTIME, "boot.pathname"))
			m.Info("runtime %v", kit.Formats(m.Confv(RUNTIME)))

			n := kit.Int(kit.Select("1", m.Conf(RUNTIME, "host.GOMAXPROCS")))
			m.Logs("host", "gomaxprocs", n)
			runtime.GOMAXPROCS(n)

			// 版本信息
			kit.Fetch(kit.UnMarshal(kit.Format(ice.Info.Build)), func(key string, value interface{}) {
				m.Conf(RUNTIME, kit.Keys("make", strings.ToLower(key)), value)
			})
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
	},
}

func init() { ice.Index.Register(Index, nil, RUNTIME, SYSTEM, DAEMON, QRCODE) }
