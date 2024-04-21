package systray

import (
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"

	"github.com/getlantern/systray"
)

type Systray struct {
	ice.Hash
	short  string `data:"name"`
	field  string `data:"time,type,name,text,icons,order,space,index"`
	create string `name:"create type name* text icons order space index"`
	list   string `name:"list name auto"`
}

func (s Systray) Init(m *ice.Message, arg ...string) {
	m = m.Spawn()
	s.Hash.Init(m, arg...).GoSleep("3s", func() {
		opened := false
		m.AdminCmd(web.SPACE).Table(func(value ice.Maps) { kit.If(value[mdb.TYPE] == web.PORTAL, func() { opened = true }) })
		opened = true
		kit.If(!opened, func() { m.Spawn().Opens(m.SpideOrigin(ice.OPS)) })
		m.Go(func() { systray.Run(func() { s.Show(m, arg...) }, func() {}) })
	})
}
func (s Systray) Show(m *ice.Message, arg ...string) {
	title := kit.JoinLine(m.SpideOrigin(ice.OPS), ice.Info.Make.Module, path.Base(kit.Path("")))
	systray.SetIcon([]byte(m.Cmdx(nfs.CAT, ice.SRC_MAIN_ICO)))
	systray.SetTitle(title)
	systray.SetTooltip(title)
	s.List(m).Table(func(value ice.Maps) {
		item := systray.AddMenuItem(value[mdb.NAME], value[mdb.TEXT])
		kit.If(value[mdb.ICONS], func(p string) { kit.If(m.Cmdx(nfs.CAT, p), func(p string) { item.SetIcon([]byte(p)) }) })
		m.Go(func() {
			for _ = range item.ClickedCh {
				if value[ctx.INDEX] == ice.EXIT {
					m.Cmd(ice.EXIT)
					break
				}
				p := m.SpideOrigin(ice.OPS)
				kit.If(value[web.SPACE], func(pod string) { p += web.S(pod) })
				kit.If(value[ctx.INDEX], func(cmd string) { p += web.C(cmd) })
				m.Opens(p)
			}
		})
	})
}
func (s Systray) List(m *ice.Message, arg ...string) *ice.Message {
	if s.Hash.List(m, arg...); len(arg) == 0 {
		if m.Action(s.Create, s.Build).Sort(mdb.ORDER, ice.INT).Length() == 0 {
			kit.For([]string{web.PORTAL, web.DESKTOP, web.DREAM, web.ADMIN, web.VIMER, ice.EXIT}, func(p string) {
				m.Search(p, func(key string, cmd *ice.Command) {
					m.Push(mdb.NAME, kit.Format("%s(%s)", key, cmd.Help)).Push(ctx.INDEX, p)
				})
			})
		}
	}
	return m
}
func (s Systray) Build(m *ice.Message, arg ...string) {
	defer m.ToastProcess()()
	m.Cmdy(cli.SYSTEM, cli.GO, cli.BUILD, "-ldflags", "-w -s -H=windowsgui", "-o", ice.USR_PUBLISH+"ice.exe",
		ice.SRC_MAIN_GO, ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO, ice.SRC_BINPACK_USR_GO)
}

func init() { ice.ChatCtxCmd(Systray{}) }
