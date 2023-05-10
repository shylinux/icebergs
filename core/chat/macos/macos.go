package macos

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const (
	USR_ICONS = "usr/icons/"
)
const MACOS = "macos"

var Index = &ice.Context{Name: MACOS, Commands: ice.Commands{ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { ice.Info.Load(m).Cmd(FINDER, ice.CTX_INIT) }}}}

func init() { chat.Index.Register(Index, nil, DESKTOP) }

func Prefix(arg ...string) string { return chat.Prefix(MACOS, kit.Keys(arg)) }

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
			mdb.HashSelect(m, arg...).Sort(mdb.NAME).Display(strings.TrimPrefix(file, ice.Info.Make.Path))
		}},
	}, ctx.CmdAction(), mdb.HashAction(mdb.SHORT, kit.Select("", arg, 0), mdb.FIELD, kit.Select("time,hash,name,icon,text,index,args", arg, 1), kit.Slice(arg, 2)))
}
