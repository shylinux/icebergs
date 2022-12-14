package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type daemon struct {
	reload string `name:"reload" help:"刷新"`
	list   string `name:"list wid tid url auto" help:"操作"`
}

func (s daemon) send(m *ice.Message, arg ...ice.Any) *ice.Message {
	return m.Cmdy(web.SPACE, "chrome", "chrome", arg)
}

func (s daemon) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case mdb.ZONE:
		s.send(m.Spawn()).Tables(func(value ice.Maps) {
			s.send(m.Spawn(), value[WID]).Tables(func(value ice.Maps) { m.Push(mdb.ZONE, kit.ParseURL(value[URL]).Host) })
		}).Sort(mdb.ZONE)
	}
}
func (s daemon) Spide(m *ice.Message, arg ...string) {
	if len(arg) < 2 {
		s.send(m, arg)
		return
	}
	s.send(m, arg[:2], "spide").Tables(func(value ice.Maps) {
		switch value[mdb.TYPE] {
		case wiki.VIDEO:
			m.PushVideos(mdb.SHOW, value[mdb.LINK])
		case wiki.IMG:
			m.PushImages(mdb.SHOW, value[mdb.LINK])
		default:
			m.Push(mdb.SHOW, "")
		}
	}).Cut("show,type,name,link")
}
func (s daemon) Reload(m *ice.Message, arg ...string) {
	s.send(m, arg[:2], "user.reload", ice.TRUE).ProcessHold()
}
func (s daemon) List(m *ice.Message, arg ...string) {
	switch len(arg) {
	case 3:
		s.send(m, arg[:2], "user.jumps", arg[2])
	case 2:
		s.Spide(m, arg...)
		m.Action(s.Reload)
	default:
		s.send(m, arg)
	}
	m.StatusTimeCount()
}

const (
	WID = "wid"
	TID = "tid"
	URL = "url"
)

func init() { ice.CodeCtxCmd(daemon{}) }
