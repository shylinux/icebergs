package nfs

import (
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
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
			p := path.Join(m.Config(kit.MDB_PATH), h[:2], h)
			os.MkdirAll(path.Dir(p), ice.MOD_DIR)
			os.Rename(name, p)
			m.Cmdy(mdb.INSERT, TRASH, "", mdb.HASH, kit.MDB_FILE, p, kit.MDB_FROM, name)
		}
	}
}

const TRASH = "trash"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TRASH: {Name: TRASH, Help: "回收站", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_FROM, kit.MDB_FIELD, "time,hash,file,from",
			kit.MDB_PATH, ice.VAR_TRASH,
		)},
	}, Commands: map[string]*ice.Command{
		TRASH: {Name: "trash file auto prunes", Help: "回收站", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
				os.Rename(m.Option(kit.MDB_FILE), m.Option(kit.MDB_FROM))
				m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				os.Remove(m.Option(kit.MDB_FILE))
				m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
			}},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...)
				m.PushAction(mdb.REVERT, mdb.REMOVE)
				return
			}
			_trash_create(m, arg[0])
		}},
	}})
}
