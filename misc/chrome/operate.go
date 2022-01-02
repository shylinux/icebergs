package chrome

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

type operate struct {
	reload string `name:"reload" help:"刷新"`
	list   string `name:"list wid tid url auto" help:"操作"`
}

func (o operate) send(m *ice.Message, arg ...interface{}) *ice.Message {
	return m.Cmdy(web.SPACE, "chrome", "chrome", arg)
}

func (o operate) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case mdb.ZONE:
		o.send(m.Spawn()).Table(func(index int, value map[string]string, head []string) {
			o.send(m.Spawn(), value[WID]).Table(func(index int, value map[string]string, head []string) {
				m.Push(mdb.ZONE, kit.ParseURL(value[URL]).Host)
			})
		}).Sort(mdb.ZONE)
	}
}
func (o operate) Spide(m *ice.Message, arg ...string) {
	if len(arg) < 2 {
		o.send(m, arg)
		return
	}
	o.send(m, arg[:2], "spide").Table(func(index int, value map[string]string, head []string) {
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
func (o operate) Reload(m *ice.Message, arg ...string) {
	o.send(m, arg[:2], "user.reload", ice.TRUE).ProcessHold()
}
func (o operate) List(m *ice.Message, arg ...string) {
	switch len(arg) {
	case 3:
		o.send(m, arg[:2], "user.jumps", arg[2])
	case 2:
		m.Action(o.Reload)
		o.Spide(m, arg...)
		m.StatusTimeCount()
	default:
		o.send(m, arg)
	}
}

const (
	WID = "wid"
	TID = "tid"
	URL = "url"
)

func init() { ice.CodeCtxCmd(operate{}) }
