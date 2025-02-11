package nfs

import (
	"io"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _trash_create(m *ice.Message, from string) {
	if m.WarnNotValid(from == "", FROM) {
		return
	}
	s, e := StatFile(m, from)
	if m.WarnNotFound(e, from) {
		return
	}
	defer Remove(m, from)
	p := path.Join(ice.VAR_TRASH, path.Base(from))
	kit.If(!s.IsDir(), func() { Open(m, from, func(r io.Reader) { p = path.Join(ice.VAR_TRASH, kit.HashsPath(r)) }) })
	RemoveAll(m, p)
	kit.If(!m.WarnNotValid(Rename(m, from, p)), func() { mdb.HashCreate(m, FROM, kit.Paths(from), FILE, p) })
}

const TRASH = "trash"

func init() {
	Index.MergeCommands(ice.Commands{
		TRASH: {Name: "trash hash auto", Help: "回收站", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				_trash_create(m, kit.Paths(m.Option(FROM)))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				Remove(m, m.Option(FILE))
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.REVERT: {Help: "恢复", Icon: "bi bi-folder-symlink", Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.HashSelect(m.Spawn(), m.Option(mdb.HASH))
				Rename(m, msg.Append(FILE), msg.Append(FROM))
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.PRUNES: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunes(m, nil).Table(func(value ice.Maps) { Remove(m, value[FILE]) })
			}},
		}, mdb.HashAction(mdb.SHORT, FROM, mdb.FIELD, "time,hash,from,file", mdb.ACTION, mdb.REVERT))},
	})
}

func Trash(m *ice.Message, p string, arg ...string) *ice.Message { return m.Cmd(TRASH, mdb.CREATE, p) }
