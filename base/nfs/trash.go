package nfs

import (
	"os"
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
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
			m.Cmdy(mdb.INSERT, TRASH, "", mdb.HASH, kit.MDB_FILE, p, kit.MDB_FROM, name)
		}
	}
}
func _trash_prunes(m *ice.Message) {
	m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
}

const TRASH = "trash"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TRASH: {Name: TRASH, Help: "回收站", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_FROM, kit.MDB_PATH, ice.VAR_TRASH,
			)},
		},
		Commands: map[string]*ice.Command{
			TRASH: {Name: "trash file auto prunes", Help: "回收站", Action: map[string]*ice.Action{
				mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
					os.Rename(m.Option(kit.MDB_FILE), m.Option(kit.MDB_FROM))
					m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					os.Remove(m.Option(kit.MDB_FILE))
					m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					_trash_prunes(m)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Fields(len(arg) == 0, "time,hash,file,from")
					m.Cmdy(mdb.SELECT, TRASH, "", mdb.HASH)
					m.PushAction(mdb.REVERT, mdb.REMOVE)
					return
				}
				_trash_create(m, arg[0])
			}},
		},
	})
}
