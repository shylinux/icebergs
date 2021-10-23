package nfs

import (
	"bufio"
	"io"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _tail_create(m *ice.Message, arg ...string) {
	h := m.Cmdx(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)

	kit.ForEach(kit.Split(m.Option(FILE)), func(file string) {
		r, w := io.Pipe()
		m.Go(func() {
			for bio := bufio.NewScanner(r); bio.Scan(); {
				m.Log_IMPORT(kit.MDB_FILE, file, kit.MDB_SIZE, len(bio.Text()))
				m.Grow(TAIL, kit.Keys(kit.MDB_HASH, h), kit.Dict(
					kit.MDB_FILE, file, kit.MDB_SIZE, len(bio.Text()), kit.MDB_TEXT, bio.Text(),
				))
			}
		})

		m.Option(cli.CMD_OUTPUT, w)
		m.Option(cli.CMD_ERRPUT, w)
		m.Option(mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
		m.Cmd(cli.DAEMON, TAIL, "-n", "0", "-f", file)
	})
}
func _tail_count(m *ice.Message, name string) string {
	return m.Conf(TAIL, kit.KeyHash(name, kit.Keym(kit.MDB_COUNT)))
}

const TAIL = "tail"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TAIL: {Name: TAIL, Help: "日志流", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,id,file,text",
		)},
	}, Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(TAIL, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				value, _ = kit.GetMeta(value), m.Option(kit.MDB_HASH, key)
				m.Cmd(TAIL, mdb.CREATE, kit.MDB_FILE, kit.Format(value[kit.MDB_FILE]), kit.MDB_NAME, kit.Format(value[kit.MDB_NAME]))
			})
		}},
		TAIL: {Name: "tail name id auto page filter:text create", Help: "日志流", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case FILE:
					m.Cmdy(DIR, kit.Select("./", arg, 1), PATH).RenameAppend(PATH, FILE)
					m.ProcessAgain()
				case kit.MDB_NAME:
					m.Push(arg[0], kit.Split(m.Option(FILE), ice.PS))
				case kit.MDB_LIMIT:
					m.Push(arg[0], kit.List("10", "20", "30", "50"))
				}
			}},
			mdb.CREATE: {Name: "create file name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_tail_create(m, arg...)
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(kit.Slice(arg, 0, 2)), "time,name,count,file", m.Config(kit.MDB_FIELD))
			m.OptionPage(kit.Slice(arg, 2)...)

			mdb.ZoneSelect(m.Spawn(c), arg...).Table(func(index int, value map[string]string, head []string) {
				if strings.Contains(value[kit.MDB_TEXT], m.Option(ice.CACHE_FILTER)) {
					m.Push("", value, head)
				}
			})

			if len(arg) == 0 {
				m.PushAction(mdb.REMOVE)
			} else {
				m.StatusTimeCountTotal(_tail_count(m, arg[0]))
			}
		}},
	}})
}
