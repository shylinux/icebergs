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
			web.DREAM: {Name: "dream name=hi repos", Help: "空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START, arg)
			}},
			"script": {Name: "script file=hi/hi.js text=", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.TEXT, strings.TrimSpace(m.Option(mdb.TEXT)))
				m.Cmdy(TEMPLATE, nfs.DEFS)
			}},
			"website": {Name: "script file=hi.iml text=", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.FILE, path.Join("website", m.Option(nfs.FILE)))
				m.Option(mdb.TEXT, strings.TrimSpace(m.Option(mdb.TEXT)))
				m.Cmdy(TEMPLATE, nfs.DEFS)
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
				}
			}},

			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN); !cli.IsSuccess(msg) {
					_inner_make(m, msg)
				} else {
					m.Cmd(UPGRADE, cli.RESTART)
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
