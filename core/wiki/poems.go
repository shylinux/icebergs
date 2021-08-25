package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	AUTHOR = "author"
	// TITLE = "title"
)
const poems = "poems"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		poems: {Name: "poems", Help: "诗词", Value: kit.Data(
			kit.MDB_SHORT, AUTHOR,
		)},
	}, Commands: map[string]*ice.Command{
		poems: {Name: "poems author title auto insert", Help: "诗词", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert author title content:textarea", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(AUTHOR))
				m.Conf(poems, kit.KeyHash(m.Option(AUTHOR), kit.MDB_META, kit.MDB_SHORT), TITLE)
				m.Cmd(mdb.INSERT, m.PrefixKey(), kit.KeyHash(m.Option(AUTHOR)), mdb.HASH, arg[2:])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Fields(len(arg), "time,author", "time,title,content"); len(arg) == 0 {
				m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH)
				return
			}
			m.Cmdy(mdb.SELECT, m.PrefixKey(), kit.KeyHash(arg[0]), mdb.HASH, AUTHOR, arg[1:])
		}},
	}})
}
