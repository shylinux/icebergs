package nfs

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _trash_create(m *ice.Message, name string) {
	if m.Warn(name == "", ice.ErrNotValid, name) {
		return
	}
	s, e := StatFile(m, name)
	if m.Warn(e, ice.ErrNotFound, name) {
		return
	}
	p := path.Join(ice.VAR_TRASH, path.Base(name))
	if !s.IsDir() {
		if f, e := OpenFile(m, name); m.Assert(e) {
			defer f.Close()
			p = path.Join(ice.VAR_TRASH, kit.HashsPath(f))
		}
	}

	MkdirAll(m, path.Dir(p))
	if RemoveAll(m, p); !m.Warn(Rename(m, name, p)) {
		mdb.HashCreate(m, FROM, name, FILE, p)
	}
}

const TRASH = "trash"

func init() {
	Index.MergeCommands(ice.Commands{
		TRASH: {Name: "trash hash auto prunes", Help: "回收站", Actions: ice.MergeAction(ice.Actions{
			mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
				Rename(m, m.Option(FILE), m.Option(FROM))
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.CREATE: {Name: "create path", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				p := path.Join(ice.VAR_TRASH, path.Base(m.Option(PATH)))
				RemoveAll(m, p)
				if MkdirAll(m, path.Dir(p)); !m.Warn(Rename(m, m.Option(PATH), p)) {
					m.Echo(p)
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				Remove(m, m.Option(FILE))
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunes(m, func(value ice.Maps) bool {
					Remove(m, value[FILE])
					return true
				})
			}},
		}, mdb.HashAction(mdb.SHORT, FROM, mdb.FIELD, "time,hash,file,from")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 || !ExistsFile(m, arg[0]) {
				m.PushAction(mdb.REVERT, mdb.REMOVE)
				return
			}
			_trash_create(m, arg[0])
		}},
	})
}
