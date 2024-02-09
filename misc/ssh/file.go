package ssh

import (
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/toolkits/file"
)

type FileSystem struct{ *sftp.Client }

func (s *FileSystem) StatFile(p string) (os.FileInfo, error) {
	return s.Client.Stat(p)
}
func (s *FileSystem) OpenFile(p string) (io.ReadCloser, error) {
	return s.Client.Open(p)
}
func (s *FileSystem) CreateFile(p string) (io.WriteCloser, string, error) {
	if f, p, e := file.CreateFiles(s, p); f != nil {
		return f, p, e
	}
	f, e := s.Client.Create(p)
	return f, p, e
}
func (s *FileSystem) AppendFile(p string) (io.ReadWriteCloser, error) {
	if f, _, e := file.CreateFiles(s, p); f != nil {
		return f, e
	}
	return s.Client.OpenFile(p, os.O_RDWR|os.O_APPEND|os.O_CREATE)
}
func (s *FileSystem) WriteFile(p string, b []byte) error {
	f, p, e := s.CreateFile(p)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(b)
	return e
}

func (s *FileSystem) ReadDir(p string) ([]os.FileInfo, error) {
	return s.Client.ReadDir(p)
}
func (s *FileSystem) MkdirAll(p string, m os.FileMode) error {
	return s.Client.MkdirAll(p)
}
func (s *FileSystem) RemoveAll(p string) error {
	return s.Client.RemoveAll(p)
}
func (s *FileSystem) Remove(p string) error {
	return s.Client.Remove(p)
}
func (s *FileSystem) Rename(oldname string, newname string) error {
	return s.Client.Rename(oldname, newname)
}
func (s *FileSystem) Symlink(oldname string, newname string) error {
	return s.Client.Symlink(oldname, newname)
}
func (s *FileSystem) Link(oldname string, newname string) error {
	return s.Client.Link(oldname, newname)
}
func (s *FileSystem) Close() error { return nil }

func Open(m *ice.Message, cb func(*FileSystem)) {
	_ssh_conn(m, func(c *ssh.Client) {
		defer c.Close()
		if s, e := sftp.NewClient(c); !m.WarnNotValid(e) {
			cb(&FileSystem{s})
		}
	})
}
