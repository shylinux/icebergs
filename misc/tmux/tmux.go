package tmux

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

const (
	TEXT    = "text"
	BUFFER  = "buffer"
	SCRIPT  = "script"
	SESSION = "session"
	WINDOW  = "window"
	PANE    = "pane"
	VIEW    = "view"
)

const TMUX = "tmux"

var Index = &ice.Context{Name: TMUX, Help: "工作台",
	Configs: map[string]*ice.Config{
		TMUX: {Name: TMUX, Help: "服务", Value: kit.Data(
			"source", "https://github.com/tmux/tmux/releases/download/3.1b/tmux-3.1b.tar.gz",
		)},
		BUFFER: {Name: BUFFER, Help: "缓存", Value: kit.Data()},
		SCRIPT: {Name: SCRIPT, Help: "脚本", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,type,name,text",
		)},
		SESSION: {Name: SESSION, Help: "会话", Value: kit.Data(
			"format", "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}",
			"fields", "id,tag,session,windows,height,width",
		)},
		WINDOW: {Name: WINDOW, Help: "窗口", Value: kit.Data(
			"format", "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
			"fields", "id,tag,window,panes,height,width",
		)},
		PANE: {Name: PANE, Help: "终端", Value: kit.Data(
			"format", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
			"fields", "id,tag,pane,tty,height,width",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		TMUX: {Name: "tmux port path auto start build download", Help: "服务", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(TMUX, kit.META_SOURCE))
			}},
			gdb.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, gdb.BUILD, m.Conf(TMUX, kit.META_SOURCE))
			}},
			gdb.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv("prepare", func(p string) []string {
					m.Option(cli.CMD_DIR, p)
					return []string{"-S", kit.Path(p, "tmux.socket"), "new-session", "-dn", "miss"}
				})
				m.Cmdy(code.INSTALL, gdb.START, m.Conf(TMUX, kit.META_SOURCE), "bin/tmux")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(TMUX, kit.META_SOURCE)), arg)
		}},

		TEXT: {Name: "text 查看:button save 清空 text:textarea", Help: "文本", Action: map[string]*ice.Action{
			"save": {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] != "" {
					m.Cmd(cli.SYSTEM, TMUX, "set-buffer", arg[0])
				}
				m.Cmdy(TEXT)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			text := m.Cmdx(cli.SYSTEM, TMUX, "show-buffer")
			m.Cmdy(wiki.SPARK, "inner", text)
			m.Cmdy(wiki.IMAGE, "qrcode", text)
			m.Render("")
		}},
		BUFFER: {Name: "buffer name value auto export import", Help: "缓存", Action: map[string]*ice.Action{
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_TEXT:
					m.Cmd(cli.SYSTEM, TMUX, "set-buffer", "-b", m.Option(kit.MDB_NAME), arg[1])
				}
			}},
			mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(BUFFER, mdb.LIST, "")
				m.Conf(BUFFER, kit.Keys(mdb.META, kit.MDB_COUNT), "0")

				m.Cmd(BUFFER).Table(func(index int, value map[string]string, head []string) {
					m.Grow(BUFFER, "", kit.Dict(
						kit.MDB_NAME, value[head[0]], kit.MDB_TEXT, m.Cmdx(cli.SYSTEM, TMUX, "show-buffer", "-b", value[head[0]]),
					))
				})
				m.Cmdy(mdb.EXPORT, m.Prefix(BUFFER), "", mdb.LIST)
			}},
			mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(BUFFER, mdb.LIST, "")
				m.Conf(BUFFER, kit.Keys(mdb.META, kit.MDB_COUNT), "0")

				m.Cmdy(mdb.IMPORT, m.Prefix(BUFFER), "", mdb.LIST)
				m.Grows(BUFFER, "", "", "", func(index int, value map[string]interface{}) {
					m.Cmd(cli.SYSTEM, TMUX, "set-buffer", "-b", value[kit.MDB_NAME], value[kit.MDB_TEXT])
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 1 && arg[1] != "" {
				// 设置缓存
				m.Cmd(cli.SYSTEM, TMUX, "set-buffer", "-b", arg[0], arg[1])
			}
			if len(arg) > 0 {
				// 查看缓存
				m.Echo(m.Cmdx(cli.SYSTEM, TMUX, "show-buffer", "-b", arg[0]))
				return
			}

			// 缓存列表
			for i, v := range kit.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-buffers"), "\n", "\n", "\n") {
				ls := strings.SplitN(v, ": ", 3)
				m.Push(kit.MDB_NAME, ls[0])
				m.Push(kit.MDB_SIZE, ls[1])
				if i < 3 {
					m.Push(kit.MDB_TEXT, m.Cmdx(cli.SYSTEM, TMUX, "show-buffer", "-b", ls[0]))
				} else {
					m.Push(kit.MDB_TEXT, ls[2][1:len(ls[2])-1])
				}
			}
		}},
		SCRIPT: {Name: "script name auto create export import", Help: "脚本", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create type=shell,tmux,vim name=hi text:textarea=pwd", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(SCRIPT), "", mdb.HASH, arg)
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, m.Prefix(SCRIPT), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME), arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.Prefix(SCRIPT), "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
			}},
			mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.EXPORT, m.Prefix(SCRIPT), "", mdb.HASH)
			}},
			mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.IMPORT, m.Prefix(SCRIPT), "", mdb.HASH)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(mdb.FIELDS, kit.Select(m.Conf(SCRIPT, kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
			m.Cmdy(mdb.SELECT, m.Prefix(SCRIPT), "", mdb.HASH, kit.MDB_NAME, arg)
			m.PushAction(mdb.REMOVE)
		}},
		SESSION: {Name: "session session window pane cmd auto create script", Help: "会话管理", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create name", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_ENV, "TMUX", "")
				if m.Option(PANE) != "" {
					m.Cmd(cli.SYSTEM, TMUX, "split-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW)+"."+m.Option(PANE))

				} else if m.Option(WINDOW) != "" {
					m.Cmd(cli.SYSTEM, TMUX, "split-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW))

				} else if m.Option(SESSION) != "" {
					// 创建窗口
					m.Cmd(cli.SYSTEM, TMUX, "new-window", "-t", m.Option(SESSION), "-dn", m.Option("name"))
				} else {
					// 创建会话
					m.Cmd(cli.SYSTEM, TMUX, "new-session", "-ds", m.Option("name"))
				}
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case WINDOW:
					// 重命名窗口
					m.Cmd(cli.SYSTEM, TMUX, "rename-window", "-t", m.Option(SESSION)+":"+m.Option(WINDOW), arg[1])
				case SESSION:
					// 重命名会话
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
				m.Option(mdb.FIELDS, "name,type,text")
				m.Cmdy(mdb.SELECT, SCRIPT, "", mdb.HASH)
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
			if len(arg) > 3 {
				// 执行命令
				target := arg[0] + ":" + arg[1] + "." + arg[2]
				m.Cmd(cli.SYSTEM, TMUX, "send-keys", "-t", target, strings.Join(arg[3:], " "), "Enter")
				m.Sleep("100ms")
			}
			if len(arg) > 2 {
				// 终端内容
				target := arg[0] + ":" + arg[1] + "." + arg[2]
				m.Echo(strings.TrimSpace(m.Cmdx(VIEW, target)))
				return
			}

			if len(arg) == 0 {
				// 会话列表
				m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-session", "-F", m.Conf(m.Prefix(cmd), "meta.format")), m.Conf(m.Prefix(cmd), "meta.fields"), ",", "\n")
				m.Table(func(index int, value map[string]string, head []string) {
					switch value["tag"] {
					case "1":
						m.PushButton("")
					default:
						m.PushButton(mdb.SELECT, mdb.REMOVE)
					}
				})
				return
			}

			if len(arg) == 1 {
				// 窗口列表
				m.Cmdy(WINDOW, arg[0])
				m.PushAction(mdb.SELECT, mdb.REMOVE)
				return
			}

			if len(arg) == 2 {
				// 终端列表
				m.Cmdy(PANE, arg[0]+":"+arg[1])
				m.PushAction(mdb.SELECT, mdb.REMOVE)
				return
			}
		}},
		WINDOW: {Name: "windows", Help: "窗口", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-windows", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(m.Prefix(cmd), "meta.format")), m.Conf(m.Prefix(cmd), "meta.fields"), ",", "\n")
		}},
		PANE: {Name: "panes", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-panes", "-t", kit.Select("", arg, 0),
				"-F", m.Conf(m.Prefix(cmd), "meta.format")), m.Conf(m.Prefix(cmd), "meta.fields"), ",", "\n")
		}},
		VIEW: {Name: "view", Help: "终端", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(cli.SYSTEM, TMUX, "capture-pane", "-pt", kit.Select("", arg, 0)).Set(ice.MSG_APPEND)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
