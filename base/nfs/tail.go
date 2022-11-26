package nfs

import (
	"bufio"
	"io"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _tail_create(m *ice.Message, arg ...string) {
	h := mdb.HashCreate(m, arg)
	kit.Fetch(kit.Split(m.Option(FILE)), func(file string) {
		m.Options("cmd_output", Pipe(m, func(text string) { mdb.ZoneInsert(m, h, FILE, file, SIZE, len(text), mdb.TEXT, text) }), mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
		m.Cmd("cli.daemon", TAIL, "-n", "0", "-f", file)
	})
}

const TAIL = "tail"

func init() {
	Index.MergeCommands(ice.Commands{
		TAIL: {Name: "tail name id auto page create", Help: "日志流", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m.Spawn(ice.OptionFields("name,file"))).Tables(func(value ice.Maps) {
					m.Cmd("", mdb.CREATE, kit.SimpleKV("name,file", value))
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.NAME:
					m.Push(arg[0], kit.Split(m.Option(FILE), ice.PS))
				case FILE:
					m.Cmdy(DIR, kit.Select(PWD, arg, 1), PATH).RenameAppend(PATH, FILE).ProcessAgain()
				}
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) { _tail_create(m, arg...) }},
		}, mdb.PageZoneAction(mdb.SHORT, mdb.NAME, mdb.FIELDS, "time,name,file,count", mdb.FIELD, "time,id,file,size,text"))},
	})
}
func Pipe(m *ice.Message, cb func(string)) io.WriteCloser {
	r, w := io.Pipe()
	m.Go(func() {
		for bio := bufio.NewScanner(r); bio.Scan(); {
			cb(bio.Text())
		}
	})
	return w
}
