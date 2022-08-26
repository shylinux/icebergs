package nfs

import (
	"io"
	"io/ioutil"
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
	pack := PackFile
	Index.MergeCommands(ice.Commands{
		PACK: {Name: "pack path auto upload create", Help: "文件系统", Actions: ice.Actions{
			mdb.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if b, h, e := m.R.FormFile(mdb.UPLOAD); m.Assert(e) {
					defer b.Close()
					if f, p, e := pack.CreateFile(path.Join(m.Option(PATH), h.Filename)); m.Assert(e) {
						defer f.Close()
						if n, e := io.Copy(f, b); m.Assert(e) {
							m.Logs(mdb.EXPORT, FILE, p, SIZE, n)
						}
					}
				}
			}},
			mdb.CREATE: {Name: "create path=h1/h2/hi.txt text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if f, p, e := pack.CreateFile(m.Option(PATH)); m.Assert(e) {
					defer f.Close()
					if n, e := f.Write([]byte(m.Option(mdb.TEXT))); m.Assert(e) {
						m.Logs(mdb.EXPORT, FILE, p, SIZE, n)
					}
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				pack.Remove(path.Clean(m.Option(PATH)))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			p := kit.Select("", arg, 0)
			if p != "" && !strings.HasSuffix(p, PS) {
				if f, e := pack.OpenFile(p); e == nil {
					defer f.Close()
					if b, e := ioutil.ReadAll(f); e == nil {
						m.Echo(string(b))
					}
				}
				return
			}

			ls, _ := pack.ReadDir(p)
			for _, f := range ls {
				m.Push(mdb.TIME, f.ModTime().Format(ice.MOD_TIME))
				m.Push(PATH, path.Join(p, f.Name())+kit.Select("", PS, f.IsDir()))
				m.Push(SIZE, kit.FmtSize(f.Size()))
			}
			m.Sort("time,path")
			m.PushAction(mdb.REMOVE)
			m.StatusTimeCount()
		}},
	})
}

var PackFile = file.NewPackFile()
var DiskFile = file.NewDiskFile()

func init() { file.Init(OptionFiles(ice.Pulse, DiskFile, PackFile)) }

func OptionFiles(m *ice.Message, f ...file.File) file.File {
	if len(f) > 1 {
		m.Optionv(ice.MSG_FILES, file.NewMultiFile(f...))
	} else if len(f) > 0 {
		m.Optionv(ice.MSG_FILES, f[0])
	}
	return m.Optionv(ice.MSG_FILES).(file.File)
}
func StatFile(m *ice.Message, p string) (os.FileInfo, error) {
	return OptionFiles(m).StatFile(p)
}
func OpenFile(m *ice.Message, p string) (io.ReadCloser, error) {
	return OptionFiles(m).OpenFile(p)
}
func CreateFile(m *ice.Message, p string) (io.WriteCloser, string, error) {
	return OptionFiles(m).CreateFile(p)
}
func AppendFile(m *ice.Message, p string) (io.ReadWriteCloser, string, error) {
	file := OptionFiles(m)
	w, e := file.AppendFile(p)
	return w, p, e
}
func WriteFile(m *ice.Message, p string, b []byte) error {
	return OptionFiles(m).WriteFile(p, b)
}

func ReadDir(m *ice.Message, p string) ([]os.FileInfo, error) {
	list, e := OptionFiles(m).ReadDir(p)
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].Name() > list[j].Name() {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	return list, e
}
func MkdirAll(m *ice.Message, p string) error {
	return OptionFiles(m).MkdirAll(p, ice.MOD_DIR)
}
func RemoveAll(m *ice.Message, p string) error {
	return OptionFiles(m).RemoveAll(p)
}
func Remove(m *ice.Message, p string) error {
	return OptionFiles(m).Remove(p)
}
func Rename(m *ice.Message, oldname string, newname string) error {
	return OptionFiles(m).Rename(oldname, newname)
}
func Symlink(m *ice.Message, oldname string, newname string) error {
	return OptionFiles(m).Symlink(oldname, newname)
}
func Link(m *ice.Message, oldname string, newname string) error {
	return OptionFiles(m).Link(oldname, newname)
}

func ExistsFile(m *ice.Message, p string) bool {
	if _, e := OptionFiles(m).StatFile(p); e == nil {
		return true
	}
	return false
}
func ReadFile(m *ice.Message, p string) ([]byte, error) {
	if f, e := OptionFiles(m).OpenFile(p); e == nil {
		defer f.Close()
		return ioutil.ReadAll(f)
	} else {
		return nil, e
	}
}
func CloseFile(m *ice.Message, p ice.Any) {
	if w, ok := p.(io.Closer); ok {
		w.Close()
	}
}

func CopyFile(m *ice.Message, to io.WriteCloser, from io.ReadCloser, cb func(int)) {
	buf := make([]byte, 1024*1024)
	total := 0
	for {
		n, e := from.Read(buf)
		to.Write(buf[:n])
		cb(n)
		total += n
		m.Debug("what %v %v %v", n, total, e)
		if e != nil {
			break
		}
	}
}

func NewWriteCloser(w func([]byte) (int, error), c func() error) io.WriteCloser {
	return file.NewWriteCloser(w, c)
}
func NewReadCloser(r io.Reader) io.ReadCloser {
	return file.NewReadCloser(r)
}
func NewCloser(c func() error) io.WriteCloser {
	return file.NewWriteCloser(func(buf []byte) (int, error) { return 0, nil }, c)
}
