package node

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type npm struct {
	ice.Code
	require string `name:"require" http:"/require/modules/"`
	list    string `name:"list auto"`
}

func (s npm) Require(m *ice.Message, arg ...string) {
	p := path.Join(ice.USR_MODULES, path.Join(arg...))
	kit.If(!nfs.ExistsFile(m, p), func() { m.Cmd(cli.SYSTEM, "npm", "install", arg[0], kit.Dict(cli.CMD_DIR, ice.USR)) })
	m.RenderDownload(p)
}
func (s npm) List(m *ice.Message) {
	m.Cmdy(nfs.DIR, ice.USR_MODULES)
}

func init() { ice.CodeCtxCmd(npm{}) }
