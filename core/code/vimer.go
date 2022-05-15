package code

import (
	"path"
	"strings"
	"sync"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

var _cache_mods = map[string]*ice.Message{}
var _cache_lock = sync.Mutex{}

func _vimer_doc(m *ice.Message, mod string, pkg string) *ice.Message {
	_cache_lock.Lock()
	defer _cache_lock.Unlock()

	key := kit.Keys(mod, pkg)
	if msg, ok := _cache_mods[key]; ok && kit.Time(msg.Time("24h")) > kit.Time(m.Time()) {
		return msg
	}

	if mod != "" {
		m.Cmd(cli.SYSTEM, "go", "get", mod)
	}
	if msg := _vimer_go_complete(m.Spawn(), key); msg.Length() > 0 {
		_cache_mods[key] = msg
		return msg
	}
	return nil
}

func _vimer_defs(m *ice.Message, ext string) string {
	defs := kit.Dict(
		nfs.SH, `
_list
`,
		nfs.SHY, `
chapter "hi"
`,
		nfs.PY, `
print "hello world"
`,
		nfs.JS, `
Volcanos("onimport", {help: "导入数据", _init: function(can, msg) {
	msg.Echo("hello world")
	msg.Dump(can)
}})
`,
		nfs.ZML, `
left
	username
	系统
		命令 index cli.system
		共享 index cli.qrcode
	代码
		趋势 index web.code.git.trend args icebergs action auto 
		状态 index web.code.git.status args icebergs
	脚本
		终端 index hi/hi.sh
		文档 index hi/hi.shy
		数据 index hi/hi.py
		后端 index hi/hi.go
		前端 index hi/hi.js
main
`,
		nfs.IML, `
系统
	命令
		cli.system
	环境
		cli.runtime
开发
	模块
		hi/hi.go
	脚本
		hi/hi.sh
		hi/hi.shy
		hi/hi.py
		hi/hi.go
		hi/hi.js
`,
	)
	return kit.Format(defs[ext])
}
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

	case nfs.WEBSITE:
		switch arg[0] {
		case nfs.FILE:
			m.Push(nfs.FILE, "hi.zml")
			m.Push(nfs.FILE, "hi.iml")
		}

	case web.DREAM:
		m.Cmdy(web.DREAM, mdb.INPUTS, arg)
	}
}
func _vimer_complete(m *ice.Message, arg ...string) {
	const (
		PRE = "pre"
		END = "end"
	)

	switch left := kit.Select("", kit.Slice(kit.Split(m.Option(PRE), "\t \n`"), -1), 0); kit.Ext(m.Option(nfs.FILE)) {
	case nfs.SH:
		if m.Option(PRE) == "" {
			// if _cache_bin != nil {
			// 	m.Copy(_cache_bin)
			// 	break
			// }
			// _cache_bin = m

			// m.Push(mdb.NAME, "_list")
			// _vimer_list(m, "/bin")
			// _vimer_list(m, "/sbin")
		}

	case nfs.SHY:
		switch strings.TrimSpace(left) {
		case cli.FG, cli.BG:
			m.Push(mdb.NAME, cli.RED)
			m.Push(mdb.NAME, cli.BLUE)
			m.Push(mdb.NAME, cli.GREEN)

		default:
			switch kit.Select("", kit.Split(m.Option(PRE)), 0) {
			case "field":
				m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "", ice.OptionFields("index,name,text"))
				_vimer_list(m, ice.SRC, ctx.INDEX)

			case "chain":
				m.Push(mdb.NAME, cli.FG)
				m.Push(mdb.NAME, cli.BG)
			}
		}

	case nfs.GO:
		if m.Option(mdb.NAME) == ice.PT {
			switch m.Option(mdb.TYPE) {
			case "msg", "m":
				m.Copy(_vimer_doc(m, "shylinux.com/x/ice", "Message"))
				m.Copy(_vimer_doc(m, "shylinux.com/x/icebergs", "Message"))

			case "ice", "*ice":
				m.Copy(_vimer_doc(m, "shylinux.com/x/ice", ""))

			case "kit":
				m.Copy(_vimer_doc(m, "shylinux.com/x/toolkits", ""))

			default:
				m.Copy(_vimer_doc(m, "", m.Option(mdb.TYPE)))
			}

		} else {
			m.Push(mdb.NAME, "msg")
			m.Push(mdb.NAME, "ice")
		}

	case nfs.JS:
		if m.Option(mdb.NAME) == ice.PT {
			switch m.Option(mdb.TYPE) {
			case "msg":
				m.Cmdy("web.code.vim.tags", "msg").Cut("name,text")
			case "can":
				m.Cmdy("web.code.vim.tags").Cut(mdb.ZONE)
			default:
				m.Cmdy("web.code.vim.tags", strings.TrimPrefix(m.Option(mdb.TYPE), "can.")).Cut("name,text")
			}
		} else {
			m.Push(mdb.NAME, "msg")
			m.Push(mdb.NAME, "can")
		}

	case nfs.ZML:
		switch left {
		case mdb.TYPE:
			m.Push(mdb.NAME, "menu")

		case ctx.INDEX:
			m.Cmdy(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, "", "", ice.OptionFields("index,name,text"))
			_vimer_list(m, ice.SRC, ctx.INDEX)

		case ctx.ACTION:
			m.Push(mdb.NAME, "auto")
			m.Push(mdb.NAME, "push")
			m.Push(mdb.NAME, "open")

		default:
			if strings.HasSuffix(m.Option(PRE), " ") {
				m.Push(mdb.NAME, "index")
				m.Push(mdb.NAME, "action")
				m.Push(mdb.NAME, "args")
				m.Push(mdb.NAME, "type")
			} else if m.Option(PRE) == "" {
				m.Push(mdb.NAME, "head")
				m.Push(mdb.NAME, "left")
				m.Push(mdb.NAME, "main")
				m.Push(mdb.NAME, "foot")
			}
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

const VIMER = "vimer"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 list", Help: "编辑器", Meta: kit.Dict(ice.DisplayLocal("", INNER)), Action: map[string]*ice.Action{
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.CONTENT, kit.Select(_vimer_defs(m, kit.Ext(m.Option(nfs.FILE))), m.Option(nfs.CONTENT)))
				m.Cmdy(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			AUTOGEN: {Name: "create main=src/main.go zone name=hi help=示例 type=Zone,Hash,Lists,Data,Code key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN); cli.IsSuccess(msg) {
					m.Cmd(UPGRADE, cli.RESTART)
				} else {
					_inner_make(m, msg)
				}
			}},
			nfs.SCRIPT: {Name: "script file=hi/hi.js", Help: "脚本", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.TEXT, strings.TrimSpace(kit.Select(_vimer_defs(m, kit.Ext(m.Option(nfs.FILE))), m.Option(mdb.TEXT))))
				m.Cmdy(TEMPLATE, nfs.DEFS)
			}},
			nfs.WEBSITE: {Name: "website file=hi.zml", Help: "网页", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.TEXT, strings.TrimSpace(kit.Select(_vimer_defs(m, kit.Ext(m.Option(nfs.FILE))), m.Option(mdb.TEXT))))
				m.Option(nfs.FILE, path.Join(nfs.WEBSITE, m.Option(nfs.FILE)))
				m.Cmdy(TEMPLATE, nfs.DEFS)
			}},
			web.DREAM: {Name: "dream name=hi repos", Help: "空间", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.DREAM, cli.START, arg)
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
				_vimer_inputs(m, arg...)
			}},
			"complete": {Name: "complete", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				_vimer_complete(m, arg...)
			}},
			"listTags": {Name: "listTags", Help: "索引", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("web.code.vim.tags", "listTags", arg)
			}},

			"unpack": {Name: "unpack", Help: "导出文件", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(BINPACK, mdb.EXPORT)
			}},
			DEVPACK: {Name: "devpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_DEBUG, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_DEBUG, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(WEBPACK, mdb.REMOVE)
				m.ProcessInner()
				m.ToastSuccess()
			}},
			BINPACK: {Name: "binpack", Help: "打包模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.LINK, ice.GO_SUM, path.Join(ice.SRC_RELEASE, ice.GO_SUM))
				m.Cmd(nfs.LINK, ice.GO_MOD, path.Join(ice.SRC_RELEASE, ice.GO_MOD))
				m.Cmdy(nfs.CAT, ice.GO_MOD)
				m.Cmdy(AUTOGEN, BINPACK)
				m.ProcessInner()
				m.ToastSuccess()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Cmdy(INNER, arg) }},
	}})
}
