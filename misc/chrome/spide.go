package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/wiki"
)

type spide struct {
	cache
	list string `name:"list wid tid url auto insert" help:"节点"`
}

func (s spide) Download(m *ice.Message, arg ...string) {
	m.Cmdy(s.cache, s.Create, arg).ProcessHold()
}
func (s spide) List(m *ice.Message, arg ...string) {
	if len(arg) < 2 || arg[0] == "" || arg[1] == "" {
		s.daemon.List(m, arg...)
		return
	}
	s.send(m, arg[:2], "spide").Tables(func(value ice.Maps) {
		switch value[mdb.TYPE] {
		case wiki.AUDIO:
			m.PushAudios(mdb.SHOW, value[mdb.LINK])
		case wiki.VIDEO:
			m.PushVideos(mdb.SHOW, value[mdb.LINK])
		case wiki.IMG:
			m.PushImages(mdb.SHOW, value[mdb.LINK])
		default:
			m.Push(mdb.SHOW, "")
		}
	}).Cut("show,type,name,link").PushAction(s.Download)
}
func init() { ice.CodeCtxCmd(spide{}) }
