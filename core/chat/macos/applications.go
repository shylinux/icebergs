package macos

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const APPLICATIONS = "applications"

func init() {
	Index.MergeCommands(ice.Commands{
		APPLICATIONS: {Help: "应用", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { FinderAppend(m, APPLICATIONS, m.PrefixKey()) }},
			code.INSTALL: {Hand: func(m *ice.Message, arg ...string) { AppInstall(m, arg[0], arg[1]) }},
			mdb.CREATE:   {Name: "create space index args name icon"},
		}, PodCmdAction(), CmdHashAction("space,index,args"), mdb.ClearOnExitHashAction())},
	})
}
func install(m *ice.Message, cmd, icon, index string, arg ...string) string {
	if icon == "" {
		return ""
	} else if !kit.HasPrefix(icon, nfs.PS, web.HTTP) && !nfs.Exists(m, icon) && !nfs.Exists(m, ice.USR_ICONS+icon, func(p string) { icon = p }) {
		if strings.Contains(kit.Path(""), nfs.USR_LOCAL_WORK) && nfs.Exists(m, path.Join(strings.Split(kit.Path(""), nfs.USR_LOCAL_WORK)[0], ice.USR_ICONS+icon)) {
			icon = ice.USR_ICONS + icon
		} else {
			icon = path.Join(path.Dir(ctx.GetCmdFile(m, m.PrefixKey())), icon)
		}
	}
	name := kit.TrimExt(path.Base(icon), nfs.PNG, nfs.JPG, nfs.JPEG)
	m.Cmd(Prefix(cmd), mdb.CREATE, mdb.NAME, name, mdb.ICON, icon, ctx.INDEX, index, arg)
	return icon
}
func AppInstall(m *ice.Message, icon, index string, arg ...string) string {
	return install(m, APPLICATIONS, icon, index, arg...)
}
