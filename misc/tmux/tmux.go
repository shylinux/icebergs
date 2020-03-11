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
		"prefix": {Name: "buffer", Help: "缓存", Value: kit.Data("cmd", []interface{}{ice.CLI_SYSTEM, "tmux"})},
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
		"view": {Name: "pane", Help: "终端", Value: kit.Data(
			"format", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
			"fields", "id,tag,pane,tty,height,width",
		)},
		"relay": {Name: "relay", Help: "跳板", Value: kit.Data(
			"tail", kit.Dict(
				"verify", "Verification code:",
				"password", "Password:",
				"login", "[relay ~]$",
			),
		)},
	},
	Commands: map[string]*ice.Command{
		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Watch(ice.DREAM_START, m.Prefix("auto"))

			if m.Richs(ice.WEB_FAVOR, nil, "tmux.auto", nil) == nil {
				m.Cmd(ice.WEB_FAVOR, "tmux.auto", ice.TYPE_SHELL, "脚本", `curl $ctx_dev/publish/auto.sh > auto.sh`)
				m.Cmd(ice.WEB_FAVOR, "tmux.auto", ice.TYPE_SHELL, "脚本", `source auto.sh`)
				m.Cmd(ice.WEB_FAVOR, "tmux.auto", ice.TYPE_SHELL, "脚本", `ShyLogin`)
			}

			for _, v := range []string{"auto.sh", "auto.vim"} {
				p := path.Join(m.Conf("web.code.publish", "meta.path"), v)
				if _, e := os.Stat(p); e != nil && os.IsNotExist(e) {
					// 下载脚本
					if h := m.Cmdx(ice.WEB_SPIDE, "shy", "cache", "GET", "/publish/"+v); h != "" {
						m.Cmd(ice.WEB_STORY, ice.STORY_WATCH, h, p)
					}
				}
			}
		}},
		"auto": {Name: "auto", Help: "自动化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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

			if m.Option("relay") != "" {
				// 自动认证
				m.Cmd("relay", arg[0], m.Option("relay"))
			}

			// 连接参数
			m.Cmdy(prefix, "send-keys", "-t", arg[0], "export ctx_dev=", dev, "Enter")
			m.Cmdy(prefix, "send-keys", "-t", arg[0], "export ctx_share=", share, "Enter")

			m.Cmd(ice.WEB_FAVOR, kit.Select("tmux.auto", arg, 1)).Table(func(index int, value map[string]string, head []string) {
				switch value["type"] {
				case "shell":
					// 发送命令
					m.Cmdy(prefix, "send-keys", "-t", arg[0], value["text"], "Enter")
				}
			})
			for _, v := range kit.Simple(m.Optionv("after")) {
				m.Cmdy(prefix, "send-keys", "-t", arg[0], v, "Enter")
			}
		}},

		"text": {Name: "text", Help: "文本", List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "value", "保存",
			kit.MDB_INPUT, "textarea", "name", "text",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			if len(arg) > 1 && arg[1] != "" {
				m.Cmd(prefix, "set-buffer", arg[1])
			}
			m.Cmdy(prefix, "show-buffer").Set("append")
		}},
		"buffer": {Name: "buffer", Help: "缓存", List: kit.List(
			kit.MDB_INPUT, "text", "name", "buffer", "action", "auto",
			kit.MDB_INPUT, "text", "name", "value",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))
			if len(arg) > 1 {
				// 设置缓存
				m.Cmd(prefix, "set-buffer", "-b", arg[0], arg[1])
			}

			if len(arg) > 0 {
				// 查看缓存
				m.Cmdy(prefix, "show-buffer", "-b", arg[0]).Set("append")
				return
			}

			// 缓存列表
			for i, v := range kit.Split(m.Cmdx(prefix, "list-buffers"), "\n") {
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
		"session": {Name: "session", Help: "会话", Meta: kit.Dict("detail", []string{"选择", "运行", "编辑", "删除", "下载"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "session", "action", "auto",
			kit.MDB_INPUT, "text", "name", "window", "action", "auto",
			kit.MDB_INPUT, "text", "name", "pane", "action", "auto",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))

			if len(arg) > 1 && arg[0] == "action" {
				switch arg[1] {
				case "运行":
					target := ""
					switch arg[2] {
					case "session":
						target = arg[3]
					}
					m.Cmd("auto", target, m.Option("hot"))
					arg = arg[:0]

				case "select":
					if arg[2] == "session" {
						// 选择会话
						m.Cmd(prefix, "switch-client", "-t", arg[3])
						arg = arg[:0]
						break
					}

					if m.Cmd(prefix, "switch-client", "-t", m.Option("session")); arg[2] == "window" {
						// 选择窗口
						m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+arg[3])
						arg = []string{m.Option("session")}
						break
					}

					if m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+m.Option("window")); arg[2] == "pane" {
						// 选择终端
						m.Cmd(prefix, "select-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[3])
						arg = []string{m.Option("session"), m.Option("window")}
					}
				case "modify":
					switch arg[2] {
					case "session":
						// 重命名会话
						m.Cmd(prefix, "rename-session", "-t", arg[4], arg[3])
						arg = arg[:0]
					case "window":
						// 重命名窗口
						m.Cmd(prefix, "rename-window", "-t", m.Option("session")+":"+arg[4], arg[3])
						arg = []string{m.Option("session")}
					}
				case "delete":
					switch arg[2] {
					case "session":
						// 删除会话
						m.Cmd(prefix, "kill-session", "-t", arg[3])
						arg = arg[:0]
					case "window":
						// 删除窗口
						m.Cmd(prefix, "kill-window", "-t", m.Option("session")+":"+arg[3])
						arg = []string{m.Option("session")}
					case "pane":
						// 删除终端
						m.Cmd(prefix, "kill-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[3])
						arg = []string{m.Option("session"), m.Option("window")}
					}
				}
			}

			if len(arg) == 0 {
				// 会话列表
				m.Split(m.Cmdx(prefix, "list-session",
					"-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
				return
			}

			target := arg[0]
			if m.Cmd(prefix, "has-session", "-t", target).Append("code") != "0" {
				// 创建会话
				m.Option("cmd_env", "TMUX", "")
				m.Option("cmd_dir", m.Conf(ice.WEB_DREAM, "meta.path"))
				m.Cmd(prefix, "new-session", "-ds", arg[0])
				m.Cmd("auto", arg[0], arg[1:])
				arg = arg[:1]
			}

			if len(arg) == 1 {
				// 窗口列表
				m.Cmdy("windows", target)
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
			m.Cmdy(prefix, "capture-pane", "-pt", kit.Select("", arg, 0)).Set("append")
		}},

		"relay": {Name: "relay", Help: "跳板", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))

			m.Richs("aaa.auth.auth", nil, kit.Select("relay", arg, 1), func(key string, value map[string]interface{}) {
				m.Cmdy(prefix, "send-keys", "-t", arg[0], kit.Format("ssh %s@%s", value["username"], value["website"]), "Enter")

				for i := 0; i < 10; i++ {
					time.Sleep(100 * time.Millisecond)
					tail := m.Cmdx(prefix, "capture-pane", "-p")
					if strings.HasSuffix(strings.TrimSpace(tail), m.Conf("relay", "meta.tail.login")) {
						for _, v := range kit.Simple(value["init"]) {
							if v != "" {
								m.Cmdy(prefix, "send-keys", "-t", arg[0], v, "Enter")
							}
						}
						break
					}

					if strings.HasSuffix(strings.TrimSpace(tail), m.Conf("relay", "meta.tail.verify")) {
						m.Cmdy(prefix, "send-keys", "-t", arg[0], m.Cmdx("aaa.auth.get", "relay"), "Enter")
						continue
					}

					if strings.HasSuffix(strings.TrimSpace(tail), m.Conf("relay", "meta.tail.password")) {
						m.Cmdy(prefix, "send-keys", "-t", arg[0], value["password"], "Enter")
						continue
					}
				}
			})
		}},
		"/favor": {Name: "/favor", Help: "收藏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// prefix := kit.Simple(m.Confv("prefix", "meta.cmd"))

			// 当前会话
			current := ""
			m.Cmd("session").Table(func(index int, value map[string]string, head []string) {
				if value["tag"] == "1" {
					current = value["session"]
				}
			})

			switch arg = kit.Split(kit.Select("tmux.auto", arg, 0)); arg[0] {
			default:
				m.Cmd("auto", current, arg)
				m.Append("_output", "void")
			}
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
