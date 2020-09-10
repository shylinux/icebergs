package tmux

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
	"time"
)

const (
	LOCAL = "local"
	RELAY = "relay"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			LOCAL: {Name: "local", Help: "虚拟机", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
			RELAY: {Name: "relay", Help: "跳板机", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
				"count", 100, "sleep", "100ms", "tail", kit.Dict(
					"verify", "Verification code:",
					"password", "Password:",
					"login", "[relay ~]$",
				),
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Cmdy(cli.SYSTEM, "tmux", "ls"); m.Append("code") != "0" {
					return
				}

				m.Cmd(web.PROXY, "add", "tmux", m.AddCmd(&ice.Command{Name: "proxy", Help: "代理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
					m.Cmd("session").Table(func(index int, value map[string]string, head []string) {
						if value["tag"] == "1" {
							m.Echo(value["session"])
						}
					})
				}}))
			}},
			"_install": {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option("cmd_dir", m.Conf("install", "meta.path"))
				m.Cmd(cli.SYSTEM, "git", "clone", "https://github.com/tmux/tmux")
			}},
			code.PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd("nfs.link", path.Join(os.Getenv("HOME"), ".tmux.conf"), "etc/conf/tmux.conf")
			}},
			code.PROJECT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			}},

			gdb.INIT: {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Watch(gdb.DREAM_START)
				return

				if m.Richs(web.FAVOR, nil, "tmux.auto", nil) == nil {
					m.Cmd(web.FAVOR, "tmux.auto", web.TYPE_SHELL, "脚本", `curl $ctx_dev/publish/auto.sh > auto.sh`)
					m.Cmd(web.FAVOR, "tmux.auto", web.TYPE_SHELL, "脚本", `source auto.sh`)
					m.Cmd(web.FAVOR, "tmux.auto", web.TYPE_SHELL, "脚本", `ShyInit && ShyLogin && trap ShyLogout EXIT`)
				}

				for _, v := range []string{"auto.sh", "auto.vim", "auto.tmux"} {
					p := path.Join(m.Conf("web.code.publish", "meta.path"), v)
					if _, e := os.Stat(p); e != nil && os.IsNotExist(e) {
						// 下载脚本
						if h := m.Cmdx(web.SPIDE, "shy", "cache", "GET", "/publish/"+v); h != "" {
							m.Cmd(web.STORY, web.WATCH, h, p)
						}
					}
				}
			}},
			gdb.AUTO: {Name: "auto", Help: "自动化", Action: map[string]*ice.Action{
				web.DREAM_START: {Name: "dream.start", Hand: func(m *ice.Message, arg ...string) {
					if m.Cmd(cli.SYSTEM, TMUX, "has-session", "-t", arg[0]).Append(cli.CMD_CODE) == "0" {
						return
					}
					// 创建会话
					m.Option(cli.CMD_ENV, "TMUX", "", "ctx_pod", arg[0], "ctx_dev", m.Conf(cli.RUNTIME, "conf.ctx_dev"))
					m.Option(cli.CMD_DIR, path.Join(m.Conf(web.DREAM, "meta.path"), arg[0]))
					m.Cmd(cli.SYSTEM, TMUX, "new-session", "-ds", arg[0])
					return

					// 共享空间
					share, dev := "", kit.Select(m.Conf(cli.RUNTIME, "conf.ctx_dev"), m.Conf(cli.RUNTIME, "host.ctx_self"))
					m.Richs(web.SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
						share = kit.Format(value["share"])
					})

					// 环境变量
					m.Option("cmd_env", "TMUX", "", "ctx_dev", dev, "ctx_share", share)
					m.Option("cmd_dir", path.Join(m.Conf(web.DREAM, "meta.path"), arg[0]))

					if arg[0] != "" && m.Cmd(cli.SYSTEM, TMUX, "has-session", "-t", arg[0]).Append("code") != "0" {
						// 创建会话
						m.Cmd(cli.SYSTEM, TMUX, "new-session", "-ds", arg[0])
					}

					if m.Option("local") != "" {
						// 创建容器
						m.Cmd("local", m.Option("local"), arg[0])
					}
					if m.Option("relay") != "" {
						// 远程登录
						m.Cmd("relay", m.Option("relay"), arg[0])
					}

					for _, v := range kit.Simple(m.Optionv("before")) {
						// 前置命令
						m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[0], v, "Enter")
					}

					// 连接参数
					m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[0], "export ctx_dev=", dev, "Enter")
					m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[0], "export ctx_share=", share, "Enter")

					m.Cmd(web.FAVOR, kit.Select("tmux.auto", arg, 1)).Table(func(index int, value map[string]string, head []string) {
						switch value["type"] {
						case web.TYPE_SHELL:
							// 发送命令
							m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[0], value["text"], "Enter")
							time.Sleep(10 * time.Millisecond)
						}
					})

					for _, v := range kit.Simple(m.Optionv("after")) {
						// 后置命令
						m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[0], v, "Enter")
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

			"make": {Name: "make name cmd...", Help: "个性化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				session := m.Conf(cli.RUNTIME, "node.name")
				if arg[1] == "session" {
					session, arg[2], arg = arg[2], arg[0], arg[2:]
				}

				if m.Warn(m.Cmd(cli.SYSTEM, TMUX, "has-session", "-t", session).Append("code") != "0", "session miss") {
					// 会话不存在
					return
				}

				if m.Cmdx("session", session, "has", arg[0]) != "" {
					// 窗口已存在
					return
				}

				switch arg[1] {
				case "init":
					m.Cmdx(cli.SYSTEM, TMUX, "rename-window", "-t", session, arg[0])
					arg[1], arg = arg[0], arg[1:]
				case "link":
					m.Cmdx(cli.SYSTEM, TMUX, "link-window", "-dt", session, "-s", arg[2])
					return
				default:
					m.Cmd(cli.SYSTEM, TMUX, "new-window", "-dt", session, "-n", arg[0])
				}

				for _, v := range arg[1:] {
					switch ls := kit.Split(v); ls[1] {
					case "v":
						m.Cmd(cli.SYSTEM, TMUX, "split-window", "-h", "-dt", session+":"+arg[0]+"."+ls[0], ls[2:])
					case "u", "split-window":
						m.Cmd(cli.SYSTEM, TMUX, "split-window", "-dt", session+":"+arg[0]+"."+ls[0], ls[2:])
					case "k":
						m.Cmd(cli.SYSTEM, TMUX, "send-key", "-t", session+":"+arg[0]+"."+ls[0], ls[2:])
					default:
						m.Cmd(cli.SYSTEM, TMUX, ls[1], "-t", session+":"+arg[0]+"."+ls[0], ls[2:])
					}
				}
			}},

			"relay": {Name: "relay [name [favor]]", Help: "跳板机", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 认证列表
					m.Richs(cmd, nil, "*", func(key string, value map[string]interface{}) {
						m.Push(key, value, []string{"name", "username", "website"})
					})
					return
				}
				if len(arg) == 1 {
					// 认证详情
					m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
						m.Push("detail", value)
					})
					return
				}

				if len(arg) > 4 {
					// 添加认证
					m.Rich(cmd, nil, kit.Dict(kit.MDB_NAME, arg[0], kit.MDB_TEXT, arg[1],
						"username", arg[2], "website", arg[3], "password", arg[4],
					))
					return
				}

				m.Richs(cmd, nil, arg[0], func(key string, value map[string]interface{}) {
					// 登录命令
					m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[1], kit.Format("ssh %s@%s", value["username"], value["website"]), "Enter")

					sleep := kit.Duration(m.Conf(cmd, "meta.sleep"))
					for i := 0; i < kit.Int(m.Conf(cmd, "meta.count")); i++ {
						time.Sleep(sleep)
						list := strings.Split(strings.TrimSpace(m.Cmdx(cli.SYSTEM, TMUX, "capture-pane", "-p")), "\n")

						if tail := list[len(list)-1]; tail == m.Conf(cmd, "meta.tail.login") {
							// 登录成功
							break
						} else if tail == m.Conf(cmd, "meta.tail.password") {
							// 输入密码
							m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[1], value["password"], "Enter")
						} else if tail == m.Conf(cmd, "meta.tail.verify") {
							// 输入密码
							m.Cmdy(cli.SYSTEM, TMUX, "send-keys", "-t", arg[1], m.Cmdx("aaa.totp.get", value["text"]), "Enter")
						}
					}
				})
			}},

			"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				current := ""
				m.Cmd("session").Table(func(index int, value map[string]string, head []string) {
					if value["tag"] == "1" {
						current = value["session"]
					}
				})

				m.Option(ice.MSG_OUTPUT, ice.RENDER_RESULT)
				switch arg = kit.Split(kit.Select("tmux.auto", arg, 0)); arg[0] {
				case "ice":
					if m.Cmdy(arg[1:]); len(m.Resultv()) == 0 {
						m.Table()
					}
				default:
					m.Cmd("auto", current, arg)
				}
			}},
		},
	}, nil)
}
