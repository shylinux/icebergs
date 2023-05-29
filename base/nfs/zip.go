package nfs

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _zip_list(m *ice.Message, p string, cb func(zip.FileHeader, io.Reader, int)) {
	if f, e := zip.OpenReader(p); m.Warn(e, ice.ErrNotValid, p) {
		return
	} else {
		defer f.Close()
		for i, f := range f.File {
			if r, e := f.Open(); e == nil {
				defer r.Close()
				cb(f.FileHeader, r, i)
			}
		}
	}
}

const ZIP = "zip"

func init() {
	Index.MergeCommands(ice.Commands{
		ZIP: {Name: "zip path file auto page", Help: "打包", Actions: ice.MergeActions(ice.Actions{
			mdb.NEXT: {Hand: func(m *ice.Message, arg ...string) { mdb.PrevPage(m, arg[0], kit.Slice(arg, 1)...) }},
			mdb.PREV: {Hand: func(m *ice.Message, arg ...string) { mdb.NextPageLimit(m, arg[0], kit.Slice(arg, 1)...) }},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) {
				list, size := kit.Dict(), 0
				_zip_list(m, m.Option(PATH), func(h zip.FileHeader, r io.Reader, i int) {
					p := path.Join(path.Dir(m.Option(PATH)), kit.Split(path.Base(m.Option(PATH)), "_-.")[0], h.Name)
					if strings.HasSuffix(h.Name, PS) {
						MkdirAll(m, p)
						return
					}
					kit.IfNoKey(list, path.Dir(p), func(p string) { MkdirAll(m, p) })
					Create(m, p, func(w io.Writer) {
						os.Chmod(p, os.FileMode(h.Mode()))
						Copy(m, w, r, func(n int) { size += n })
						kit.If(m.Option(FILE), func() { m.Cmdy(DIR, p).Cmdy(CAT, p) })
					})
				})
			}},
		}, mdb.PageListAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], PS) {
				m.Cmdy(DIR, arg)
				return
			}
			page, size := mdb.OptionPages(m, kit.Slice(arg, 2)...)
			_zip_list(m, arg[0], func(h zip.FileHeader, r io.Reader, i int) {
				if len(kit.Slice(arg, 0, 2)) > 1 {
					if h.Name != arg[1] {
						return
					}
					m.Echo(string(ReadAll(m, r)[:]))
				}
				if i >= (page-1)*size && i < page*size {
					m.Push(mdb.TIME, h.ModTime().Format(ice.MOD_TIME)).Push(FILE, h.Name).Push(SIZE, kit.FmtSize(int64(h.UncompressedSize)))
				}
			})
			m.PushAction(mdb.EXPORT)
		}},
	})
}
