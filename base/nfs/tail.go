package nfs

import (
	"bufio"
	"io"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _tail_create(m *ice.Message, arg ...string) {
	h := m.Cmdx(mdb.INSERT, m.Prefix(TAIL), "", kit.MDB_HASH, arg)

	kit.ForEach(kit.Split(m.Option(kit.MDB_FILE), ","), func(file string) {
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
	return m.Conf(TAIL, kit.Keys(kit.MDB_HASH, kit.Hashs(name), kit.MDB_META, kit.MDB_COUNT))

}

const TAIL = "tail"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TAIL: {Name: TAIL, Help: "日志流", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(TAIL, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					value, _ = kit.GetMeta(value), m.Option(kit.MDB_HASH, key)
					m.Cmd(TAIL, mdb.CREATE, kit.MDB_FILE, kit.Format(value[kit.MDB_FILE]), kit.MDB_NAME, kit.Format(value[kit.MDB_NAME]))
				})
			}},
			TAIL: {Name: "tail name id auto create page", Help: "日志流", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create file name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_tail_create(m, arg...)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, TAIL, "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
				"prev": {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) {
					mdb.PrevPage(m, _tail_count(m, arg[0]), kit.Slice(arg, 2)...)
				}},
				"next": {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) {
					mdb.NextPage(m, _tail_count(m, arg[0]), kit.Slice(arg, 2)...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) < 2 || (len(arg) > 1 && arg[1] == ""), kit.Select("time,name,count,name,file", "time,id,file,text", len(arg) > 0 && arg[0] != ""))
				m.Option(mdb.CACHE_FILTER, kit.Select("", arg, 4))
				m.Option(mdb.CACHE_OFFEND, kit.Select("0", arg, 3))
				m.Option(mdb.CACHE_LIMIT, kit.Select("10", arg, 2))

				m.Cmd(mdb.SELECT, TAIL, "", mdb.ZONE, arg).Table(func(index int, value map[string]string, head []string) {
					if strings.Contains(value[kit.MDB_TEXT], m.Option(mdb.CACHE_FILTER)) {
						m.Push("", value, head)
					}
				})

				if len(arg) == 0 {
					m.PushAction(mdb.REMOVE)
				} else {
					m.StatusTimeCount("total", m.Conf(TAIL, kit.Keys(kit.MDB_HASH, kit.Hashs(arg[0]), kit.MDB_META, kit.MDB_COUNT)))
				}
			}},
		},
	})
}
