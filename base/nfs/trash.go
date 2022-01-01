package nfs

import (
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _trash_create(m *ice.Message, name string) {
	if s, e := os.Stat(name); e == nil {
		if s.IsDir() {
			tar := path.Base(name) + ".tar.gz"
			m.Cmd(TAR, tar, name)
			name = tar
		}

		if f, e := os.Open(name); m.Assert(e) {
			defer f.Close()

			h := kit.Hashs(f)
			p := path.Join(m.Config(PATH), h[:2], h)
			os.MkdirAll(path.Dir(p), ice.MOD_DIR)
			os.Rename(name, p)
			m.Cmdy(mdb.INSERT, TRASH, "", mdb.HASH, FILE, p, FROM, name)
		}
	}
}

const (
	FROM = "from"
)
const TRASH = "trash"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TRASH: {Name: TRASH, Help: "回收站", Value: kit.Data(
			mdb.SHORT, FROM, mdb.FIELD, "time,hash,file,from", PATH, ice.VAR_TRASH,
		)},
	}, Commands: map[string]*ice.Command{
		TRASH: {Name: "trash hash auto prunes", Help: "回收站", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
				os.Rename(m.Option(FILE), m.Option(FROM))
				m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				os.Remove(m.Option(FILE))
				m.Cmd(mdb.DELETE, TRASH, "", mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunes(m, func(value map[string]string) bool {
					os.Remove(value[FILE])
					return false
				})
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || m.Length() > 0 {
				m.PushAction(mdb.REVERT, mdb.REMOVE)
				return
			}
			_trash_create(m, arg[0])
		}},
	}})
}
