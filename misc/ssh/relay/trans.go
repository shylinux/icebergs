package relay

import (
	"io"
	"os"
	"path"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/misc/ssh"
	kit "shylinux.com/x/toolkits"
)

const (
	TRANS     = "trans"
	SSH_TRANS = "ssh.trans"
)

type trans struct {
	relay
	send string `name:"send from path file"`
	list string `name:"list machine path auto" help:"远程文件"`
}

func (s trans) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case nfs.FROM:
		m.Cmdy(nfs.DIR, ice.USR_PUBLISH, nfs.PATH, nfs.SIZE, mdb.TIME)
	}
}
func (s trans) Send(m *ice.Message, arg ...string) {
	defer m.ToastProcess()()
	nfs.Open(m.Message, m.Option(nfs.FROM), func(r io.Reader, info os.FileInfo) {
		s.open(m, func(fs *ssh.FileSystem) {
			nfs.Create(m.Message, path.Join(m.Option(nfs.PATH), m.OptionDefault(nfs.FILE, path.Base(m.Option(nfs.FROM)))), func(w io.Writer, p string) {
				m.Logs(tcp.SEND, nfs.TO, p, m.OptionSimple(tcp.HOST, nfs.FROM), nfs.SIZE, kit.FmtSize(info.Size()))
				m.GoToast(func(toast func(name string, count, total int)) {
					last := 0
					nfs.CopyStream(m.Message, w, r, 81920, int(info.Size()), func(count, total int) {
						if size := count / 1024 / 1024; size != last {
							toast(p, count, total)
							last = size
						}
					})
				})
			})
		})
	})
}
func (s trans) Upload(m *ice.Message, arg ...string) {
	ls := web.Upload(m.Message)
	s.Send(m.Options(nfs.FROM, m.Cmd(web.CACHE, ls[0]).Append(nfs.FILE), nfs.FILE, ls[1]))
}
func (s trans) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		m.Cmdy(s.relay).Cut("time,machine,username,host,port").PushAction().Action()
	} else {
		s.open(m, func(fs *ssh.FileSystem) {
			m.Cmdy(nfs.CAT, arg[1:]).PushAction(s.Trash)
		}, arg...).Action(s.Send, s.Upload)
	}
}
func (s trans) Trash(m *ice.Message, arg ...string) {
	defer m.ToastProcess()()
	s.open(m, func(fs *ssh.FileSystem) { fs.RemoveAll(m.Option(nfs.PATH)) })
}

func init() { ice.Cmd(SSH_TRANS, trans{}) }

func (s trans) open(m *ice.Message, cb func(*ssh.FileSystem), arg ...string) *ice.Message {
	ssh.Open(m.Options(m.Cmd(s.relay, kit.Select(m.Option(MACHINE), arg, 0)).AppendSimple()).Message, func(fs *ssh.FileSystem) {
		defer m.Options(ice.MSG_FILES, fs).Options(ice.MSG_FILES, nil)
		cb(fs)
	})
	return m
}
