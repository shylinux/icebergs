package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

func _vimer_list(m *ice.Message, dir string, arg ...string) { // field
	m.Copy(m.Cmd(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, dir, nfs.DIR_DEEP, ice.TRUE)).Cut(nfs.PATH).RenameAppend(nfs.PATH, kit.Select(mdb.NAME, arg, 0)))
}
func _vimer_inputs(m *ice.Message, arg ...string) {
	switch m.Option(ctx.ACTION) {
	case nfs.SCRIPT:
		switch arg[0] {
		case nfs.FILE:
			file, ext := m.Option(nfs.FILE), kit.Ext(m.Option(nfs.FILE))
			for _, t := range []string{nfs.SH, nfs.SHY, nfs.PY, nfs.JS} {
				m.Push(nfs.FILE, strings.ReplaceAll(file, ice.PT+ext, ice.PT+t))
			}
		}
	case web.WEBSITE:
		switch arg[0] {
		case nfs.FILE:
			name := kit.TrimExt(kit.Select("hi.zml", arg, 1), "")
			m.Push(nfs.FILE, name+".zml")
			m.Push(nfs.FILE, name+".iml")
		}
	case web.DREAM:
		m.Cmdy(web.DREAM, mdb.INPUTS, arg)
	case AUTOGEN:
		m.Cmdy(AUTOGEN, mdb.INPUTS, arg)
	case XTERM:
		m.Cmdy(XTERM, mdb.INPUTS, arg)
	default:
		switch arg[0] {
		case ctx.INDEX:
			m.OptionFields(ctx.INDEX)
			m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, kit.Select("", arg, 1), "")
		case nfs.FILE:
			p := kit.Select(nfs.PWD, arg, 1)
			m.Option(nfs.DIR_ROOT, m.Option(nfs.PATH))
			m.Option(nfs.DIR_DEEP, strings.HasPrefix(m.Option(nfs.PATH), ice.SRC))
			m.Cmdy(nfs.DIR, kit.Select(path.Dir(p), p, strings.HasSuffix(p, ice.FS))+ice.PS, nfs.DIR_CLI_FIELDS)
			m.ProcessAgain()
		}
	}
}
func _vimer_go_complete(m *ice.Message, name string, arg ...string) *ice.Message {
	kit.Fetch(kit.Split(m.Cmdx(cli.SYSTEM, GO, "doc", name), ice.NL, ice.NL, ice.NL), func(index int, value string) {
		if ls := kit.Split(value); len(ls) > 1 {
			switch ls[0] {
			case "const", "type", "func", "var":
				if ls[1] == "(" {
					m.Push(mdb.NAME, strings.TrimSpace(ls[5]))
				} else {
					m.Push(mdb.NAME, strings.TrimSpace(ls[1]))
				}
				m.Push(mdb.TEXT, strings.TrimSpace(value))
			}
		}
	})
	return m
}

const (
	VIMER_TEMPLATE = "vimer.template"
	VIMER_COMPLETE = "vimer.complete"
)
const VIMER = "vimer"

func init() {
	Index.MergeCommands(ice.Commands{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 list", Help: "编辑器", Meta: kit.Dict(ice.DisplayLocal("", INNER)), Actions: ice.Actions{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] == "" {
					m.PushSearch(mdb.TYPE, "go", mdb.NAME, "src/main.go", mdb.TEXT, chat.MergeCmd(m, ""))
				}
			}},
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(nfs.CONTENT) == "" {
					m.Option(nfs.CONTENT, gdb.Event(m.Spawn(), VIMER_TEMPLATE).Result())
				}
				m.Cmdy(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			nfs.TRASH: {Name: "trash path", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.TRASH, arg[0])
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN); cli.IsSuccess(msg) {
					m.Cmd(UPGRADE, cli.RESTART)
				} else {
					_inner_make(m, msg)
				}
			}},
			AUTOGEN: {Name: "create name=hi help=示例 type=Zone,Hash,Data,Code main=main.go zone key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			nfs.SCRIPT: {Name: "script file=hi/hi.js", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DEFS, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)), m.Cmdx(TEMPLATE, kit.Ext(m.Option(nfs.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH)))
			}},
			web.WEBSITE: {Name: "website file=hi.zml", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(mdb.TEXT) == "" {
					m.Option(mdb.TEXT, gdb.Event(m.Spawn(), VIMER_TEMPLATE).Result())
				}
				m.Option(nfs.FILE, path.Join(web.WEBSITE, m.Option(nfs.FILE)))
				m.Cmdy(TEMPLATE, nfs.DEFS)
			}},
			web.DREAM: {Name: "dream name=hi repos", Help: "空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START, arg)
			}},
			XTERM: {Name: "xterm type=sh name", Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(XTERM, mdb.CREATE, arg)
			}},
			PUBLISH: {Name: "publish", Help: "发布", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(PUBLISH, ice.CONTEXTS)
			}},

			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.FILE:
					m.Cmdy(COMPLETE, mdb.FOREACH, arg[1], m.Option(ctx.ACTION))
				}
				_vimer_inputs(m, arg...)
			}},
			TEMPLATE: {Name: "template", Help: "模板", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TEMPLATE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},
			COMPLETE: {Name: "complete", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(COMPLETE, kit.Ext(m.Option(mdb.FILE)), m.Option(nfs.FILE), m.Option(nfs.PATH))
			}},

			"listTags": {Name: "listTags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("web.code.vim.tags", "listTags", arg)
			}},

			DEVPACK: {Name: "devpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_DEBUG, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_DEBUG, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(WEBPACK, mdb.REMOVE)
				m.ProcessInner()
				web.ToastSuccess(m)
			}},
			BINPACK: {Name: "binpack", Help: "打包模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(WEBPACK, mdb.CREATE)
				m.Cmdy(AUTOGEN, BINPACK)
				m.ProcessInner()
				web.ToastSuccess(m)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(INNER, arg)
			m.Option("tabs", m.Config("show.tabs"))
			m.Option("plug", m.Config("show.plug"))
			m.Option("exts", m.Config("show.exts"))

			if arg[0] != ctx.ACTION {
				ctx.DisplayLocal(m, "")
				m.Action(nfs.SAVE, COMPILE, AUTOGEN, nfs.SCRIPT, chat.WEBSITE, web.DREAM, XTERM)
			}
		}},
	})
}
func ProcessVimer(m *ice.Message, path, file, line string, arg ...string) {
	ctx.ProcessField(m, Prefix(VIMER), kit.Simple(path, file, line), arg...)
}

const TEMPLATE = "template"
const COMPLETE = "complete"
const NAVIGATE = "navigate"

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
	return

	if strings.TrimSpace(text) == "" {
		m.Push(mdb.TEXT, kit.Simple(data[""]))
		return
	}

	name := kit.Slice(kit.Split(text), -1)[0]
	if name == "" {
		m.Push(mdb.TEXT, kit.Simple(data[""]))
		return
	}

	key := kit.Slice(kit.Split(name, "."), -1)[0]
	if strings.HasSuffix(name, ".") {
		m.Push(mdb.TEXT, kit.Simple(data[key]))
	} else {
		for k := range data {
			if strings.HasPrefix(k, key) {
				m.Push(mdb.TEXT, key)
			}
		}
		list := kit.Simple(data[key])
		for i, v := range list {
			list[i] = "." + v
		}
		m.Push(mdb.TEXT, list)
	}
}

func init() {
	Index.MergeCommands(ice.Commands{COMPLETE: {Name: "complete type name text auto", Help: "补全", Actions: mdb.RenderAction()}})
}
func init() {
	Index.MergeCommands(ice.Commands{TEMPLATE: {Name: "template type name text auto", Help: "模板", Actions: mdb.RenderAction()}})
}
func init() {
	Index.MergeCommands(ice.Commands{NAVIGATE: {Name: "navigate type name text auto", Help: "跳转", Actions: mdb.RenderAction()}})
}
