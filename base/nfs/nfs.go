package nfs

import (
	"io"
	"io/ioutil"
	"os"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/toolkits/file"
)

var Index = &ice.Context{Name: "nfs", Help: "存储模块"}

func init() {
	ice.Index.Register(Index, nil, TAR, CAT, DIR, DEFS, SAVE, PUSH, COPY, LINK, TAIL, TRASH, GREP)
}

var DiskFile = file.NewDiskFile()
var PackFile = file.NewPackFile()

func init() { file.Init(ice.Pulse.OptionFiles(DiskFile, PackFile)) }

func StatFile(m *ice.Message, p string) (os.FileInfo, error) {
	file := m.OptionFiles()
	return file.StatFile(p)
}
func OpenFile(m *ice.Message, p string) (io.ReadCloser, error) {
	file := m.OptionFiles()
	return file.OpenFile(p)
}
func CreateFile(m *ice.Message, p string) (io.WriteCloser, string, error) {
	file := m.OptionFiles()
	return file.CreateFile(p)
}
func AppendFile(m *ice.Message, p string) (io.ReadWriteCloser, string, error) {
	file := m.OptionFiles()
	w, e := file.AppendFile(p)
	return w, p, e
}
func WriteFile(m *ice.Message, p string, b []byte) error {
	file := m.OptionFiles()
	return file.WriteFile(p, b)
}

func ReadDir(m *ice.Message, p string) ([]os.FileInfo, error) {
	file := m.OptionFiles()
	list, e := file.ReadDir(p)
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
	file := m.OptionFiles()
	return file.MkdirAll(p, ice.MOD_DIR)
}
func RemoveAll(m *ice.Message, p string) error {
	file := m.OptionFiles()
	return file.RemoveAll(p)
}
func Remove(m *ice.Message, p string) error {
	file := m.OptionFiles()
	return file.Remove(p)
}
func Rename(m *ice.Message, oldname string, newname string) error {
	file := m.OptionFiles()
	return file.Rename(oldname, newname)
}
func Symlink(m *ice.Message, oldname string, newname string) error {
	file := m.OptionFiles()
	return file.Symlink(oldname, newname)
}
func Link(m *ice.Message, oldname string, newname string) error {
	file := m.OptionFiles()
	return file.Link(oldname, newname)
}
func Close(m *ice.Message) {
	file := m.OptionFiles()
	file.Close()
}

func ExistsFile(m *ice.Message, p string) bool {
	file := m.OptionFiles()
	if _, e := file.StatFile(p); e == nil {
		return true
	}
	return false
}
func ReadFile(m *ice.Message, p string) ([]byte, error) {
	file := m.OptionFiles()
	if f, e := file.OpenFile(p); e == nil {
		return ioutil.ReadAll(f)
	} else {
		return nil, e
	}
}

func NewWriteCloser(w func([]byte) (int, error), c func() error) io.WriteCloser {
	return file.NewWriteCloser(w, c)
}
func NewReadCloser(r io.Reader) io.ReadCloser {
	return file.NewReadCloser(r)
}
