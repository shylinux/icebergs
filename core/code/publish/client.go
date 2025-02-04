package publish

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

type client struct {
	create string `name:"create origin* name icons"`
	list   string `name:"list list" help:"软件包"`
}

func (s client) Create(m *ice.Message, arg ...string) {
	m.Cmd(web.SPIDE, mdb.CREATE, m.OptionSimple("origin,name,icons"), mdb.TYPE, nfs.REPOS)
}
func (s client) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		m.Cmd(web.SPIDE).Table(func(value ice.Maps) {
			if value[web.CLIENT_TYPE] == nfs.REPOS {
				m.PushRecord(value, mdb.ICONS, web.CLIENT_NAME)
			}
		})
		m.Action(s.Create)
		m.Display("")
	} else {
		m.SplitIndex(m.Cmdx(web.SPIDE, arg[0], "/c/"+m.Prefix("server"))).PushAction(s.Download)
	}
}
func (s client) Download(m *ice.Message, arg ...string) {
	name := path.Base(m.Option(nfs.PATH))
	web.GoToast(m.Message, func(toast func(string, int, int)) (res []string) {
		m.Cmd(web.SPIDE, m.Option(web.CLIENT_NAME), web.SPIDE_SAVE, nfs.USR_PUBLISH+name, "/publish/"+name, func(count, total, value int) {
			toast(name, count, total)
		})
		return nil
	})
}

func init() { ice.Cmd("web.code.publish.client", client{}) }
