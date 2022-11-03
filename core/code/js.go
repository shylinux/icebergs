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

func _js_parse(m *ice.Message, arg ...string) {

}
func _js_show(m *ice.Message, arg ...string) {
	kind := ""
	m.Cmd(nfs.CAT, path.Join(arg[2], arg[1]), func(text string, index int) {
		ls := kit.Split(text, "\t ", ":,()")
		if strings.HasPrefix(text, "Volcanos(") {
			if kind = ls[2]; strings.Contains(text, "_init: ") {
				m.Push("line", index+1)
				m.Push("kind", kind)
				m.Push("name", "_init")
				m.Push("type", "function")
			}
			return
		}
		indent := 0
		for _, c := range text {
			if c == '\t' {
				indent++
			} else if c == ' ' {
				indent++
			} else {
				break
			}
		}
		if len(ls) > 2 && ls[1] == ":" {
			if indent > 1 {
				return
			}
			m.Push("line", index+1)
			m.Push("kind", kind)
			m.Push("name", ls[0])
			m.Push("type", ls[2])
		}
	})
	m.StatusTimeCount()
}
func _js_exec(m *ice.Message, arg ...string) {
	if arg[2] == "usr/volcanos/" && strings.HasPrefix(arg[1], "plugin/local/") {
		key := "web." + strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(arg[1], "plugin/local/"), ".js"), ice.PS, ice.PT)
		ctx.ProcessCommand(m, kit.Select("can.plugin", key), kit.Simple())
		return
	}
	m.Display(path.Join("/require", path.Join(arg[2], arg[1])))
	key := ctx.GetFileCmd(kit.Replace(path.Join(arg[2], arg[1]), ".js", ".go"))
	ctx.ProcessCommand(m, kit.Select("can.code.inner._plugin", key), kit.Simple())
	return
	args := kit.Simple("node", "-e", kit.Join(_js_main_script(m, arg...), ice.NL))
	m.Cmdy(cli.SYSTEM, args).StatusTime(ctx.ARGS, kit.Join(append([]string{ice.ICE_BIN, m.PrefixKey(), m.ActionKey()}, arg...), ice.SP))
}

const JS = "js"
const HTML = "html"
const JSON = "json"

func init() {
	Index.MergeCommands(ice.Commands{
		JS: {Name: "js path auto", Help: "前端", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _js_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _js_exec(m, arg...) }},

			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_js_template) }},
			COMPLETE: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == mdb.FOREACH { // 文件
					switch m.Option(ctx.ACTION) {
					case nfs.SCRIPT:
						m.Push(nfs.PATH, strings.ReplaceAll(arg[1], ice.PT+kit.Ext(arg[1]), ice.PT+JS))
						m.Option(nfs.DIR_REG, kit.FileReg(nfs.SH, nfs.PY, nfs.SHY, nfs.JS))
						nfs.DirDeepAll(m, ice.SRC, nfs.PWD, nil).Cut(nfs.PATH)
					}

				} else if strings.HasSuffix(m.Option(mdb.TEXT), ice.PT) { // 方法
					key := kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t ."), -1)[0]
					switch key {
					case "msg":
						m.Cmdy("web.code.vim.tags", "msg").Cut("name,text")
					case "can":
						m.Cmdy("web.code.vim.tags").Cut(mdb.ZONE)
					default:
						m.Cmdy("web.code.vim.tags", strings.TrimPrefix(m.Option(mdb.TYPE), "can.")).Cut("name,text")
					}

				} else { // 类型
					m.Cmdy("web.code.vim.tags").Cut(mdb.ZONE)
				}
			}},
		}, PlugAction(), LangAction())},
	})
}

var _js_template = `
Volcanos(chat.ONIMPORT, {help: "导入数据", _init: function(can, msg) {
	msg.Echo("hello world").Dump(can)
}})
`
