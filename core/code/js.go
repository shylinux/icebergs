package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const JS = "js"

func init() {
	Index.MergeCommands(ice.Commands{
		JS: {Name: "js path auto", Help: "前端", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				cmds, text := "node", kit.Format(`require("./usr/volcanos/proto.js"), require("./usr/volcanos/publish/client/nodejs/proto.js"), Volcanos.meta._main("%s")`, path.Join(ice.PS, arg[2], arg[1]))
				_xterm_show(m, cmds, text)
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if arg[2] == ice.USR_VOLCANOS {
					if strings.HasPrefix(arg[1], "plugin/local/") {
						ctx.ProcessCommand(m, kit.Select(ice.CAN_PLUGIN, "web."+strings.Replace(kit.TrimExt(strings.TrimPrefix(arg[1], "plugin/local/"), JS), ice.PS, ice.PT, -1)), kit.Simple())
					}
				} else {
					ctx.DisplayBase(m, path.Join("/require", path.Join(arg[2], arg[1])))
					ctx.ProcessCommand(m, kit.Select("can._plugin", ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO))), kit.Simple())
				}
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) { m.Echo(_js_template) }},
		}, PlugAction())},
	})
}

var _js_template = `
Volcanos(chat.ONIMPORT, {_init: function(can, msg) {
	msg.Echo("hello world").Dump(can)
}})
`
