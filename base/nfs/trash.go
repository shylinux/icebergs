package nfs

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
)

func _trash_create(m *ice.Message, name string) {
	if s, e := os.Stat(name); e == nil {
		if s.IsDir() {
			tar := path.Base(name) + ".tar.gz"
			m.Cmd(cli.SYSTEM, "tar", "zcf", tar, name)
			name = tar
		}

		if f, e := os.Open(name); m.Assert(e) {
			defer f.Close()

			h := kit.Hashs(f)
			p := path.Join(m.Conf(TRASH, kit.META_PATH), h[:2], h)
			os.MkdirAll(path.Dir(p), ice.MOD_DIR)
			os.Rename(name, p)
			m.Cmdy(mdb.INSERT, TRASH, "", mdb.LIST, kit.MDB_FILE, p, kit.MDB_FROM, name)
		}
	}
}

const TRASH = "trash"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TRASH: {Name: TRASH, Help: "删除", Value: kit.Data(kit.MDB_PATH, "var/trash")},
		},
		Commands: map[string]*ice.Command{
			TRASH: {Name: "trash file auto", Help: "删除", Action: map[string]*ice.Action{
				"recover": {Name: "recover", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
					os.Rename(m.Option(kit.MDB_FILE), m.Option(kit.MDB_FROM))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,id,file,from")
					m.Cmdy(mdb.SELECT, TRASH, "", mdb.LIST)
					m.PushAction("recover")
					return
				}
				_trash_create(m, arg[0])
			}},
		},
	})
}