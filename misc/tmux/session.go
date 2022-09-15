package tmux

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _tmux_key(arg ...string) string {
	if len(arg) > 2 {
		return arg[0] + ice.DF + arg[1] + ice.PT + arg[2]
	} else if len(arg) > 1 {
		return arg[0] + ice.DF + arg[1]
	} else if len(arg) > 0 {
		return arg[0]
	} else {
		return "miss"
	}
}
func _tmux_cmd(m *ice.Message, arg ...string) *ice.Message {
	return m.Cmd(cli.SYSTEM, TMUX, arg)
}
func _tmux_cmds(m *ice.Message, arg ...string) string {
	return _tmux_cmd(m, arg...).Result()
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
const (
	NEW_SESSION  = "new-session"
	NEW_WINDOW   = "new-window"
	LINK_WINDOW  = "link-window"
	SPLIT_WINDOW = "split-window"

	KILL_PANE    = "kill-pane"
	KILL_WINDOW  = "kill-window"
	KILL_SESSION = "kill-session"

	RENAME_WINDOW  = "rename-window"
	RENAME_SESSION = "rename-session"
	SWITCH_CLIENT  = "switch-client"
	SELECT_WINDOW  = "select-window"
	SELECT_PANE    = "select-pane"

	LIST_SESSION = "list-session"
	LIST_WINDOWS = "list-windows"
	LIST_PANES   = "list-panes"

	SEND_KEYS    = "send-keys"
	CAPTURE_PANE = "capture-pane"
	ENTER        = "Enter"
)

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
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
	}, Commands: ice.Commands{
		SESSION: {Name: "session session window pane cmd auto", Help: "会话管理", Actions: ice.Actions{
			web.DREAM_CREATE: {Name: "dream.create", Help: "梦想", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd("", m.Option(mdb.NAME)).Length() == 0 {
					m.Cmd("", mdb.CREATE)
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
				m.Option(cli.CMD_ENV, "TMUX", "")
				if m.Option(PANE) != "" { // 创建终端
					_tmux_cmd(m, SPLIT_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))

				} else if m.Option(WINDOW) != "" { // 创建终端
					_tmux_cmd(m, SPLIT_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)))

				} else if m.Option(SESSION) != "" { // 创建窗口
					_tmux_cmd(m, NEW_WINDOW, "-d", "-t", m.Option(SESSION), "-n", m.Option(mdb.NAME))

				} else { // 创建会话
					m.Option(cli.CMD_DIR, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))

					ls := kit.Split(m.Option(mdb.NAME), "-")
					name := kit.Select(ls[0], ls, 1)
					_tmux_cmd(m, NEW_SESSION, "-d", "-s", m.Option(mdb.NAME), "-n", name)

					name = _tmux_key(m.Option(mdb.NAME), name)
					_tmux_cmd(m, SPLIT_WINDOW, "-t", kit.Keys(name, "1"), "-p", "40")
					m.Go(func() {
						m.Sleep("1s")
						_tmux_cmd(m, SEND_KEYS, "-t", kit.Keys(name, "2"), "ish_miss_log", ENTER)
						_tmux_cmd(m, SEND_KEYS, "-t", kit.Keys(name, "1"), "vi etc/miss.sh", ENTER)
					})

					_tmux_cmd(m, LINK_WINDOW, "-s", name, "-t", "miss:")
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(PANE) != "" { // 删除终端
					_tmux_cmd(m, KILL_PANE, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))

				} else if m.Option(WINDOW) != "" { // 删除窗口
					_tmux_cmd(m, KILL_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)))

				} else if m.Option(SESSION) != "" { // 删除会话
					_tmux_cmd(m, KILL_SESSION, "-t", m.Option(SESSION))
				}
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case WINDOW: // 重命名窗口
					_tmux_cmd(m, RENAME_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)), arg[1])

				case SESSION: // 重命名会话
					_tmux_cmd(m, RENAME_SESSION, "-t", m.Option(SESSION), arg[1])
				}
			}},
			mdb.SELECT: {Name: "select", Help: "进入", Hand: func(m *ice.Message, arg ...string) {
				_tmux_cmd(m, SWITCH_CLIENT, "-t", m.Option(SESSION))
				if m.Option(WINDOW) != "" { // 切换窗口
					_tmux_cmd(m, SELECT_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)))
				}
				if m.Option(PANE) != "" { // 切换终端
					_tmux_cmd(m, SELECT_PANE, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))
				}
			}},

			code.XTERM: {Name: "xterm", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				code.ProcessXterm(m, []string{mdb.TYPE, "tmux attach -t " + m.Option(SESSION)}, arg...)
			}},
			SCRIPT: {Name: "script name", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SCRIPT, m.Option(mdb.NAME), func(value ice.Maps) {
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
							_tmux_cmd(m, line)
						}
					}
				})
				m.Sleep30ms()
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
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

			} else if len(arg) > 0 { // 窗口列表
				m.Cmdy(WINDOW, arg[0])

			} else { // 会话列表
				m.Split(_tmux_cmd(m, LIST_SESSION, "-F", m.Config(FORMAT)).Result(), m.Config(FIELDS), ice.FS, ice.NL)
			}

			m.Tables(func(value ice.Maps) {
				switch value["tag"] {
				case "1":
					m.PushButton(code.XTERM, "")
				default:
					m.PushButton(code.XTERM, mdb.SELECT, mdb.REMOVE)
				}
			})
		}},
		WINDOW: {Name: "windows", Help: "窗口", Hand: func(m *ice.Message, arg ...string) {
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, LIST_WINDOWS, "-t", kit.Select("", arg, 0), "-F", m.Config(FORMAT)), m.Config(FIELDS), ice.FS, ice.NL)
		}},
		PANE: {Name: "panes", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
			m.Split(_tmux_cmds(m, LIST_PANES, "-t", kit.Select("", arg, 0), "-F", m.Config(FORMAT)), m.Config(FIELDS), ice.FS, ice.NL)
		}},
		VIEW: {Name: "view", Help: "内容", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(_tmux_cmds(m, CAPTURE_PANE, "-p", "-t", kit.Select("", arg, 0)))
		}},
		CMD: {Name: "cmd", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
			_tmux_cmd(m, SEND_KEYS, "-t", arg[0], strings.Join(arg[1:], ice.SP), ENTER)
		}},
	}})
}
