package code

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

const VIMER = "vimer"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 auto", Help: "编辑器", Meta: kit.Dict(ice.DisplayLocal("", INNER)), Action: map[string]*ice.Action{
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			AUTOGEN: {Name: "create main=src/main.go zone name=hi help type=Zone,Hash,Lists,Data,Code list key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN); !cli.IsSuccess(msg) {
					_inner_make(m, msg)
				} else {
					m.Cmd(UPGRADE, cli.RESTART)
				}
			}},
			"script": {Name: "script file=hi/hi.js text=", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.TEXT, strings.TrimSpace(m.Option(mdb.TEXT)))
				m.Cmdy(TEMPLATE, nfs.DEFS)
			}},
			web.DREAM: {Name: "dream name=hi repos", Help: "空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START, arg)
			}},
			"website": {Name: "script file=hi.zml@key text@key", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.FILE, path.Join("website", m.Option(nfs.FILE)))
				m.Option(mdb.TEXT, strings.TrimSpace(m.Option(mdb.TEXT)))
				m.Cmdy(TEMPLATE, nfs.DEFS)
			}},
			PUBLISH: {Name: "publish", Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(PUBLISH, ice.CONTEXTS)
			}},

			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.PushSearch(mdb.TYPE, "go", mdb.NAME, "src/main.go", mdb.TEXT, m.MergeCmd(""))
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case web.DREAM:
					m.Cmdy(web.DREAM, mdb.INPUTS, arg)
				case "script":
					switch arg[0] {
					case nfs.FILE:
						file, ext := m.Option(nfs.FILE), kit.Ext(m.Option(nfs.FILE))
						for _, t := range []string{nfs.SH, nfs.SHY, nfs.PY, nfs.JS} {
							m.Push(nfs.FILE, strings.ReplaceAll(file, ice.PT+ext, ice.PT+t))
						}
					case mdb.TEXT:
						m.Push(mdb.TEXT, `echo "hello world"`)
						m.Push(mdb.TEXT, `chapter "hi"`)
						m.Push(mdb.TEXT, `print "hello world"`)
						m.Push(mdb.TEXT, `Volcanos("onimport", {help: "导入数据", list:[], _init: function(can, msg, cb, target) {
	msg.Echo("hello world")
	can.onappend.table(can, msg)
	can.onappend.board(can, msg)
}})`)
					}
				case "website":
					switch arg[0] {
					case nfs.FILE:
						m.Push(nfs.FILE, "hi.zml")
						m.Push(nfs.FILE, "hi.iml")
					case mdb.TEXT:
						m.Push(mdb.TEXT, `
hi
	he
		cli.runtime
		cli.system
		hi/hi.sh
		hi/hi.go
		hi/hi.js
`)
						m.Push(mdb.TEXT, `
left
	username
	系统
		命令 index cli.system
		共享 index cli.qrcode
	代码
		趋势 index web.code.git.trend args icebergs action auto 
		状态 index web.code.git.status args icebergs
main
`)
					}
				default:
				}
			}},
			"complete": {Name: "complete", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option("key") {
				case "ice", "*ice":
					m.Push("name", "Message")
					m.Push("name", "Context")
				default:
					if strings.HasSuffix(m.Option("pre"), " index ") {
						m.OptionFields("index")
						m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "")
					} else if strings.HasSuffix(m.Option("pre"), " action ") {
						m.Push("name", "auto")
					} else if strings.HasSuffix(m.Option("pre"), " type ") {
						m.Push("name", "menu")
					} else if strings.HasSuffix(m.Option("pre"), " ") {
						m.Push("name", "index")
						m.Push("name", "action")
						m.Push("name", "args")
						m.Push("name", "type")
					} else if m.Option("pre") == "" {
						m.Push("name", "left")
						m.Push("name", "head")
						m.Push("name", "main")
						m.Push("name", "foot")
					}
				}
			}},

			"unpack": {Name: "unpack", Help: "导出文件", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(BINPACK, mdb.EXPORT)
			}},
			BINPACK: {Name: "binpack", Help: "打包模式", Hand: func(m *ice.Message, arg ...string) {
				if kit.FileExists(ice.SRC_RELEASE) {
					m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_RELEASE, ice.GO_MOD))
					m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_RELEASE, ice.GO_SUM))
				}
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(AUTOGEN, BINPACK)
				m.ToastSuccess()
				m.ProcessInner()
			}},
			DEVPACK: {Name: "devpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				if kit.FileExists(ice.SRC_DEBUG) {
					m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_DEBUG, ice.GO_MOD))
					m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_DEBUG, ice.GO_SUM))
				}
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(WEBPACK, mdb.REMOVE)
				m.ToastSuccess()
				m.ProcessInner()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Cmdy(INNER, arg) }},
	}})
}
