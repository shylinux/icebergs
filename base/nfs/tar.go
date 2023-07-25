package nfs

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _tar_list(m *ice.Message, p string, cb func(*tar.Header, io.Reader, int)) {
	Open(m, p, func(r io.Reader) {
		for {
			switch kit.Ext(p) {
			case TGZ:
				p = kit.Keys(kit.TrimExt(p, kit.Ext(p)), TAR, GZ)
			case GZ:
				if f, e := gzip.NewReader(r); m.Warn(e, ice.ErrNotValid, p) {
					return
				} else {
					defer f.Close()
					r, p = f, kit.TrimExt(p, GZ)
				}
			case TAR:
				i := 0
				for r := tar.NewReader(r); ; i++ {
					h, e := r.Next()
					if m.Warn(e) || e == io.EOF {
						break
					}
					if h.Size == 0 {
						i--
						continue
					}
					cb(h, r, i)
				}
				m.StatusTimeCount(mdb.TOTAL, i)
				return
			default:
				return
			}
		}
	})
}

const (
	XZ  = "xz"
	GZ  = "gz"
	TGZ = "tgz"
)
const TAR = "tar"

func init() {
	Index.MergeCommands(ice.Commands{
		TAR: {Name: "tar path file auto page", Help: "打包", Actions: ice.MergeActions(ice.Actions{
			mdb.NEXT: {Hand: func(m *ice.Message, arg ...string) { mdb.PrevPage(m, arg[0], kit.Slice(arg, 1)...) }},
			mdb.PREV: {Hand: func(m *ice.Message, arg ...string) { mdb.NextPageLimit(m, arg[0], kit.Slice(arg, 1)...) }},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) {
				if kit.Ext(m.Option(PATH)) == ZIP {
					m.Cmdy(ZIP, mdb.EXPORT, arg)
					return
				}
				list, size := kit.Dict(), 0
				_tar_list(m, m.Option(PATH), func(h *tar.Header, r io.Reader, i int) {
					if h.Name == m.Option(FILE) || m.Option(FILE) == "" {
						p := path.Join(path.Dir(m.Option(PATH)), h.Name)
						if strings.HasSuffix(h.Name, PS) {
							MkdirAll(m, p)
							return
						}
						kit.IfNoKey(list, path.Dir(p), func(p string) { MkdirAll(m, p) })
						Create(m, p, func(w io.Writer) {
							os.Chmod(p, os.FileMode(h.Mode))
							Copy(m, w, r, func(n int) { size += n })
							kit.If(m.Option(FILE), func() { m.Cmdy(DIR, p).Cmdy(CAT, p) })
						})
					}
				})
			}},
		}, mdb.PageListAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], PS) {
				m.Cmdy(DIR, arg)
				return
			}
			if kit.Ext(arg[0]) == ZIP {
				m.Cmdy(ZIP, arg)
				return
			}
			page, size := mdb.OptionPages(m, kit.Slice(arg, 2)...)
			_tar_list(m, arg[0], func(h *tar.Header, r io.Reader, i int) {
				if len(kit.Slice(arg, 0, 2)) > 1 {
					if h.Name != arg[1] {
						return
					}
					m.Echo(string(ReadAll(m, r)[:]))
				}
				if i >= (page-1)*size && i < page*size {
					m.Push(mdb.TIME, h.ModTime.Format(ice.MOD_TIME)).Push(FILE, h.Name).Push(SIZE, kit.FmtSize(h.Size))
				}
			})
			m.PushAction(mdb.EXPORT)
		}},
	})
}
func TarExport(m *ice.Message, path string, file ...string) {
	m.Cmd(TAR, mdb.EXPORT, ice.Maps{PATH: path, FILE: kit.Select("", file, 0)})
}
