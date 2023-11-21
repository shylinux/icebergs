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
	if arg[1] == MAIN_JS {
		ctx.ProcessField(m, web.CHAT_IFRAME, kit.Simple(web.UserHost(m)))
	} else if arg[2] == ice.USR_VOLCANOS {
		if strings.HasPrefix(arg[1], "publish/client/mp/") {
			ctx.ProcessField(m, "web.chat.wx.ide", nil)
		} else if strings.HasPrefix(arg[1], "plugin/local/") {
			ctx.ProcessField(m, kit.Select(ice.CAN_PLUGIN, "web."+strings.Replace(strings.TrimSuffix(strings.TrimPrefix(arg[1], "plugin/local/"), nfs.PT+JS), nfs.PS, nfs.PT, -1)), nil)
		}
	} else {
		ctx.DisplayBase(m, require(arg[2], arg[1]))
		ctx.ProcessField(m, kit.Select(ice.CAN_PLUGIN, ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO))), kit.Simple())
	}
}

const JS = "js"

func init() {
	Index.MergeCommands(ice.Commands{
		JS: {Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _js_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _js_show(m, arg...) }},
			TEMPLATE:   {Hand: func(m *ice.Message, arg ...string) { m.Echo(nfs.Template(m, DEMO_JS)) }},
		}, PlugAction())},
	})
}
