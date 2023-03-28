package nfs

import (
	"io"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _trash_create(m *ice.Message, from string) {
	if m.Warn(from == "", ice.ErrNotValid, FROM) {
		return
	}
	s, e := StatFile(m, from)
	defer os.Remove(from)
	if m.Warn(e, ice.ErrNotFound, from) {
		return
	}
	p := path.Join(ice.VAR_TRASH, path.Base(from))
	kit.If(!s.IsDir(), func() { Open(m, from, func(r io.Reader) { p = path.Join(ice.VAR_TRASH, kit.HashsPath(r)) }) })
	RemoveAll(m, p)
	kit.If(!m.Warn(Rename(m, from, p)), func() { mdb.HashCreate(m, FROM, from, FILE, p) })
}

const TRASH = "trash"

func init() {
	Index.MergeCommands(ice.Commands{
		TRASH: {Name: "trash hash auto prunes", Help: "回收站", Actions: ice.MergeActions(ice.Actions{
			mdb.REVERT: {Hand: func(m *ice.Message, arg ...string) {
				Rename(m, m.Option(FILE), m.Option(FROM))
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) {
				_trash_create(m, m.Option(FROM))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				Remove(m, m.Option(FILE))
				mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
			}},
			mdb.PRUNES: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunes(m, nil).Table(func(value ice.Maps) { Remove(m, value[FILE]) })
			}},
		}, mdb.HashAction(mdb.SHORT, FROM, mdb.FIELD, "time,hash,from,file", mdb.ACTION, mdb.REVERT))},
	})
}

func Trash(m *ice.Message, p string) *ice.Message { return m.Cmd(TRASH, mdb.CREATE, p) }
