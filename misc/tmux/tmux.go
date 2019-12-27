package tmux

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
	"path"
	"strings"
	"time"
)

type Frame struct {
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "tmux", Help: "终端模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"buffer": {Name: "buffer", Help: "缓存", Value: kit.Data()},
		"session": {Name: "session", Help: "会话", Value: kit.Data(
			"format", "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}",
			"fields", "id,tag,session,windows,height,width",
			"cmd", []interface{}{"cli.system", "tmux", "list-session"},
		)},
		"windows": {Name: "window", Help: "窗口", Value: kit.Data(
			"format", "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
			"fields", "id,tag,window,panes,height,width",
			"cmd", []interface{}{"cli.system", "tmux", "list-windows"},
		)},
		"panes": {Name: "pane", Help: "终端", Value: kit.Data(
			"format", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
			"fields", "id,tag,pane,tty,height,width",
			"cmd", []interface{}{"cli.system", "tmux", "list-panes"},
		)},
		"view": {Name: "pane", Help: "终端", Value: kit.Data(
			"format", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
			"fields", "id,tag,pane,tty,height,width",
			"cmd", []interface{}{"cli.system", "tmux", "list-panes"},
		)},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.GDB_EVENT, "listen", ice.DREAM_START, "cli.tmux.auto")
			if m.Richs(ice.WEB_STORY, "head", "auto.sh", nil) == nil {
				m.Cmd(ice.WEB_STORY, "add", "shell", "auto.sh", m.Cmdx(ice.WEB_SPIDE, "shy", "GET", "/publish/auto.sh"))
			}
			if m.Richs(ice.WEB_FAVOR, nil, ice.FAVOR_TMUX, nil) == nil {
				m.Cmd(ice.WEB_FAVOR, ice.FAVOR_TMUX, ice.TYPE_SHELL, "下载脚本", `curl -s "$ctx_dev/code/zsh?cmd=download&arg=auto.sh" > auto.sh`)
				m.Cmd(ice.WEB_FAVOR, ice.FAVOR_TMUX, ice.TYPE_SHELL, "加载脚本", `source auto.sh`)
			}
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		"buffer": {Name: "buffer", Help: "终端",
			List: kit.List(
				kit.MDB_INPUT, "text", "name", "buffer", "action", "auto",
				kit.MDB_INPUT, "text", "name", "value",
				kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
				kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
			),
			Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 缓存列表
					for i, v := range kit.Split(m.Cmdx("cli.system", "tmux", "list-buffers"), "\n") {
						ls := strings.SplitN(v, ": ", 3)
						m.Push("buffer", ls[0])
						m.Push("size", ls[1])
						if i < 3 {
							m.Push("text", m.Cmdx("cli.system", "tmux", "show-buffer", "-b", ls[0]))
						} else {
							m.Push("text", ls[2][1:])
						}
					}
					return
				}

				if len(arg) > 1 {
					// 设置缓存
					m.Cmd("cli.system", "tmux", "set-buffer", "-b", arg[0], arg[1])
				}

				// 查看缓存
				m.Cmdy("cli.system", "tmux", "show-buffer", "-b", arg[0]).Set("append")
			}},
		"session": {Name: "session", Help: "会话", Meta: kit.Dict("detail", []string{"选择", "运行", "编辑", "删除", "下载"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "session", "action", "auto",
			kit.MDB_INPUT, "text", "name", "window", "action", "auto",
			kit.MDB_INPUT, "text", "name", "pane", "action", "auto",
			kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
			kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{"cli.system", "tmux"}
			if len(arg) > 3 {
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
				m.Split(m.Cmdx(m.Confv(cmd, "meta.cmd"),
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
			m.Cmdy("view", target)
		}},
		"windows": {Name: "window", Help: "窗口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(m.Confv(cmd, "meta.cmd"), "-t", kit.Select("", arg, 0),
				"-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
		}},
		"panes": {Name: "pane", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(m.Confv(cmd, "meta.cmd"), "-t", kit.Select("", arg, 0),
				"-F", m.Conf(cmd, "meta.format")), m.Conf(cmd, "meta.fields"), ",", "\n")
		}},
		"view": {Name: "view", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("cli.system", "tmux", "capture-pane", "-pt", kit.Select("", arg, 0)).Set("append")
		}},
		"auto": {Name: "auto", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			prefix := []string{"cli.system", "tmux"}

			m.Option("cmd_env", "TMUX", "")
			m.Option("cmd_dir", path.Join(m.Conf(ice.WEB_DREAM, "meta.path"), arg[0]))

			// 创建会话
			if m.Cmd(prefix, "has-session", "-t", arg[0]).Append("code") != "0" {
				m.Cmd(prefix, "new-session", "-ds", arg[0])
			}

			m.Richs(ice.WEB_SPACE, nil, arg[0], func(key string, value map[string]interface{}) {
				m.Cmdy(prefix, "send-keys", "-t", arg[0], "export ctx_dev=", kit.Select(m.Conf(ice.CLI_RUNTIME, "host.ctx_dev"), m.Conf(ice.CLI_RUNTIME, "host.ctx_self")), "Enter")
				m.Cmdy(prefix, "send-keys", "-t", arg[0], "export ctx_share=", value["share"], "Enter")
			})

			m.Cmd(ice.WEB_FAVOR, kit.Select("tmux.init", arg, 1)).Table(func(index int, value map[string]string, head []string) {
				switch value["type"] {
				case "shell":
					m.Cmdy(prefix, "send-keys", "-t", arg[0], value["text"], "Enter")
				}
			})
		}},
	},
}

func init() { cli.Index.Register(Index, &Frame{}) }
