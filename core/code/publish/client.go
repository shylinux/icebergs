package publish

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

type client struct {
	list string `name:"list client.name auto" help:"软件包"`
}

func (s client) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		m.Cmd(web.SPIDE).Table(func(value ice.Maps) {
			if value[web.CLIENT_TYPE] == nfs.REPOS {
				m.PushRecord(value, mdb.ICONS, web.CLIENT_NAME)
			}
		})
		m.Display("")
	} else {
		m.SplitIndex(m.Cmdx(web.SPIDE, arg[0], "/c/"+m.Prefix("server")))
		m.PushAction(s.Download)
	}
}
func (s client) Download(m *ice.Message, arg ...string) {
	name := path.Base(m.Option(nfs.PATH))
	m.Cmd(web.SPIDE, m.Option(web.CLIENT_NAME), web.SPIDE_SAVE, nfs.USR+name, "/publish/"+name)
}

func init() { ice.Cmd("web.code.publish.client", client{}) }
