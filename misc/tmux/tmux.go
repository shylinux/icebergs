package tmux

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
	"time"
)

const TMUX = "tmux"
const (
	TEXT    = "text"
	BUFFER  = "buffer"
	SESSION = "session"
	WINDOW  = "window"
	PANE    = "pane"
	VIEW    = "view"

	LOCAL = "local"
	RELAY = "relay"
)

var _tmux = []string{cli.SYSTEM, TMUX}

var Index = &ice.Context{Name: TMUX, Help: "工作台",
	Configs: map[string]*ice.Config{
		SESSION: {Name: "session", Help: "会话", Value: kit.Data(
			"format", "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}",
			"fields", "id,tag,session,windows,height,width",
		)},
		WINDOW: {Name: "windows", Help: "窗口", Value: kit.Data(
			"format", "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
			"fields", "id,tag,window,panes,height,width",
		)},
		PANE: {Name: "panes", Help: "终端", Value: kit.Data(
			"format", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
			"fields", "id,tag,pane,tty,height,width",
		)},

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
		TEXT: {Name: "text 保存:button 清空:button text:textarea", Help: "文本", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] != "" {
				m.Cmd(_tmux, "set-buffer", arg[0])
			}

			text := m.Cmdx(_tmux, "show-buffer")
			m.Cmdy("web.wiki.image", "qrcode", text)

			m.Echo("\n")
			m.Echo(text)
			m.Echo("\n")
			m.Render("")
		}},
		BUFFER: {Name: "buffer [buffer=auto [value]] auto", Help: "缓存", Action: map[string]*ice.Action{
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case "text":
					m.Cmd(_tmux, "set-buffer", "-b", m.Option("buffer"), arg[1])
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 {
				// 设置缓存
				m.Cmd(_tmux, "set-buffer", "-b", arg[0], arg[1])
			}
			if len(arg) > 0 {
				// 查看缓存
				m.Echo(m.Cmdx(_tmux, "show-buffer", "-b", arg[0]))
				return
			}

			// 缓存列表
			for i, v := range kit.Split(m.Cmdx(_tmux, "list-buffers"), "\n", "\n", "\n") {
				ls := strings.SplitN(v, ": ", 3)
				m.Push("buffer", ls[0])
				m.Push("size", ls[1])
				if i < 3 {
					m.Push("text", m.Cmdx(_tmux, "show-buffer", "-b", ls[0]))
				} else {
					m.Push("text", ls[2][1:])
				}
			}
		}},
		SESSION: {Name: "session [session=auto [window=auto [pane=auto [cmd]]]] auto", Help: "会话管理", Meta: kit.Dict(
			"detail", []string{"选择", "编辑", "删除"},
		), Action: map[string]*ice.Action{
			"choice": {Name: "choice", Help: "选择", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(_tmux, "switch-client", "-t", m.Option(SESSION))
				if m.Option(WINDOW) != "" {
					m.Cmd(_tmux, "select-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW))
				}
				if m.Option(PANE) != "" {
					m.Cmd(_tmux, "select-pane", "-t", m.Option(SESSION)+":"+m.Option(WINDOW)+"."+m.Option(PANE))
				}
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case SESSION:
					// 重命名会话
					m.Cmd(_tmux, "rename-session", "-t", m.Option(SESSION), arg[1])
				case WINDOW:
					// 重命名窗口
					m.Cmd(_tmux, "rename-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW), arg[1])
				}
			}},
			gdb.CLEAR: {Name: "clear", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(PANE) != "" {
					m.Cmd(_tmux, "kill-pane", "-t", m.Option(SESSION)+":"+m.Option(WINDOW)+"."+m.Option(PANE))
					return
				}
				if m.Option(WINDOW) != "" {
					m.Cmd(_tmux, "kill-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW))
					return
				}
				if m.Option(SESSION) != "" {
					m.Cmd(_tmux, "kill-session", "-t", m.Option(SESSION))
					return
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			target := ""
			if len(arg) > 3 {
				// 执行命令
				target = arg[0] + ":" + arg[1] + "." + arg[2]
				m.Cmd(_tmux, "send-keys", "-t", target, strings.Join(arg[3:], " "), "Enter")
				time.Sleep(1 * time.Second)
			}
			if len(arg) > 2 {
				// 终端内容
				m.Echo(strings.TrimSpace(m.Cmdx(VIEW, target)))
				return
			}

			if len(arg) == 0 {
				// 会话列表
				m.Split(m.Cmdx(_tmux, "list-session", "-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")

				m.Table(func(index int, value map[string]string, head []string) {
					for _, k := range []string{"clear"} {
						m.Push(k, m.Cmdx(mdb.RENDER, web.RENDER.Button, k))
					}
				})
				return
			}

			target = arg[0]
			if m.Cmd(_tmux, "has-session", "-t", target).Append(cli.CMD_CODE) != "0" {
				// 创建会话
				m.Option(cli.CMD_ENV, "TMUX", "")
				m.Option(cli.CMD_DIR, m.Conf(web.DREAM, "meta.path"))
				m.Cmd(_tmux, "new-session", "-ds", arg[0])
				// m.Cmd("auto", arg[0])
			}
			if len(arg) == 1 {
				// 窗口列表
				m.Cmdy(WINDOW, target)
				return
			}

			switch arg[1] {
			case "has":
				m.Cmd(WINDOW, target).Table(func(index int, value map[string]string, head []string) {
					if value["window"] == arg[2] {
						m.Echo("true")
					}
				})
				return
			}

			if target = arg[0] + ":" + arg[1]; m.Cmd(_tmux, "rename-window", "-t", target, arg[1]).Append(cli.CMD_CODE) != "0" {
				// 创建窗口
				m.Cmd(_tmux, "switch-client", "-t", arg[0])
				m.Cmd(_tmux, "new-window", "-t", arg[0], "-dn", arg[1])
			}
			if len(arg) == 2 {
				// 终端列表
				m.Cmdy(PANE, target)
				return
			}
		}},
		WINDOW: {Name: "windows", Help: "窗口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(_tmux, "list-windows", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
		}},
		PANE: {Name: "panes", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(_tmux, "list-panes", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
		}},
		VIEW: {Name: "view", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(_tmux, "capture-pane", "-pt", kit.Select("", arg, 0)).Set(ice.MSG_APPEND)
		}},

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
				if m.Cmd(_tmux, "has-session", "-t", arg[0]).Append(cli.CMD_CODE) == "0" {
					return
				}
				// 创建会话
				m.Option(cli.CMD_ENV, "TMUX", "", "ctx_pod", arg[0], "ctx_dev", m.Conf(cli.RUNTIME, "conf.ctx_dev"))
				m.Option(cli.CMD_DIR, path.Join(m.Conf(web.DREAM, "meta.path"), arg[0]))
				m.Cmd(_tmux, "new-session", "-ds", arg[0])
				return

				// 共享空间
				share, dev := "", kit.Select(m.Conf(cli.RUNTIME, "conf.ctx_dev"), m.Conf(cli.RUNTIME, "host.ctx_self"))
				m.Richs(web.SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
					share = kit.Format(value["share"])
				})

				// 环境变量
				m.Option("cmd_env", "TMUX", "", "ctx_dev", dev, "ctx_share", share)
				m.Option("cmd_dir", path.Join(m.Conf(web.DREAM, "meta.path"), arg[0]))

				if arg[0] != "" && m.Cmd(_tmux, "has-session", "-t", arg[0]).Append("code") != "0" {
					// 创建会话
					m.Cmd(_tmux, "new-session", "-ds", arg[0])
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
					m.Cmdy(_tmux, "send-keys", "-t", arg[0], v, "Enter")
				}

				// 连接参数
				m.Cmdy(_tmux, "send-keys", "-t", arg[0], "export ctx_dev=", dev, "Enter")
				m.Cmdy(_tmux, "send-keys", "-t", arg[0], "export ctx_share=", share, "Enter")

				m.Cmd(web.FAVOR, kit.Select("tmux.auto", arg, 1)).Table(func(index int, value map[string]string, head []string) {
					switch value["type"] {
					case web.TYPE_SHELL:
						// 发送命令
						m.Cmdy(_tmux, "send-keys", "-t", arg[0], value["text"], "Enter")
						time.Sleep(10 * time.Millisecond)
					}
				})

				for _, v := range kit.Simple(m.Optionv("after")) {
					// 后置命令
					m.Cmdy(_tmux, "send-keys", "-t", arg[0], v, "Enter")
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"make": {Name: "make name cmd...", Help: "个性化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			session := m.Conf(cli.RUNTIME, "node.name")
			if arg[1] == "session" {
				session, arg[2], arg = arg[2], arg[0], arg[2:]
			}

			if m.Warn(m.Cmd(_tmux, "has-session", "-t", session).Append("code") != "0", "session miss") {
				// 会话不存在
				return
			}

			if m.Cmdx("session", session, "has", arg[0]) != "" {
				// 窗口已存在
				return
			}

			switch arg[1] {
			case "init":
				m.Cmdx(_tmux, "rename-window", "-t", session, arg[0])
				arg[1], arg = arg[0], arg[1:]
			case "link":
				m.Cmdx(_tmux, "link-window", "-dt", session, "-s", arg[2])
				return
			default:
				m.Cmd(_tmux, "new-window", "-dt", session, "-n", arg[0])
			}

			for _, v := range arg[1:] {
				switch ls := kit.Split(v); ls[1] {
				case "v":
					m.Cmd(_tmux, "split-window", "-h", "-dt", session+":"+arg[0]+"."+ls[0], ls[2:])
				case "u", "split-window":
					m.Cmd(_tmux, "split-window", "-dt", session+":"+arg[0]+"."+ls[0], ls[2:])
				case "k":
					m.Cmd(_tmux, "send-key", "-t", session+":"+arg[0]+"."+ls[0], ls[2:])
				default:
					m.Cmd(_tmux, ls[1], "-t", session+":"+arg[0]+"."+ls[0], ls[2:])
				}
			}
		}},

		"local": {Name: "local name name", Help: "虚拟机", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("web.code.docker.auto", arg[1])
			m.Cmdy(_tmux, "send-keys", "-t", arg[1], "docker exec -it ", arg[1], " bash", "Enter")
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
				m.Cmdy(_tmux, "send-keys", "-t", arg[1], kit.Format("ssh %s@%s", value["username"], value["website"]), "Enter")

				sleep := kit.Duration(m.Conf(cmd, "meta.sleep"))
				for i := 0; i < kit.Int(m.Conf(cmd, "meta.count")); i++ {
					time.Sleep(sleep)
					list := strings.Split(strings.TrimSpace(m.Cmdx(_tmux, "capture-pane", "-p")), "\n")

					if tail := list[len(list)-1]; tail == m.Conf(cmd, "meta.tail.login") {
						// 登录成功
						break
					} else if tail == m.Conf(cmd, "meta.tail.password") {
						// 输入密码
						m.Cmdy(_tmux, "send-keys", "-t", arg[1], value["password"], "Enter")
					} else if tail == m.Conf(cmd, "meta.tail.verify") {
						// 输入密码
						m.Cmdy(_tmux, "send-keys", "-t", arg[1], m.Cmdx("aaa.totp.get", value["text"]), "Enter")
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
}

func init() { code.Index.Register(Index, &web.Frame{}) }
