package code

import (
	"os"
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

func _js_main_script(m *ice.Message, arg ...string) (res []string) {
	res = append(res, kit.Format(`global.plugin = "%s"`, kit.Path(arg[2], arg[1])))
	if _, e := nfs.DiskFile.StatFile("usr/volcanos/proto.js"); e == nil {
		res = append(res, kit.Format(`require("%s")`, kit.Path("usr/volcanos/proto.js")))
		res = append(res, kit.Format(`require("%s")`, kit.Path("usr/volcanos/publish/client/nodejs/proto.js")))
	} else {
		for _, file := range []string{"proto.js", "frame.js", "lib/base.js", "lib/core.js", "lib/misc.js", "lib/page.js", "publish/client/nodejs/proto.js"} {
			res = append(res, `_can_name = "`+kit.Path(ice.USR_VOLCANOS, file)+`"`)
			if b, e := nfs.ReadFile(m, path.Join(ice.USR_VOLCANOS, file)); e == nil {
				res = append(res, string(b))
			}
		}
	}
	if _, e := nfs.DiskFile.StatFile(path.Join(arg[2], arg[1])); os.IsNotExist(e) {
		res = append(res, `_can_name = "`+kit.Path(arg[2], arg[1])+`"`)
		if b, e := nfs.ReadFile(m, path.Join(arg[2], arg[1])); e == nil {
			res = append(res, string(b))
		}
	}
	return
}

func _js_exec(m *ice.Message, arg ...string) {
	if m.Option("some") == "run" {
		args := kit.Simple("node", "-e", kit.Join(_js_main_script(m, arg...), ice.NL))
		m.Cmdy(cli.SYSTEM, args)
		m.StatusTime("args", kit.Join([]string{"./bin/ice.bin", "web.code.js.js", "exec", path.Join(arg[2], arg[1])}, " "))
		return
	}

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
}

const JS = "js"
const CSS = "css"
const HTML = "html"
const JSON = "json"
const NODE = "node"

func init() {
	Index.Register(&ice.Context{Name: JS, Help: "前端", Commands: ice.Commands{
		JS: {Name: "js", Help: "前端", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{
					TEMPLATE, COMPLETE, NAVIGATE,
					mdb.PLUGIN, mdb.RENDER, mdb.ENGINE, mdb.SEARCH,
				} {
					m.Cmd(cmd, mdb.CREATE, JSON, m.PrefixKey())
					m.Cmd(cmd, mdb.CREATE, JS, m.PrefixKey())
				}
				LoadPlug(m, JS)
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				if kit.Ext(m.Option(mdb.FILE)) != m.CommandKey() {
					return
				}
				m.Echo(`
Volcanos(chat.ONIMPORT, {help: "导入数据", _init: function(can, msg) {
	msg.Echo("hello world")
	msg.Dump(can)
}})
`)
			}},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == mdb.FOREACH {
					switch m.Option(ctx.ACTION) {
					case nfs.SCRIPT:
						m.Push(nfs.PATH, strings.ReplaceAll(arg[1], ice.PT+kit.Ext(arg[1]), ice.PT+JS))
						m.Cmdy(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, "src/", nfs.DIR_REG, `.*.(sh|py|shy|js)`, nfs.DIR_DEEP, ice.TRUE), nfs.PATH)
					}
					return
				}
				Complete(m, m.Option("text"), kit.Dict(
					"", kit.List("function", "if"),
					"msg", kit.List("Push", "Echo"),
				))
			}},
			NAVIGATE: {Hand: func(m *ice.Message, arg ...string) {
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				key := ctx.GetFileCmd(kit.Replace(path.Join(arg[2], arg[1]), ".js", ".go"))
				if key == "" {
					for p, k := range ice.Info.File {
						if strings.HasPrefix(p, path.Dir(path.Join(arg[2], arg[1]))) {
							key = k
						}
					}
				}
				m.Display(path.Join("/require", path.Join(arg[2], arg[1])))
				ctx.ProcessCommand(m, kit.Select("can.code.inner._plugin", key), kit.Simple())
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				_js_exec(m, arg...)
			}},
			"exec": {Hand: func(m *ice.Message, arg ...string) {
				m.Option("some", "run")
				_js_exec(m, "", arg[0], "")
			}},

			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				_go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				_go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
			}},
		}, PlugAction())},
		NODE: {Name: "node auto download", Help: "前端", Actions: ice.Actions{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(INSTALL, m.Config(nfs.SOURCE))
			}},
		}},
	}, Configs: ice.Configs{
		NODE: {Name: NODE, Help: "前端", Value: kit.Data(
			nfs.SOURCE, "https://nodejs.org/dist/v10.13.0/node-v10.13.0-linux-x64.tar.xz",
		)},
		JS: {Name: JS, Help: "js", Value: kit.Data(INSTALL, kit.List(kit.Dict(
			cli.OSID, cli.ALPINE, ice.CMD, kit.List("apk", "add", "nodejs"),
		)), PLUG, kit.Dict(PREFIX, kit.Dict("// ", COMMENT, "/* ", COMMENT, "* ", COMMENT), PREPARE, kit.Dict(
			KEYWORD, kit.Simple(
				"import", "from", "export",

				"var", "new", "instanceof", "typeof", "let", "const",
				"delete",

				"if", "else", "for", "in", "do", "while", "break", "continue", "switch", "case", "default",
				"try", "throw", "catch", "finally",
				"return",

				"can", "sub", "msg", "res",

				"event", "target", "debugger", "alert",
				"window", "screen", "console", "navigator",
				"location", "history",
				"document",
			),
			CONSTANT, kit.Simple(
				"true", "false",
				"-1", "0", "1", "2", "10",
				"undefined", "null", "NaN",
			),
			FUNCTION, kit.Simple(
				"function", "arguments", "this",
				"shy", "Volcanos", "cb", "cbs",

				"parseInt", "parseFloat",
				"Number", "String", "Boolean",
				"Object", "Array",
				"RegExp", "XMLHttpRequest",
				"Promise",
				"Math", "Date", "JSON",
				"setTimeout",
			),
		), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
