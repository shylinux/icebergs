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

func _js_show(m *ice.Message, arg ...string) {
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
}
func _js_exec(m *ice.Message, arg ...string) {
	args := kit.Simple("node", "-e", kit.Join(_js_main_script(m, arg...), ice.NL))
	m.Cmdy(cli.SYSTEM, args)
	m.StatusTime(ctx.ARGS, kit.Join([]string{"./bin/ice.bin", "web.code.js.js", "exec", path.Join(arg[2], arg[1])}, ice.SP))
}

const JS = "js"
const CSS = "css"
const HTML = "html"
const JSON = "json"

func init() {
	Index.MergeCommands(ice.Commands{
		JS: {Name: "js path auto", Help: "前端", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _js_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _js_exec(m, arg...) }},

			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
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
						m.Cmdy(nfs.DIR, nfs.PWD, kit.Dict(nfs.DIR_ROOT, "src/", nfs.DIR_REG, `.*\.(sh|shy|py|js)$`, nfs.DIR_DEEP, ice.TRUE), nfs.PATH)
					}
					return
				}
				if strings.HasSuffix(m.Option(mdb.TEXT), ice.PT) {
					key := kit.Slice(kit.Split(m.Option(mdb.TEXT), "\t ."), -1)[0]
					switch key {
					case "msg":
						m.Cmdy("web.code.vim.tags", "msg").Cut("name,text")
					case "can":
						m.Cmdy("web.code.vim.tags").Cut(mdb.ZONE)
					default:
						m.Cmdy("web.code.vim.tags", strings.TrimPrefix(m.Option(mdb.TYPE), "can.")).Cut("name,text")
					}
				} else {
					m.Cmdy("web.code.vim.tags").Cut(mdb.ZONE)
				}
			}},
		}, PlugAction(), LangAction())},
	})
}
