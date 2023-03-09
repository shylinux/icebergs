package tmux

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/ssh"
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
func _tmux_cmd(m *ice.Message, arg ...string) *ice.Message { return m.Cmd(cli.SYSTEM, TMUX, arg) }
func _tmux_cmds(m *ice.Message, arg ...string) string      { return _tmux_cmd(m, arg...).Results() }

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
		SESSION: {Value: kit.Data(
			FORMAT, "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}",
			FIELDS, "id,tag,session,windows,height,width",
		)},
		WINDOW: {Value: kit.Data(
			FORMAT, "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
			FIELDS, "id,tag,window,panes,height,width",
		)},
		PANE: {Value: kit.Data(
			FORMAT, "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width},#{pane_current_command}",
			FIELDS, "id,tag,pane,tty,height,width,cmd",
		)},
	}, Commands: ice.Commands{
		SESSION: {Name: "session session window pane cmds auto", Help: "会话", Actions: ice.MergeActions(ice.Actions{
			web.DREAM_CREATE: {Hand: func(m *ice.Message, arg ...string) { m.Cmd("", mdb.CREATE) }},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.Cmd("", ice.OptionFields(""), func(value ice.Maps) {
						m.PushSearch(mdb.TYPE, ssh.SHELL, mdb.NAME, value[SESSION], mdb.TEXT, "tmux attach -t "+value[SESSION], value)
					})
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ctx.ACTION) == SCRIPT {
					m.Cmdy(SCRIPT, mdb.INPUTS, arg)
					return
				}
				switch arg[0] {
				case mdb.NAME:
					m.Cmdy(web.DREAM).Cut("name,size,time")
				}
			}},
			mdb.CREATE: {Name: "create name", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(cli.CMD_ENV, "TMUX", ""); m.Option(PANE) != "" {
					_tmux_cmd(m, SPLIT_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))
				} else if m.Option(WINDOW) != "" {
					_tmux_cmd(m, SPLIT_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)))
				} else if m.Option(SESSION) != "" {
					_tmux_cmd(m, NEW_WINDOW, "-d", "-t", m.Option(SESSION), "-n", m.Option(mdb.NAME))
				} else {
					m.Option(cli.CMD_DIR, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
					ls := kit.Split(m.Option(mdb.NAME), "-")
					name := kit.Select(ls[0], ls, 1)
					if !cli.IsSuccess(_tmux_cmd(m, NEW_SESSION, "-d", "-s", m.Option(mdb.NAME), "-n", name)) {
						return
					}
					name = _tmux_key(m.Option(mdb.NAME), name)
					if !cli.IsSuccess(_tmux_cmd(m, SPLIT_WINDOW, "-t", kit.Keys(name, "1"), "-p", "40")) {
						return
					}
					m.Go(func() {
						m.Sleep("1s")
						_tmux_cmd(m, SEND_KEYS, "-t", kit.Keys(name, "2"), "ish_miss_log", ENTER)
						_tmux_cmd(m, SEND_KEYS, "-t", kit.Keys(name, "1"), "vi etc/miss.sh", ENTER)
					})
					if m.Cmd(PANE, _tmux_key("miss", name)).Length() == 0 {
						_tmux_cmd(m, LINK_WINDOW, "-s", name, "-t", "miss:")
					}
				}
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(PANE) != "" {
					_tmux_cmd(m, KILL_PANE, "-t", kit.Select(_tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)), m.Option(mdb.ID)))
				} else if m.Option(WINDOW) != "" {
					_tmux_cmd(m, KILL_WINDOW, "-t", kit.Select(_tmux_key(m.Option(SESSION), m.Option(WINDOW)), m.Option(mdb.ID)))
				} else if m.Option(SESSION) != "" {
					_tmux_cmd(m, KILL_SESSION, "-t", m.Option(SESSION))
				}
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case WINDOW:
					_tmux_cmd(m, RENAME_WINDOW, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW)), arg[1])
				case SESSION:
					_tmux_cmd(m, RENAME_SESSION, "-t", m.Option(SESSION), arg[1])
				}
			}},
			mdb.SELECT: {Help: "切入", Hand: func(m *ice.Message, arg ...string) {
				if _tmux_cmd(m, SWITCH_CLIENT, "-t", m.Option(SESSION)); m.Option(WINDOW) != "" {
					_tmux_cmd(m, SELECT_WINDOW, "-t", kit.Select(_tmux_key(m.Option(SESSION), m.Option(WINDOW)), m.Option(mdb.ID)))
				}
				if m.Option(PANE) != "" {
					_tmux_cmd(m, SELECT_PANE, "-t", _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)))
				}
			}},
			code.XTERM: {Help: "切入", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(WINDOW) == "" {
					ctx.ProcessField(m, web.CODE_XTERM, []string{"tmux attach -t " + m.Option(SESSION)}, arg...)
				}
			}},
			SCRIPT: {Name: "script name", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SCRIPT, m.Option(mdb.NAME), func(value ice.Maps) {
					kit.Fetch(kit.SplitLine(value[mdb.TEXT]), func(line string) {
						kit.Switch(value[mdb.TYPE],
							"shell", func() { m.Cmd(CMD, _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)), line) },
							"vim", func() { m.Cmd(CMD, _tmux_key(m.Option(SESSION), m.Option(WINDOW), m.Option(PANE)), line) },
							"tmux", func() { _tmux_cmd(m, line) },
						)
					})
				}).Sleep30ms()
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Action(SCRIPT); len(arg) > 3 {
				m.Cmd(CMD, _tmux_key(arg[0], arg[1], arg[2]), arg[3:])
			}
			if len(arg) > 2 {
				m.Echo(strings.TrimSpace(m.Cmdx(VIEW, _tmux_key(arg[0], arg[1], arg[2]))))
				return
			}
			if m.Action(mdb.CREATE); len(arg) > 1 {
				m.Cmdy(PANE, _tmux_key(arg[0], arg[1]))
			} else if len(arg) > 0 {
				m.Cmdy(WINDOW, arg[0])
			} else {
				m.Split(_tmux_cmd(m, LIST_SESSION, "-F", m.Config(FORMAT)).Result(), m.Config(FIELDS), ice.FS, ice.NL)
			}
			m.Tables(func(value ice.Maps) {
				kit.If(value["tag"] == "1", func() { m.PushButton("") }, func() { m.PushButton(code.XTERM, mdb.SELECT, mdb.REMOVE) })
			}).StatusTimeCount()
		}},
		WINDOW: {Hand: func(m *ice.Message, arg ...string) {
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, LIST_WINDOWS, "-t", kit.Select("", arg, 0), "-F", m.Config(FORMAT)), m.Config(FIELDS), ice.FS, ice.NL)
		}},
		PANE: {Hand: func(m *ice.Message, arg ...string) {
			m.Split(_tmux_cmds(m, LIST_PANES, "-t", kit.Select("", arg, 0), "-F", m.Config(FORMAT)), m.Config(FIELDS), ice.FS, ice.NL)
		}},
		VIEW: {Hand: func(m *ice.Message, arg ...string) {
			m.Echo(_tmux_cmds(m, CAPTURE_PANE, "-p", "-t", kit.Select("", arg, 0)))
		}},
		CMD: {Hand: func(m *ice.Message, arg ...string) {
			_tmux_cmd(m, SEND_KEYS, "-t", arg[0], strings.Join(arg[1:], ice.SP), ENTER)
		}},
	}})
}
