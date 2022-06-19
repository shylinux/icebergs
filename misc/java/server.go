package java

import (
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	JAVA  = "java"
	JAVAC = "javac"
)

type server struct {
	ice.Code
	regexp string `data:".*.java"`
	linux  string `data:"https://mirrors.tencent.com/repository/generic/konajdk/8/0/10/linux-x86_64/b1/TencentKona8.0.10.b1_jdk_linux-x86_64_8u332.tar.gz"`
	list   string `name:"list path auto listScript order install" help:"编译执行"`
}

func (s server) Order(m *ice.Message) {
	s.Code.Order(m, "", ice.BIN)
}
func (s server) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}
func (s server) RunScript(m *ice.Message) {
	if s.Code.System(m, nfs.PWD, JAVAC, "-d", ice.BIN, m.Option(nfs.PATH)); cli.IsSuccess(m.Message) {
		s.Code.System(m, nfs.PWD, JAVA, "-cp", kit.Path(ice.BIN), strings.TrimPrefix(strings.TrimSuffix(m.Option(nfs.PATH), ".java"), "src/"))
	}
}

func init() { ice.CodeCtxCmd(server{}) }
