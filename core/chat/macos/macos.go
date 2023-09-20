package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const (
	USR_ICONS = "usr/icons/"
)
const MACOS = "macos"

var Index = &ice.Context{Name: MACOS, Commands: ice.Commands{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
	ice.Info.Load(m).Cmd(FINDER, ice.CTX_INIT)
}}}}

func init() { chat.Index.Register(Index, nil, DESKTOP, APPLICATIONS) }

func Prefix(arg ...string) string { return chat.Prefix(MACOS, kit.Keys(arg)) }

func PodCmdAction(arg ...string) ice.Actions {
	file := kit.FileLine(2, 100)
	return ice.Actions{
		mdb.SELECT: {Name: "list hash auto create", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Spawn()
			mdb.HashSelect(msg, arg...).Sort(mdb.NAME)
			web.PushPodCmd(msg, m.PrefixKey(), arg...)
			has := map[string]bool{}
			msg.Table(func(index int, value ice.Maps, head []string) {
				if !has[value[ctx.INDEX]] {
					has[value[ctx.INDEX]] = true
					m.Push("", value, head)
				}
			})
			m.Display(ctx.FileURI(file))
		}},
	}
}
func CmdHashAction(arg ...string) ice.Actions {
	file := kit.FileLine(2, 100)
	return ice.MergeActions(ice.Actions{
		mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
			switch mdb.HashInputs(m, arg); arg[0] {
			case mdb.NAME:
				m.Cmd(nfs.DIR, USR_ICONS, func(value ice.Maps) { m.Push(arg[0], kit.TrimExt(value[nfs.PATH], nfs.PNG)) })
			case mdb.ICON:
				m.Cmd(nfs.DIR, USR_ICONS, func(value ice.Maps) { m.Push(arg[0], value[nfs.PATH]) })
			}
		}},
		mdb.SELECT: {Name: "list hash auto create", Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Sort(mdb.NAME).Display(ctx.FileURI(file))
		}},
	}, mdb.HashAction(mdb.SHORT, kit.Select("", arg, 0), mdb.FIELD, kit.Select("time,hash,icon,name,text,space,index,args", arg, 1), kit.Slice(arg, 2)))
}
