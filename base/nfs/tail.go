package nfs

import (
	"bufio"
	"io"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _tail_create(m *ice.Message, arg ...string) {
	h := m.Cmdx(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)

	kit.ForEach(kit.Split(m.Option(FILE)), func(file string) {
		r, w := io.Pipe()
		m.Go(func() {
			for bio := bufio.NewScanner(r); bio.Scan(); {
				m.Logs(mdb.IMPORT, FILE, file, SIZE, len(bio.Text()))
				mdb.Grow(m, TAIL, kit.Keys(mdb.HASH, h), kit.Dict(
					FILE, file, SIZE, len(bio.Text()), mdb.TEXT, bio.Text(),
				))
			}
		})

		m.Option("cmd_output", w)
		m.Option(mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
		m.Cmd("cli.daemon", TAIL, "-n", "0", "-f", file)
	})
}
func _tail_count(m *ice.Message, name string) string {
	return m.Conf(TAIL, kit.KeyHash(name, kit.Keym(mdb.COUNT)))
}

const TAIL = "tail"

func init() {
	Index.MergeCommands(ice.Commands{
		TAIL: {Name: "tail name id auto page filter:text create", Help: "日志流", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.Richs(m, TAIL, "", mdb.FOREACH, func(key string, value ice.Map) {
					value, _ = kit.GetMeta(value), m.Option(mdb.HASH, key)
					m.Cmd(TAIL, mdb.CREATE, kit.SimpleKV("file,name", value))
				})
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case FILE:
					m.Cmdy(DIR, kit.Select(PWD, arg, 1), PATH).RenameAppend(PATH, FILE)
					m.ProcessAgain()
				case mdb.NAME:
					m.Push(arg[0], kit.Split(m.Option(FILE), ice.PS))
				case mdb.LIMIT:
					m.Push(arg[0], kit.List("10", "20", "30", "50"))
				}
			}},
			mdb.CREATE: {Name: "create file name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_tail_create(m, arg...)
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,id,file,text")), Hand: func(m *ice.Message, arg ...string) {
			m.Fields(len(kit.Slice(arg, 0, 2)), "time,name,count,file", mdb.ZoneField(m))
			mdb.OptionPage(m, kit.Slice(arg, 2)...)

			mdb.ZoneSelect(m.Spawn(), arg...).Table(func(index int, value ice.Maps, head []string) {
				if strings.Contains(value[mdb.TEXT], m.Option(mdb.CACHE_FILTER)) {
					m.Push("", value, head)
				}
			})

			if len(arg) > 0 {
				m.StatusTimeCountTotal(_tail_count(m, arg[0]))
			}
		}},
	})
}
