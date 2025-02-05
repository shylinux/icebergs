package publish

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
)

type server struct {
	ice.Hash
	short string `data:"name"`
	field string `data:"time,name*,text,path,version*,compile,runtime,os,cpu"`
	list  string `name:"list name auto" help:"软件源" role:"void"`
}

func (s server) Inputs(m *ice.Message, arg ...string) {
	switch arg[0] {
	case nfs.PATH:
		m.Cmdy(nfs.DIR, nfs.USR_PUBLISH, nfs.PATH)
	case code.COMPILE:
		m.Push(arg[0], "go")
		m.Push(arg[0], "javac")
	case cli.RUNTIME:
		m.Push(arg[0], "python")
		m.Push(arg[0], "java")
		m.Push(arg[0], "php")
	case cli.OS:
		m.Push(arg[0], "Linux")
		m.Push(arg[0], "macOS")
		m.Push(arg[0], "Windows")
	case cli.CPU:
		m.Push(arg[0], "amd64")
		m.Push(arg[0], "x86")
		m.Push(arg[0], "arm")
		m.Push(arg[0], "arm64")
	default:
		s.Hash.Inputs(m, arg...)
	}
}
func (s server) Upload(m *ice.Message, arg ...string) {
	s.Modify(m, mdb.NAME, m.Option(mdb.NAME), nfs.PATH, m.UploadSave(nfs.USR_PUBLISH))
}
func (s server) List(m *ice.Message, arg ...string) {
	if s.Hash.List(m, arg...); m.IsTech() {
		m.PushAction(s.Detail, s.Upload, s.Remove)
	}
}

func init() { ice.Cmd("web.code.publish.server", server{}) }
