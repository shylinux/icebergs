package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const JS = "js"

func init() {
	Index.MergeCommands(ice.Commands{
		JS: {Name: "js path auto", Help: "前端", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				ProcessXterm(m, "node", kit.Format(`require("./usr/volcanos/proto.js"), require("./usr/volcanos/publish/client/nodejs/proto.js"), Volcanos.meta._main("%s")`, path.Join(nfs.PS, arg[2], arg[1])))
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if arg[2] == ice.USR_VOLCANOS {
					m.Debug("what %v %v", arg[1], ice.PLUGIN_LOCAL)
					if strings.HasPrefix(arg[1], "plugin/local/") {
						ctx.ProcessCommand(m, kit.Select(ice.CAN_PLUGIN, "web."+strings.Replace(strings.TrimSuffix(strings.TrimPrefix(arg[1], "plugin/local/"), nfs.PT+JS), nfs.PS, nfs.PT, -1)), kit.Simple())
						m.Debug("what %v", m.FormatMeta())
					}
					m.Debug("what %v", m.FormatMeta())
				} else {
					ctx.DisplayBase(m, require(arg[2], arg[1]))
					ctx.ProcessCommand(m, kit.Select(ice.CAN_PLUGIN, ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO))), kit.Simple())
				}
			}},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) { m.Echo(nfs.Template(m, "demo.js")) }},
		}, PlugAction())},
	})
}
