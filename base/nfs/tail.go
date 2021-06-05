package nfs

import (
	"bufio"
	"io"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
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

		m.Option(cli.CMD_OUTPUT, w)
		m.Option(cli.CMD_ERRPUT, w)
		m.Option(mdb.CACHE_CLEAR_ON_EXIT, ice.TRUE)
		m.Cmd(cli.DAEMON, TAIL, "-n", "0", "-f", file)
	})
}

const TAIL = "tail"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TAIL: {Name: TAIL, Help: "日志流", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			TAIL: {Name: "tail hash id limit offend auto prev next create", Help: "日志流", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create file name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_tail_create(m, arg...)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, TAIL, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				"prev": {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) {
					offend := kit.Int(kit.Select("0", arg, 3)) + kit.Int(kit.Select("10", arg, 2))
					total := kit.Int(m.Conf(TAIL, kit.Keys(kit.MDB_HASH, arg[0], kit.MDB_META, kit.MDB_COUNT)))
					if offend >= total {
						offend = total - kit.Int(kit.Select("10", arg, 2))
						m.Toast("已经是最后一页啦!")
					}
					m.ProcessRewrite("offend", offend)
				}},
				"next": {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) {
					offend := kit.Int(kit.Select("0", arg, 3)) - kit.Int(kit.Select("10", arg, 2))
					if offend < 0 {
						offend = 0
						m.Toast("已经是第一页啦!")
					}
					m.ProcessRewrite("offend", offend)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(true, kit.Select("time,hash,count,name,file", "time,id,file,text", len(arg) > 1 && arg[1] != ""))
				m.Option(mdb.CACHE_OFFEND, kit.Select("0", arg, 3))
				m.Option(mdb.CACHE_LIMIT, kit.Select("10", arg, 2))

				if m.Cmdy(mdb.SELECT, TAIL, "", mdb.ZONE, arg); len(arg) == 0 {
					m.PushAction(mdb.REMOVE)
				}
			}},
		},
	})
}
