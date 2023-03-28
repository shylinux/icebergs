package log

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _tail_create(m *ice.Message, arg ...string) {
	h := mdb.HashCreate(m, arg)
	kit.For(kit.Split(m.Option(nfs.FILE)), func(file string) {
		m.Options(cli.CMD_OUTPUT, nfs.Pipe(m, func(text string) { mdb.ZoneInsert(m, h, nfs.FILE, file, nfs.SIZE, len(text), mdb.TEXT, text) }), mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
		m.Cmd(cli.DAEMON, TAIL, "-n", "0", "-f", file)
	})
}

const TAIL = "tail"

func init() {
	Index.MergeCommands(ice.Commands{
		TAIL: {Name: "tail name id auto page create", Help: "日志流", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m.Spawn(ice.OptionFields("name,file"))).Table(func(value ice.Maps) {
					m.Cmd("", mdb.CREATE, kit.SimpleKV("name,file", value))
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.NAME:
					m.Push(arg[0], kit.Split(m.Option(FILE), ice.PS))
				case nfs.FILE:
					m.Cmdy(nfs.DIR, kit.Select(nfs.PWD, arg, 1), nfs.PATH).RenameAppend(nfs.PATH, nfs.FILE).ProcessAgain()
				}
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) { _tail_create(m, arg...) }},
		}, mdb.PageZoneAction(mdb.SHORT, mdb.NAME, mdb.FIELDS, "time,name,file,count", mdb.FIELD, "time,id,file,size,text"))},
	})
}
