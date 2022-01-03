package tmux

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _tmux_key(arg ...string) string {
	if len(arg) > 2 {
		return arg[0] + ice.DF + arg[1] + ice.PT + arg[2]
	} else if len(arg) > 1 {
		return arg[0] + ice.DF + arg[1]
	} else {
		return arg[0]
	}
}

const (
	FORMAT = "format"
	FIELDS = "fields"
)

const (
	SESSION = "session"
	WINDOW  = "window"
	PANE    = "pane"
	VIEW    = "view"
	CMD     = "cmd"
)

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SESSION: {Name: SESSION, Help: "会话", Value: kit.Data(
			FORMAT, "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}",
			FIELDS, "id,tag,session,windows,height,width",
		)},
		WINDOW: {Name: WINDOW, Help: "窗口", Value: kit.Data(
			FORMAT, "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
			FIELDS, "id,tag,window,panes,height,width",
		)},
		PANE: {Name: PANE, Help: "终端", Value: kit.Data(
			FORMAT, "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width},#{pane_current_command}",
			FIELDS, "id,tag,pane,tty,height,width,cmd",
		)},
	}, Commands: map[string]*ice.Command{
		SESSION: {Name: "session session window pane cmd auto", Help: "会话管理", Action: map[string]*ice.Action{
			web.DREAM_CREATE: {Name: "dream.create type name", Help: "梦想", Hand: func(m *ice.Message, arg ...string) {
				if kit.IndexOf(m.Cmd(m.PrefixKey()).Appendv(SESSION), m.Option(mdb.NAME)) == -1 {
					m.Cmd(m.PrefixKey(), mdb.CREATE)
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ctx.ACTION) == SCRIPT {
					m.Cmdy(SCRIPT, mdb.INPUTS, arg)
					return
				}
				switch arg[0] {
				case mdb.NAME:
					m.Cmdy(web.DREAM).Cut("name,size,time")
				}
			}},
			mdb.CREATE: {Name: "create name", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_ENV, TMUX, "")
				if m.Option(PANE) != "" { // 创建终端
					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))

				} else if m.Option(WINDOW) != "" { // 创建终端
					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)))

				} else if m.Option(SESSION) != "" { // 创建窗口
					m.Cmdy(cli.SYSTEM, TMUX, "new-window", "-t", m.Option(SESSION), "-dn", m.Option(mdb.NAME))

				} else { // 创建会话
					m.Option(cli.CMD_DIR, path.Join(m.Conf(web.DREAM, kit.Keym(nfs.PATH)), m.Option(mdb.NAME)))
					ls := kit.Split(m.Option(mdb.NAME), "-_")
					name := ls[len(ls)-1]

					m.Cmdy(cli.SYSTEM, TMUX, "new-session", "-ds", m.Option(mdb.NAME), "-n", name)
					name = _tmux_key(m.Option(mdb.NAME), ls[len(ls)-1])

					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", kit.Keys(name, "1"), "-p", "20")
					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", kit.Keys(name, "2"), "-h")

					m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", kit.Keys(name, "3"), "ish_miss_log", "Enter")
					m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", kit.Keys(name, "2"), "ish_miss_space dev ops")
					m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", kit.Keys(name, "1"), "vi etc/miss.sh", "Enter")

					m.Cmdy(cli.SYSTEM, TMUX, "link-window", "-s", name, "-t", "miss:")
				}
				m.ProcessRefresh30ms()
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(PANE) != "" { // 删除终端
					m.Cmd(cli.SYSTEM, TMUX, "kill-pane", "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))

				} else if m.Option(WINDOW) != "" { // 删除窗口
					m.Cmd(cli.SYSTEM, TMUX, "kill-window", "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)))

				} else if m.Option(SESSION) != "" { // 删除会话
					m.Cmd(cli.SYSTEM, TMUX, "kill-session", "-t", m.Option(SESSION))
				}
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case WINDOW: // 重命名窗口
					m.Cmd(cli.SYSTEM, TMUX, "rename-window", "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)), arg[1])

				case SESSION: // 重命名会话
					m.Cmd(cli.SYSTEM, TMUX, "rename-session", "-t", m.Option(SESSION), arg[1])
				}
			}},
			mdb.SELECT: {Name: "select", Help: "进入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.SYSTEM, TMUX, "switch-client", "-t", m.Option(SESSION))
				if m.Option(WINDOW) != "" { // 切换窗口
					m.Cmd(cli.SYSTEM, TMUX, "select-window", "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)))
				}
				if m.Option(PANE) != "" { // 切换终端
					m.Cmd(cli.SYSTEM, TMUX, "select-pane", "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))
				}
			}},

			SCRIPT: {Name: "script name", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SCRIPT, m.Option(mdb.NAME)).Table(func(index int, value map[string]string, head []string) {
					switch value[mdb.TYPE] {
					case "shell":
						for _, line := range kit.Split(value[mdb.TEXT], ice.NL, ice.NL, ice.NL) {
							m.Cmd(CMD, _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)), line)
						}
					case "vim":
						for _, line := range kit.Split(value[mdb.TEXT], ice.NL, ice.NL, ice.NL) {
							m.Cmd(CMD, _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)), line)
						}
					case "tmux":
						for _, line := range kit.Split(value[mdb.TEXT], ice.NL, ice.NL, ice.NL) {
							m.Cmd(cli.SYSTEM, TMUX, line)
						}
					}
				})
				m.Sleep30ms()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Action(SCRIPT)
			if len(arg) > 3 { // 执行命令
				m.Cmd(CMD, _tmux_key(arg[0], arg[1], arg[2]), arg[3:])
			}
			if len(arg) > 2 { // 终端内容
				m.Echo(strings.TrimSpace(m.Cmdx(VIEW, _tmux_key(arg[0], arg[1], arg[2]))))
				return
			}
			m.Action(mdb.CREATE)
			if len(arg) > 1 { // 终端列表
				m.Cmdy(PANE, _tmux_key(arg[0], arg[1]))
				m.PushAction(mdb.SELECT, mdb.REMOVE)
				return
			}
			if len(arg) > 0 { // 窗口列表
				m.Cmdy(WINDOW, arg[0])
				m.PushAction(mdb.SELECT, mdb.REMOVE)
				return
			}

			// 会话列表
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-session", "-F", m.Config(FORMAT)), m.Config(FIELDS), ice.FS, ice.NL)
			m.Table(func(index int, value map[string]string, head []string) {
				switch value["tag"] {
				case "1":
					m.PushButton("")
				default:
					m.PushButton(mdb.SELECT, mdb.REMOVE)
				}
			})
		}},
		WINDOW: {Name: "windows", Help: "窗口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-windows", "-t", kit.Select("", arg, 0), "-F", m.Config(FORMAT)), m.Config(FIELDS), ice.FS, ice.NL)
		}},
		PANE: {Name: "panes", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-panes", "-t", kit.Select("", arg, 0), "-F", m.Config(FORMAT)), m.Config(FIELDS), ice.FS, ice.NL)
		}},
		VIEW: {Name: "view", Help: "内容", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(cli.SYSTEM, TMUX, "capture-pane", "-pt", kit.Select("", arg, 0)).Set(ice.MSG_APPEND)
		}},
		CMD: {Name: "cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", arg[0], strings.Join(arg[1:], ice.SP), "Enter")
			m.Sleep300ms()
		}},
	}})
}
