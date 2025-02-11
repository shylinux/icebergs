package java

import (
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	MVN   = "mvn"
	JAVA  = "java"
	JAVAC = "javac"
)

type java struct {
	ice.Code
	ice.Lang
	regexp string `data:"java"`
	linux  string `data:"https://mirrors.huaweicloud.com/openjdk/18/openjdk-18_linux-x64_bin.tar.gz"`
	list   string `name:"list path auto listScript order install" help:"编译器"`
}

func (s java) Init(m *ice.Message) {
	cli.IsAlpine(m.Message, JAVA, "openjdk8")
	cli.IsAlpine(m.Message, JAVAC, "openjdk8")
	cli.IsAlpine(m.Message, MVN, "maven openjdk8")
	cli.IsRedhat(m.Message, JAVA, "java-1.8.0-openjdk-devel.x86_64")
	cli.IsRedhat(m.Message, JAVAC, "java-1.8.0-openjdk-devel.x86_64")
	cli.IsRedhat(m.Message, MVN, "maven java-1.8.0-openjdk-devel.x86_64")
	s.Lang.Init(m, nfs.SCRIPT, m.Resource(""))
}
func (s java) Render(m *ice.Message, arg ...string) {
}
func (s java) Engine(m *ice.Message, arg ...string) {
}
func (s java) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}
func (s java) RunScript(m *ice.Message) {
	if s.Code.System(m, nfs.PWD, JAVAC, "-d", ice.BIN, m.Option(nfs.PATH)); cli.IsSuccess(m.Message) {
		s.Code.System(m, nfs.PWD, JAVA, "-cp", kit.Path(ice.BIN), strings.TrimPrefix(strings.TrimSuffix(m.Option(nfs.PATH), ".java"), "src/"))
	}
}
func init() { ice.CodeCtxCmd(java{}) }
