package nfs

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"bufio"
	"io"
)

func _tail_create(m *ice.Message, arg ...string) {
	if m.Option(kit.MDB_HASH) == "" {
		m.Option(kit.MDB_HASH, m.Cmdx(mdb.INSERT, TAIL, "", mdb.HASH, arg))
	}

	kit.ForEach(kit.Split(m.Option(kit.MDB_FILE), ","), func(file string) {
		r, w := io.Pipe()
		m.Go(func() {
			for bio := bufio.NewScanner(r); bio.Scan(); {
				m.Log_IMPORT(kit.MDB_FILE, file, kit.MDB_SIZE, len(bio.Text()))
				m.Grow(TAIL, kit.Keys(kit.MDB_HASH, m.Option(kit.MDB_HASH)), kit.Dict(
					kit.MDB_FILE, file, kit.MDB_SIZE, len(bio.Text()), kit.MDB_TEXT, bio.Text(),
				))
			}
		})

		m.Option(cli.CMD_STDOUT, w)
		m.Option(cli.CMD_STDERR, w)
		m.Option(mdb.CACHE_CLEAR_ON_EXIT, "true")
		m.Cmd(cli.DAEMON, TAIL, "-n", "0", "-f", file)
	})
}

const TAIL = "tail"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TAIL: {Name: TAIL, Help: "跟踪", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			TAIL: {Name: "tail hash id auto create", Help: "跟踪", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create file name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_tail_create(m, arg...)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, TAIL, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,count,name,file", kit.Select("time,id,file,text", mdb.DETAIL, len(arg) > 1), len(arg) > 0))

				if m.Cmdy(mdb.SELECT, TAIL, "", mdb.ZONE, arg); len(arg) == 0 {
					m.PushAction(mdb.REMOVE)
				} else if len(arg) == 1 {
					m.Option(ice.MSG_CONTROL, ice.CONTROL_PAGE)
				}
			}},
		},
	})
}
