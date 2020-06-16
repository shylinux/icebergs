package cli

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
)

var RUNTIME = ice.Name("runtime", nil)

var Index = &ice.Context{Name: "cli", Help: "命令模块",
	Configs: map[string]*ice.Config{
		RUNTIME: {Name: "runtime", Help: "运行环境", Value: kit.Dict()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			// 启动配置
			m.Conf(RUNTIME, "conf.ctx_self", os.Getenv("ctx_self"))
			m.Conf(RUNTIME, "conf.ctx_dev", os.Getenv("ctx_dev"))
			m.Conf(RUNTIME, "conf.ctx_shy", os.Getenv("ctx_shy"))
			m.Conf(RUNTIME, "conf.ctx_pid", os.Getenv("ctx_pid"))

			// 主机信息
			m.Conf(RUNTIME, "host.GOARCH", runtime.GOARCH)
			m.Conf(RUNTIME, "host.GOOS", runtime.GOOS)
			m.Conf(RUNTIME, "host.pid", os.Getpid())

			n := kit.Int(kit.Select("20", m.Conf(RUNTIME, "host.GOMAXPROCS")))
			m.Logs("host", "gomaxprocs", n)
			runtime.GOMAXPROCS(n)

			// 启动信息
			if name, e := os.Hostname(); e == nil {
				m.Conf(RUNTIME, "boot.hostname", kit.Select(name, os.Getenv("HOSTNAME")))
			}
			if user, e := user.Current(); e == nil {
				m.Conf(RUNTIME, "boot.username", path.Base(kit.Select(user.Name, os.Getenv("USER"))))
				m.Cmd(ice.AAA_ROLE, "root", m.Conf(RUNTIME, "boot.username"))
			}
			if name, e := os.Getwd(); e == nil {
				name = path.Base(kit.Select(name, os.Getenv("PWD")))
				ls := strings.Split(name, "/")
				name = ls[len(ls)-1]
				ls = strings.Split(name, "\\")
				name = ls[len(ls)-1]
				m.Conf(RUNTIME, "boot.pathname", name)
			}

			// 启动记录
			count := m.Confi(RUNTIME, "boot.count") + 1
			m.Conf(RUNTIME, "boot.count", count)

			// 节点信息
			m.Conf(RUNTIME, "node.time", m.Time())
			m.Conf(RUNTIME, "node.type", ice.WEB_WORKER)
			m.Conf(RUNTIME, "node.name", m.Conf(RUNTIME, "boot.pathname"))
			m.Log("info", "runtime %v", kit.Formats(m.Confv(RUNTIME)))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(RUNTIME, SYSTEM)
		}},

		RUNTIME: {Name: "runtime", Help: "运行环境", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(ice.CTX_CONFIG, RUNTIME, arg)
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
