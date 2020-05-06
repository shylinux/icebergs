package tmux

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
	"time"
)

var Index = &ice.Context{Name: "tmux", Help: "工作台",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"prefix": {Name: "prefix", Help: "前缀", Value: kit.Data("cmd", []interface{}{ice.CLI_SYSTEM, "tmux"})},
		"buffer": {Name: "buffer", Help: "缓存", Value: kit.Data()},

		"session": {Name: "session", Help: "会话", Value: kit.Data(
			"format", "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}",
			"fields", "id,tag,session,windows,height,width",
		)},
		"windows": {Name: "window", Help: "窗口", Value: kit.Data(
			"format", "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
			"fields", "id,tag,window,panes,height,width",
		)},
		"panes": {Name: "pane", Help: "终端", Value: kit.Data(
			"format", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
			"fields", "id,tag,pane,tty,height,width",
		)},
		"view": {Name: "pane", Help: "终端", Value: kit.Data()},

		"local": {Name: "local", Help: "虚拟机", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		"relay": {Name: "relay", Help: "跳板机", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME,
			"count", 100, "sleep", "100ms", "tail", kit.Dict(
				"verify", "Verification code:",
				"password", "Password:",
				"login", "[relay ~]$",
			),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.WEB_PROXY, "add", "tmux", m.AddCmd(&ice.Command{Name: "proxy", Help: "代理", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd("session").Table(func(index int, value map[string]string, head []string) {
					if value["tag"] == "1" {
						m.Echo(value["session"])
					}
				})
			}}))
		}},
		ice.CODE_INSTALL: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("cmd_dir", m.Conf("install", "meta.path"))
			m.Cmd(ice.CLI_SYSTEM, "git", "clone", "https://github.com/tmux/tmux")
		}},
		ice.CODE_PREPARE: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("nfs.link", path.Join(os.Getenv("HOME"), ".tmux.conf"), "etc/conf/tmux.conf")
		}},
		ice.CODE_PROJECT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Watch(ice.DREAM_START, m.Prefix("auto"))

			if m.Richs(ice.WEB_FAVOR, nil, "tmux.auto", nil) == nil {
				m.Cmd(ice.WEB_FAVOR, "tmux.auto", ice.TYPE_SHELL, "脚本", `curl $ctx_dev/publish/auto.sh > auto.sh`)
				m.Cmd(ice.WEB_FAVOR, "tmux.auto", ice.TYPE_SHELL, "脚本", `source auto.sh`)
				m.Cmd(ice.WEB_FAVOR, "tmux.auto", ice.TYPE_SHELL, "脚本", `ShyInit && ShyLogin && trap ShyLogout EXIT`)
			}

			for _, v := range []string{"auto.sh", "auto.vim", "auto.tmux"} {
				p := path.Join(m.Conf("web.code.publish", "meta.path"), v)
				if _, e := os.Stat(p); e != nil && os.IsNotExist(e) {
					// 下载脚本
					if h := m.Cmdx(ice.WEB_SPIDE, "shy", "cache", "GET", "/publish/"+v); h != "" {
						m.Cmd(ice.WEB_STORY, ice.STORY_WATCH, h, p)
					}
				}
			}
		}},
		"auto": {Name: "auto dream", Help: "自动化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))

			// 共享空间
			share, dev := "", kit.Select(m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev"), m.Conf(ice.CLI_RUNTIME, "host.ctx_self"))
			m.Richs(ice.WEB_SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
				share = kit.Format(value["share"])
			})

			// 环境变量
			m.Option("cmd_env", "TMUX", "", "ctx_dev", dev, "ctx_share", share)
			m.Option("cmd_dir", path.Join(m.Conf(ice.WEB_DREAM, "meta.path"), arg[0]))

			if arg[0] != "" && m.Cmd(prefix, "has-session", "-t", arg[0]).Append("code") != "0" {
				// 创建会话
				m.Cmd(prefix, "new-session", "-ds", arg[0])
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
				m.Cmdy(prefix, "send-keys", "-t", arg[0], v, "Enter")
			}

			// 连接参数
			m.Cmdy(prefix, "send-keys", "-t", arg[0], "export ctx_dev=", dev, "Enter")
			m.Cmdy(prefix, "send-keys", "-t", arg[0], "export ctx_share=", share, "Enter")

			m.Cmd(ice.WEB_FAVOR, kit.Select("tmux.auto", arg, 1)).Table(func(index int, value map[string]string, head []string) {
				switch value["type"] {
				case ice.TYPE_SHELL:
					// 发送命令
					m.Cmdy(prefix, "send-keys", "-t", arg[0], value["text"], "Enter")
					time.Sleep(10 * time.Millisecond)
				}
			})

			for _, v := range kit.Simple(m.Optionv("after")) {
				// 后置命令
				m.Cmdy(prefix, "send-keys", "-t", arg[0], v, "Enter")
			}
		}},
		"make": {Name: "make name cmd...", Help: "个性化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			session := m.Conf(ice.CLI_RUNTIME, "node.name")
			if arg[1] == "session" {
				session, arg[2], arg = arg[2], arg[0], arg[2:]
			}

			if m.Warn(m.Cmd(prefix, "has-session", "-t", session).Append("code") != "0", "session miss") {
				// 会话不存在
				return
			}

			if m.Cmdx("session", session, "has", arg[0]) != "" {
				// 窗口已存在
				return
			}

			switch arg[1] {
			case "init":
				m.Cmdx(prefix, "rename-window", "-t", session, arg[0])
				arg[1], arg = arg[0], arg[1:]
			case "link":
				m.Cmdx(prefix, "link-window", "-dt", session, "-s", arg[2])
				return
			default:
				m.Cmd(prefix, "new-window", "-dt", session, "-n", arg[0])
			}

			for _, v := range arg[1:] {
				switch ls := kit.Split(v); ls[1] {
				case "v":
					m.Cmd(prefix, "split-window", "-h", "-dt", session+":"+arg[0]+"."+ls[0], ls[2:])
				case "u", "split-window":
					m.Cmd(prefix, "split-window", "-dt", session+":"+arg[0]+"."+ls[0], ls[2:])
				case "k":
					m.Cmd(prefix, "send-key", "-t", session+":"+arg[0]+"."+ls[0], ls[2:])
				default:
					m.Cmd(prefix, ls[1], "-t", session+":"+arg[0]+"."+ls[0], ls[2:])
				}
			}
		}},

		"text": {Name: "text name 保存:button text:textarea", Help: "文本", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			if len(arg) > 1 && arg[1] != "" {
				m.Cmd(prefix, "set-buffer", arg[1])
			}
			m.Cmdy(prefix, "show-buffer").Set(ice.MSG_APPEND)
		}},
		"buffer": {Name: "buffer [buffer=auto [value]] auto", Help: "缓存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			if len(arg) > 1 {
				// 设置缓存
				m.Cmd(prefix, "set-buffer", "-b", arg[0], arg[1])
			}

			if len(arg) > 0 {
				// 查看缓存
				m.Cmdy(prefix, "show-buffer", "-b", arg[0]).Set(ice.MSG_APPEND)
				return
			}

			// 缓存列表
			for i, v := range kit.Split(m.Cmdx(prefix, "list-buffers"), "\n", "\n", "\n") {
				ls := strings.SplitN(v, ": ", 3)
				m.Push("buffer", ls[0])
				m.Push("size", ls[1])
				if i < 3 {
					m.Push("text", m.Cmdx(prefix, "show-buffer", "-b", ls[0]))
				} else {
					m.Push("text", ls[2][1:])
				}
			}
		}},
		"session": {Name: "session [session=auto [window=auto [pane=auto [cmd]]]] auto", Help: "会话", Meta: kit.Dict(
			"detail", []string{"选择", "编辑", "删除", "下载"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))

			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "select", "选择":
					if arg[2] == "session" {
						// 选择会话
						m.Cmd(prefix, "switch-client", "-t", arg[3])
						break
					}

					if m.Cmd(prefix, "switch-client", "-t", m.Option("session")); arg[2] == "window" {
						// 选择窗口
						m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+arg[3])
						break
					}

					if m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+m.Option("window")); arg[2] == "pane" {
						// 选择终端
						m.Cmd(prefix, "select-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[3])
					}
				case "modify", "编辑":
					switch arg[2] {
					case "session":
						// 重命名会话
						m.Cmd(prefix, "rename-session", "-t", arg[4], arg[3])
					case "window":
						// 重命名窗口
						m.Cmd(prefix, "rename-window", "-t", m.Option("session")+":"+arg[4], arg[3])
					}
				case "delete", "删除":
					switch arg[2] {
					case "session":
						// 删除会话
						m.Cmd(prefix, "kill-session", "-t", arg[3])
					case "window":
						// 删除窗口
						m.Cmd(prefix, "kill-window", "-t", m.Option("session")+":"+arg[3])
					case "pane":
						// 删除终端
						m.Cmd(prefix, "kill-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[3])
					}
				}
				return
			}

			if len(arg) == 0 {
				// 会话列表
				m.Split(m.Cmdx(prefix, "list-session", "-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
				return
			}

			target := arg[0]
			if m.Cmd(prefix, "has-session", "-t", target).Append("code") != "0" {
				// 创建会话
				m.Option("cmd_env", "TMUX", "")
				m.Option("cmd_dir", m.Conf(ice.WEB_DREAM, "meta.path"))
				m.Cmd(prefix, "new-session", "-ds", arg[0])
				m.Cmd("auto", arg[0])
			}

			if len(arg) == 1 {
				// 窗口列表
				m.Cmdy("windows", target)
				return
			}
			switch arg[1] {
			case "has":
				m.Cmd("windows", target).Table(func(index int, value map[string]string, head []string) {
					if value["window"] == arg[2] {
						m.Echo("true")
					}
				})
				return
			}

			if target = arg[0] + ":" + arg[1]; m.Cmd(prefix, "rename-window", "-t", target, arg[1]).Append("code") != "0" {
				// 创建窗口
				m.Cmd(prefix, "switch-client", "-t", arg[0])
				m.Cmd(prefix, "new-window", "-t", arg[0], "-dn", arg[1])
			}

			if len(arg) == 2 {
				// 终端列表
				m.Cmdy("panes", target)
				return
			}

			if target = arg[0] + ":" + arg[1] + "." + arg[2]; len(arg) > 3 {
				// 执行命令
				m.Cmd(prefix, "send-keys", "-t", target, strings.Join(arg[3:], " "), "Enter")
				time.Sleep(1 * time.Second)
			}

			// 终端内容
			m.Echo(strings.TrimSpace(m.Cmdx("view", target)))
		}},
		"windows": {Name: "window", Help: "窗口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			m.Split(m.Cmdx(prefix, "list-windows", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
		}},
		"panes": {Name: "pane", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			m.Split(m.Cmdx(prefix, "list-panes", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
		}},
		"view": {Name: "view", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			m.Cmdy(prefix, "capture-pane", "-pt", kit.Select("", arg, 0)).Set(ice.MSG_APPEND)
		}},

		"local": {Name: "local name name", Help: "虚拟机", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			m.Cmd("web.code.docker.auto", arg[1])
			m.Cmdy(prefix, "send-keys", "-t", arg[1], "docker exec -it ", arg[1], " bash", "Enter")
		}},
		"relay": {Name: "relay [name [favor]]", Help: "跳板机", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
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
				m.Cmdy(prefix, "send-keys", "-t", arg[1], kit.Format("ssh %s@%s", value["username"], value["website"]), "Enter")

				sleep := kit.Duration(m.Conf(cmd, "meta.sleep"))
				for i := 0; i < kit.Int(m.Conf(cmd, "meta.count")); i++ {
					time.Sleep(sleep)
					list := strings.Split(strings.TrimSpace(m.Cmdx(prefix, "capture-pane", "-p")), "\n")

					if tail := list[len(list)-1]; tail == m.Conf(cmd, "meta.tail.login") {
						// 登录成功
						break
					} else if tail == m.Conf(cmd, "meta.tail.password") {
						// 输入密码
						m.Cmdy(prefix, "send-keys", "-t", arg[1], value["password"], "Enter")
					} else if tail == m.Conf(cmd, "meta.tail.verify") {
						// 输入密码
						m.Cmdy(prefix, "send-keys", "-t", arg[1], m.Cmdx("aaa.totp.get", value["text"]), "Enter")
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
