package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _js_show(m *ice.Message, arg ...string) {
	if arg[2] == ice.USR_VOLCANOS {
		if strings.HasPrefix(arg[1], "plugin/local/") {
			ctx.ProcessField(m, kit.Select(ice.CAN_PLUGIN, "web."+strings.Replace(strings.TrimSuffix(strings.TrimPrefix(arg[1], "plugin/local/"), nfs.PT+JS), nfs.PS, nfs.PT, -1)), kit.Simple())
		}
	} else {
		ctx.DisplayBase(m, require(arg[2], arg[1]))
		ctx.ProcessField(m, kit.Select(ice.CAN_PLUGIN, ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO))), kit.Simple())
	}

}

const JS = "js"

func init() {
	Index.MergeCommands(ice.Commands{
		JS: {Name: "js path auto", Help: "前端", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if arg[1] == "main.js" {
					ctx.ProcessField(m, "web.chat.iframe", kit.Simple(web.UserHost(m)))
				} else {
					_js_show(m, arg...)
				}
				return
				ProcessXterm(m, "node", kit.Format(`require("./usr/volcanos/proto.js"), require("./usr/volcanos/publish/client/nodejs/proto.js"), Volcanos.meta._main("%s")`, path.Join(nfs.PS, arg[2], arg[1])))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				_js_show(m, arg...)
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) { m.Echo(nfs.Template(m, "demo.js")) }},
		}, PlugAction())},
	})
}
