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
				m.Log_IMPORT(FILE, file, SIZE, len(bio.Text()))
				m.Grow(TAIL, kit.Keys(mdb.HASH, h), kit.Dict(
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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TAIL: {Name: TAIL, Help: "日志流", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,id,file,text",
		)},
	}, Commands: map[string]*ice.Command{
		TAIL: {Name: "tail name id auto page filter:text create", Help: "日志流", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Richs(TAIL, "", mdb.FOREACH, func(key string, value ice.Map) {
					value, _ = kit.GetMeta(value), m.Option(mdb.HASH, key)
					m.Cmd(TAIL, mdb.CREATE, kit.SimpleKV("file,name", value))
				})
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
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
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(kit.Slice(arg, 0, 2)), "time,name,count,file", m.Config(mdb.FIELD))
			m.OptionPage(kit.Slice(arg, 2)...)

			mdb.ZoneSelect(m.Spawn(c), arg...).Table(func(index int, value map[string]string, head []string) {
				if strings.Contains(value[mdb.TEXT], m.Option(ice.CACHE_FILTER)) {
					m.Push("", value, head)
				}
			})

			if len(arg) > 0 {
				m.StatusTimeCountTotal(_tail_count(m, arg[0]))
			}
		}},
	}})
}
