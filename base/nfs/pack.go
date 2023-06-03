package nfs

import (
	"io"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

const PACK = "pack"

func init() {
	Index.MergeCommands(ice.Commands{
		PACK: {Name: "pack path auto upload create", Help: "文件系统", Actions: ice.Actions{
			mdb.CREATE: {Name: "create path*=src/hi/hi.txt text*=hello", Hand: func(m *ice.Message, arg ...string) {
				OptionFiles(m, PackFile)
				Create(m, m.Option(PATH), func(w io.Writer, p string) {
					Save(m, w, m.Option(mdb.TEXT), func(n int) { m.Logs(LOAD, FILE, p, SIZE, n) })
				})
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { PackFile.Remove(path.Clean(m.Option(PATH))) }},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH && arg[1] != "" {
					m.Cmd(DIR, SRC, PATH, kit.Dict(DIR_REG, arg[1], DIR_DEEP, ice.TRUE, DIR_TYPE, CAT), func(value ice.Maps) {
						if strings.HasPrefix(value[PATH], ice.SRC_TEMPLATE) {
							return
						}
						m.PushSearch(mdb.TYPE, kit.Ext(value[PATH]), mdb.NAME, path.Base(value[PATH]), mdb.TEXT, value[PATH])
					})
					OptionFiles(m, PackFile)
					m.Cmd(DIR, USR, PATH, kit.Dict(DIR_REG, arg[1], DIR_DEEP, ice.TRUE, DIR_TYPE, CAT), func(value ice.Maps) {
						m.PushSearch(mdb.TYPE, kit.Ext(value[PATH]), mdb.NAME, path.Base(value[PATH]), mdb.TEXT, value[PATH])
					})
				}
			}},
			mdb.IMPORT: {Hand: func(m *ice.Message, arg ...string) {
				OptionFiles(m, DiskFile)
				Open(m, path.Join(m.Option(PATH), m.Option(FILE)), func(r io.Reader, p string) {
					OptionFiles(m, PackFile)
					Create(m, p, func(w io.Writer) { Copy(m, w, r, func(n int) { m.Logs(LOAD, FILE, p, SIZE, n) }) })
				})
			}},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) {
				OptionFiles(m, PackFile)
				Open(m, path.Join(m.Option(PATH), m.Option(FILE)), func(r io.Reader, p string) {
					OptionFiles(m, DiskFile)
					Create(m, p, func(w io.Writer) { Copy(m, w, r, func(n int) { m.Logs(LOAD, FILE, p, SIZE, n) }) })
				})
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			OptionFiles(m, PackFile)
			if p := kit.Select("", arg, 0); p != "" && !strings.HasSuffix(p, PS) {
				Open(m, p, func(r io.Reader) { m.Echo(string(ReadAll(m, r))) })
			} else {
				Open(m, path.Join(p)+PS, func(s os.FileInfo) {
					m.Push(mdb.TIME, s.ModTime().Format(ice.MOD_TIME))
					m.Push(PATH, path.Join(p, s.Name())+kit.Select("", PS, s.IsDir()))
					m.Push(SIZE, kit.FmtSize(s.Size()))
				})
				m.PushAction(mdb.REMOVE).StatusTimeCount()
			}
		}},
	})
}

var PackFile = file.NewPackFile()
var DiskFile = file.NewDiskFile()

func init() { file.Init(OptionFiles(ice.Pulse, DiskFile, PackFile)) }
func init() { ice.Info.OpenFile = OpenFile }

type optionMessage interface {
	Optionv(key string, arg ...ice.Any) ice.Any
}

func OptionFiles(m optionMessage, f ...file.File) file.File {
	if len(f) > 1 {
		m.Optionv(ice.MSG_FILES, file.NewMultiFile(f...))
	} else if len(f) > 0 {
		m.Optionv(ice.MSG_FILES, f[0])
	}
	return m.Optionv(ice.MSG_FILES).(file.File)
}
func StatFile(m optionMessage, p string) (os.FileInfo, error)  { return OptionFiles(m).StatFile(p) }
func OpenFile(m *ice.Message, p string) (io.ReadCloser, error) { return OptionFiles(m).OpenFile(p) }
func CreateFile(m optionMessage, p string) (io.WriteCloser, string, error) {
	return OptionFiles(m).CreateFile(p)
}
func AppendFile(m optionMessage, p string) (io.ReadWriteCloser, string, error) {
	w, e := OptionFiles(m).AppendFile(p)
	return w, p, e
}
func WriteFile(m optionMessage, p string, b []byte) error { return OptionFiles(m).WriteFile(p, b) }

func ReadDir(m optionMessage, p string) ([]os.FileInfo, error) {
	list, e := OptionFiles(m).ReadDir(p)
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].IsDir() && !list[j].IsDir() {
				continue
			} else if !list[i].IsDir() && list[j].IsDir() || list[i].Name() > list[j].Name() {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	return list, e
}
func MkdirAll(m optionMessage, p string) string {
	OptionFiles(m).MkdirAll(p, ice.MOD_DIR)
	return p
}
func RemoveAll(m optionMessage, p string) error { return OptionFiles(m).RemoveAll(p) }
func Remove(m optionMessage, p string) error    { return OptionFiles(m).Remove(p) }
func Rename(m optionMessage, oldname string, newname string) error {
	MkdirAll(m, path.Dir(newname))
	return OptionFiles(m).Rename(oldname, newname)
}
func Symlink(m optionMessage, oldname string, newname string) error {
	return OptionFiles(m).Symlink(oldname, newname)
}
func Link(m optionMessage, oldname string, newname string) error {
	return OptionFiles(m).Link(oldname, newname)
}

func Exists(m optionMessage, p string) bool {
	if _, e := OptionFiles(m).StatFile(p); e == nil {
		return true
	}
	return false
}
func NewReadCloser(r io.Reader) io.ReadCloser { return file.NewReadCloser(r) }
func NewWriteCloser(w func([]byte) (int, error), c func() error) io.WriteCloser {
	return file.NewWriteCloser(w, c)
}
func Close(m optionMessage, p ice.Any) {
	if w, ok := p.(io.Closer); ok {
		w.Close()
	}
}
