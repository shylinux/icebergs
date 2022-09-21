package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const VIMER = "vimer"

func init() {
	Index.MergeCommands(ice.Commands{
		VIMER: {Name: "vimer path=src/ file=main.go line=1@keyboard list", Help: "编辑器", Actions: ice.Actions{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case AUTOGEN, web.DREAM, XTERM:
					m.Cmdy(m.Option(ctx.ACTION), mdb.INPUTS, arg)
				default:
					switch arg[0] {
					case ctx.INDEX:
						m.OptionFields(ctx.INDEX)
						m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, kit.Select("", arg, 1), "")
					default:
						m.Cmdy(COMPLETE, mdb.FOREACH, arg[1], m.Option(ctx.ACTION))
					}
				}
			}},
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.CONTENT) == "" {
					m.Option(nfs.CONTENT, m.Cmdx(TEMPLATE, m.Option(mdb.TYPE), m.Option(nfs.FILE), m.Option(nfs.PATH)))
				}
				m.Cmdy(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			nfs.TRASH: {Name: "trash path", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.TRASH, arg[0])
			}},
			nfs.SCRIPT: {Name: "script file=hi/hi.js", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx(TEMPLATE, kit.Ext(m.Option(nfs.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH)))
			}},
			web.WEBSITE: {Name: "website file=hi.zml", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.PATH, "src/website/")
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx(TEMPLATE, kit.Ext(m.Option(nfs.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH)))
			}},
			web.DREAM: {Name: "dream name=hi repos", Help: "空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START, arg)
			}},
			XTERM: {Name: "xterm type=sh name text", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(XTERM, mdb.CREATE, arg)
			}},

			TEMPLATE: {Name: "template", Help: "模板", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEMPLATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPLETE: {Name: "complete", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(COMPLETE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},

			"keyboard": {Name: "keyboard", Help: "远程控制", Hand: func(m *ice.Message, arg ...string) {
				hash := m.Cmdx("web.chat.keyboard", mdb.CREATE, "space", m.Option(ice.MSG_DAEMON), "index", m.Option("index"), "input", "")
				link := web.MergePodCmd(m, "", "web.chat.keyboard", mdb.HASH, hash)
				link = tcp.ReplaceLocalhost(m, link)
				m.PushQRCode(mdb.TEXT, link)
				m.Push(mdb.NAME, link)
			}},

			DEVPACK: {Name: "devpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_DEBUG, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_DEBUG, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(WEBPACK, mdb.REMOVE)
				web.ToastSuccess(m)
				m.ProcessInner()
			}},
			BINPACK: {Name: "binpack", Help: "打包模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(WEBPACK, mdb.CREATE)
				m.Cmdy(AUTOGEN, BINPACK)
				web.ToastSuccess(m)
				m.ProcessInner()
			}},
			AUTOGEN: {Name: "create name=hi help=示例 type=Zone,Hash,Data,Code main=main.go zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN); cli.IsSuccess(msg) {
					m.Cmd(UPGRADE, cli.RESTART)
				} else {
					_inner_make(m, nfs.PWD, msg)
				}
			}},
			PUBLISH: {Name: "publish", Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(PUBLISH, ice.CONTEXTS)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdy(INNER, arg); arg[0] != ctx.ACTION {
				m.Action(nfs.SAVE, COMPILE, AUTOGEN, nfs.SCRIPT, chat.WEBSITE, web.DREAM, XTERM)
				m.Option("tabs", m.Config("show.tabs"))
				m.Option("plug", m.Config("show.plug"))
				m.Option("exts", m.Config("show.exts"))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}

const TEMPLATE = "template"
const COMPLETE = "complete"
const NAVIGATE = "navigate"

func init() {
	Index.MergeCommands(ice.Commands{TEMPLATE: {Name: "template type name text auto", Help: "模板", Actions: mdb.RenderAction()}})
	Index.MergeCommands(ice.Commands{COMPLETE: {Name: "complete type name text auto", Help: "补全", Actions: mdb.RenderAction()}})
	Index.MergeCommands(ice.Commands{NAVIGATE: {Name: "navigate type name text auto", Help: "跳转", Actions: mdb.RenderAction()}})
}
func Complete(m *ice.Message, text string, data ice.Map) {
	if strings.HasSuffix(text, ".") {
		key := kit.Slice(kit.Split(text, " ."), -1)[0]
		m.Push(mdb.TEXT, kit.Simple(data[key]))
	} else {
		m.Push(mdb.TEXT, data[""])
		for k := range data {
			m.Push(mdb.TEXT, k)
		}
	}
}
