package tmux

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	FORMAT = "format"
	FIELDS = "fields"
)

const (
	SESSION = "session"
	WINDOW  = "window"
	PANE    = "pane"
	VIEW    = "view"
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
			FORMAT, "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
			FIELDS, "id,tag,pane,tty,height,width",
		)},
	}, Commands: map[string]*ice.Command{
		SESSION: {Name: "session session window pane cmd auto create script", Help: "会话管理", Action: map[string]*ice.Action{
			web.DREAM_CREATE: {Name: "dream.create type name", Help: "梦想", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(m.Prefix(SESSION), mdb.CREATE)
			}},
			mdb.CREATE: {Name: "create name", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_ENV, TMUX, "")
				if m.Option(PANE) != "" {
					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW)+"."+m.Option(PANE))

				} else if m.Option(WINDOW) != "" {
					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW))

				} else if m.Option(SESSION) != "" { // 创建窗口
					m.Cmdy(cli.SYSTEM, TMUX, "new-window", "-t", m.Option(SESSION), "-dn", m.Option(kit.MDB_NAME))

				} else { // 创建会话
					m.Option(cli.CMD_DIR, path.Join(m.Conf(web.DREAM, kit.META_PATH), m.Option(kit.MDB_NAME)))
					ls := kit.Split(m.Option(kit.MDB_NAME), "-_")
					name := ls[len(ls)-1]

					m.Cmdy(cli.SYSTEM, TMUX, "new-session", "-ds", m.Option(kit.MDB_NAME), "-n", name)
					name = m.Option(kit.MDB_NAME) + ":" + ls[len(ls)-1]

					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", kit.Keys(name, "1"), "-p", "20")
					m.Cmdy(cli.SYSTEM, TMUX, "split-window", "-t", kit.Keys(name, "2"), "-h")

					m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", kit.Keys(name, "3"), "ish_miss_log", "Enter")
					m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", kit.Keys(name, "2"), "ish_miss_space dev ops")
					m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", kit.Keys(name, "1"), "vi etc/miss.sh", "Enter")

					m.Cmdy(cli.SYSTEM, TMUX, "link-window", "-s", name, "-t", "miss:")
				}
				m.ProcessRefresh30ms()
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case WINDOW: // 重命名窗口
					m.Cmd(cli.SYSTEM, TMUX, "rename-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW), arg[1])

				case SESSION: // 重命名会话
					m.Cmd(cli.SYSTEM, TMUX, "rename-session", "-t", m.Option(SESSION), arg[1])
				}
			}},
			mdb.SELECT: {Name: "select", Help: "进入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.SYSTEM, TMUX, "switch-client", "-t", m.Option(SESSION))
				if m.Option(WINDOW) != "" {
					m.Cmd(cli.SYSTEM, TMUX, "select-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW))
				}
				if m.Option(PANE) != "" {
					m.Cmd(cli.SYSTEM, TMUX, "select-pane", "-t", m.Option(SESSION)+":"+m.Option(WINDOW)+"."+m.Option(PANE))
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(PANE) != "" {
					m.Cmd(cli.SYSTEM, TMUX, "kill-pane", "-t", m.Option(SESSION)+":"+m.Option(WINDOW)+"."+m.Option(PANE))

				} else if m.Option(WINDOW) != "" {
					m.Cmd(cli.SYSTEM, TMUX, "kill-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW))

				} else if m.Option(SESSION) != "" {
					m.Cmd(cli.SYSTEM, TMUX, "kill-session", "-t", m.Option(SESSION))
				}
			}},

			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_NAME:
					m.Option(nfs.DIR_ROOT, m.Conf(web.DREAM, kit.META_PATH))
					m.Cmdy(nfs.DIR, "./", "name size time")
				default:
					m.Option(mdb.FIELDS, "name,type,text")
					m.Cmdy(mdb.SELECT, SCRIPT, "", mdb.HASH)
				}
			}},

			SCRIPT: {Name: "script name", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.SELECT, SCRIPT, "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME)).Table(func(index int, value map[string]string, head []string) {
					switch value[kit.MDB_TYPE] {
					case "shell":
						for _, line := range kit.Split(value[kit.MDB_TEXT], "\n", "\n", "\n") {
							m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", m.Option(SESSION)+":"+m.Option(WINDOW)+"."+m.Option(PANE), line, "Enter")
						}
					case "tmux":
						for _, line := range kit.Split(value[kit.MDB_TEXT], "\n", "\n", "\n") {
							m.Cmd(cli.SYSTEM, TMUX, line)
						}
					case "vim":
					}
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 3 { // 执行命令
				target := arg[0] + ":" + arg[1] + "." + arg[2]
				m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", target, strings.Join(arg[3:], " "), "Enter")
				m.Sleep("100ms")
			}
			if len(arg) > 2 { // 终端内容
				target := arg[0] + ":" + arg[1] + "." + arg[2]
				m.Echo(strings.TrimSpace(m.Cmdx(VIEW, target)))
				return
			}
			if len(arg) == 2 { // 终端列表
				m.Cmdy(PANE, arg[0]+":"+arg[1])
				m.PushAction(mdb.SELECT, mdb.REMOVE)
				return
			}
			if len(arg) == 1 { // 窗口列表
				m.Cmdy(WINDOW, arg[0])
				m.PushAction(mdb.SELECT, mdb.REMOVE)
				return
			}

			// 会话列表
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-session", "-F", m.Conf(m.Prefix(cmd), kit.Keym(FORMAT))), m.Conf(m.Prefix(cmd), kit.Keym(FIELDS)), ",", "\n")
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
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-windows", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(m.Prefix(cmd), kit.Keym(FORMAT))), m.Conf(m.Prefix(cmd), kit.Keym(FIELDS)), ",", "\n")
		}},
		PANE: {Name: "panes", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-panes", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(m.Prefix(cmd), kit.Keym(FORMAT))), m.Conf(m.Prefix(cmd), kit.Keym(FIELDS)), ",", "\n")
		}},
		VIEW: {Name: "view", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(cli.SYSTEM, TMUX, "capture-pane", "-pt", kit.Select("", arg, 0)).Set(ice.MSG_APPEND)
		}},
	}})
}
