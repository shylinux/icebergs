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

func _tar_list(m *ice.Message, p string, cb func(*tar.Header, *tar.Reader, int)) *ice.Message {
	const (
		GZ = "gz"
	)
	var r io.Reader
	if f, e := os.Open(p); m.Warn(e, ice.ErrNotValid, p) {
		return m
	} else {
		defer f.Close()
		r = f
	}
	for {
		switch kit.Ext(p) {
		case GZ:
			if f, e := gzip.NewReader(r); m.Warn(e, ice.ErrNotValid, p) {
				return m
			} else {
				defer f.Close()
				r = f
			}
			p = kit.TrimExt(p, GZ)
		case TAR:
			i := 0
			for r := tar.NewReader(r); ; i++ {
				h, e := r.Next()
				if m.Warn(e) {
					break
				}
				cb(h, r, i)
			}
			m.StatusTimeCount(mdb.TOTAL, i)
			m.Debug("what %v", i)
			return m
		default:
			return m
		}
	}
	return m
}

const TAR = "tar"

func init() {
	Index.MergeCommands(ice.Commands{
		TAR: {Name: "tar path file auto page", Help: "打包", Actions: ice.MergeActions(ice.Actions{
			mdb.NEXT: {Hand: func(m *ice.Message, arg ...string) { mdb.PrevPage(m, arg[0], kit.Slice(arg, 1)...) }},
			mdb.PREV: {Hand: func(m *ice.Message, arg ...string) { mdb.NextPageLimit(m, arg[0], kit.Slice(arg, 1)...) }},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) {
				list := map[string]bool{}
				_tar_list(m, m.Option(PATH), func(h *tar.Header, r *tar.Reader, i int) {
					if h.Name == m.Option(FILE) || m.Option(FILE) == "" {
						p := path.Join(path.Dir(m.Option(PATH)), h.Name)
						if strings.HasSuffix(h.Name, ice.PS) {
							MkdirAll(m, p)
							return
						}
						if !list[path.Dir(p)] {
							list[path.Dir(p)] = true
							MkdirAll(m, path.Dir(p))
						}
						if f, p, e := CreateFile(m, p); !m.Warn(e) {
							defer f.Close()
							if m.Option(FILE) != "" {
								defer m.Cmdy(DIR, p, "time,path,size")
								defer m.Cmdy(CAT, p)
							}
							n, e := io.Copy(f, r)
							m.Logs(mdb.EXPORT, FILE, p, SIZE, n).Warn(e)
							os.Chmod(p, os.FileMode(h.Mode))
						}
					}
				})
			}},
		}, mdb.PageListAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || strings.HasSuffix(arg[0], ice.PS) {
				m.Cmdy(DIR, arg)
				return
			}
			page, size := mdb.OptionPages(m, kit.Slice(arg, 2)...)
			_tar_list(m, arg[0], func(h *tar.Header, r *tar.Reader, i int) {
				if len(kit.Slice(arg, 0, 2)) > 1 {
					if h.Name != arg[1] {
						return
					}
					m.Echo(string(ReadAll(m, r)[:]))
				}
				if i >= (page-1)*size && i < page*size {
					m.Push(mdb.TIME, h.ModTime.Format(ice.MOD_TIME)).Push(FILE, h.Name).Push(SIZE, kit.FmtSize(h.Size))
				}
			}).PushAction(mdb.EXPORT)
		}},
	})
}
func ReadAll(m *ice.Message, r io.Reader) []byte {
	if buf, e := io.ReadAll(r); m.Warn(e) {
		return buf
	} else {
		return buf
	}
}
