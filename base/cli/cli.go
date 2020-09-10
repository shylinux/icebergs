package cli

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
)

var UserName = ""
var PassWord = ""
var HostName = ""
var PathName = ""
var NodeName = ""
var NodeType = ""

func NodeInfo(m *ice.Message, kind, name string) {
	m.Conf(RUNTIME, "node.type", kind)
	m.Conf(RUNTIME, "node.name", name)
	NodeName = name
	NodeType = kind
}

const RUNTIME = "runtime"

var Index = &ice.Context{Name: "cli", Help: "命令模块",
	Configs: map[string]*ice.Config{
		RUNTIME: {Name: "runtime", Help: "运行环境", Value: kit.Dict()},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			// 启动配置
			m.Conf(RUNTIME, "conf.ctx_self", os.Getenv("ctx_self"))
			m.Conf(RUNTIME, "conf.ctx_dev", os.Getenv("ctx_dev"))
			m.Conf(RUNTIME, "conf.ctx_shy", os.Getenv("ctx_shy"))
			m.Conf(RUNTIME, "conf.ctx_pid", os.Getenv("ctx_pid"))
			m.Conf(RUNTIME, "conf.ctx_user", os.Getenv("ctx_user"))
			m.Conf(RUNTIME, "conf.ctx_pod", os.Getenv("ctx_pod"))

			// 主机信息
			m.Conf(RUNTIME, "host.GOARCH", runtime.GOARCH)
			m.Conf(RUNTIME, "host.GOOS", runtime.GOOS)
			m.Conf(RUNTIME, "host.pid", os.Getpid())

			// 启动信息
			if m.Conf(RUNTIME, "boot.username", kit.Select(os.Getenv("USER"), os.Getenv("ctx_user"))) == "" {
				if user, e := user.Current(); e == nil && user.Name != "" {
					m.Conf(RUNTIME, "boot.username", kit.Select(user.Name, os.Getenv("ctx_user")))
				}
			}
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
			UserName = m.Conf(RUNTIME, "boot.username")
			HostName = m.Conf(RUNTIME, "boot.hostname")
			PathName = m.Conf(RUNTIME, "boot.pathname")

			// 启动记录
			count := m.Confi(RUNTIME, "boot.count") + 1
			m.Conf(RUNTIME, "boot.count", count)

			// 节点信息
			m.Conf(RUNTIME, "node.time", m.Time())
			NodeInfo(m, "worker", m.Conf(RUNTIME, "boot.pathname"))
			m.Info("runtime %v", kit.Formats(m.Confv(RUNTIME)))

			n := kit.Int(kit.Select("20", m.Conf(RUNTIME, "host.GOMAXPROCS")))
			m.Logs("host", "gomaxprocs", n)
			runtime.GOMAXPROCS(n)

			m.Cmdy(mdb.ENGINE, mdb.CREATE, "shell", m.AddCmd(&ice.Command{Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(SYSTEM, arg[2])
			}}))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(RUNTIME, SYSTEM, DAEMON)
		}},

		"proc": {Name: "proc name=auto PID=auto auto", Help: "进程管理", Action: map[string]*ice.Action{
			"kill": {Name: "kill", Help: "结束", Hand: func(m *ice.Message, arg ...string) {
				if p, e := os.FindProcess(kit.Int(m.Option("PID"))); m.Assert(e) {
					m.Assert(p.Kill())
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			msg := m.Spawn()
			msg.Split(m.Cmdx(SYSTEM, "ps", "ux"), "", " ", "\n")
			msg.Table(func(index int, value map[string]string, head []string) {
				if m.Appendv(ice.MSG_APPEND, "action", head); len(arg) == 2 && value["PID"] == arg[1] {
					m.PushRender("action", "button", "结束")
					m.Push("", value)
					return
				}
				if len(arg) == 0 || len(arg) == 1 && strings.Contains(value["COMMAND"], arg[0]) {
					m.PushRender("action", "button", "结束")
					m.Push("", value)
				}
			})
		}},
		RUNTIME: {Name: "runtime name auto", Help: "运行环境", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch kit.Select("", arg, 0) {
			case "procinfo":
				m.Split(m.Cmdx(SYSTEM, "ps", "u"), "", " ", "\n")

			case "hostinfo":
				if f, e := os.Open("/proc/cpuinfo"); e == nil {
					defer f.Close()
					if b, e := ioutil.ReadAll(f); e == nil {
						m.Push("nCPU", bytes.Count(b, []byte("processor")))
					}
				}
				if f, e := os.Open("/proc/meminfo"); e == nil {
					defer f.Close()
					if b, e := ioutil.ReadAll(f); e == nil {
						for i, ls := range strings.Split(string(b), "\n") {
							vs := kit.Split(ls, ": ")
							m.Push(strings.TrimSpace(vs[0]), kit.FmtSize(kit.Int64(strings.TrimSpace(vs[1]))*1024))
							if i > 1 {
								break
							}
						}
					}
				}
				m.Push("uptime", kit.Split(m.Cmdx(SYSTEM, "uptime"), ",")[0])
			case "diskinfo":
				m.Spawn().Split(m.Cmdx(SYSTEM, "df", "-h"), "", " ", "\n").Table(func(index int, value map[string]string, head []string) {
					if strings.HasPrefix(value["Filesystem"], "/dev") {
						m.Push("", value, head)
					}
				})
			case "ifconfig":
				m.Cmdy("tcp.ip")
			case "userinfo":
				m.Split(m.Cmdx(SYSTEM, "who"), "user term time", " ", "\n")

			case "hostname":
				m.Conf(RUNTIME, "boot.hostname", arg[1])
				HostName = arg[1]
				m.Echo(HostName)
			default:
				m.Cmdy(ctx.CONFIG, RUNTIME, arg)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil, RUNTIME, SYSTEM, DAEMON, PYTHON) }
