package java

import (
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type spring struct {
	ice.Code
	linux string `data:"https://mirrors.tencent.com/macports/distfiles/spring-boot-cli/spring-boot-cli-2.7.0-bin.tar.gz"`
	ice.Hash
	short  string `data:"name"`
	field  string `data:"time,name,path"`
	create string `name:"create name path"`
	start  string `name:"start server.port"`
	list   string `name:"list name auto create order install" help:"服务框架"`
}

func (s spring) Build(m *ice.Message) {
	s.Code.Stream(m, m.Option(nfs.PATH), MVN, "package")
}
func (s spring) Start(m *ice.Message, arg ...string) {
	s.Code.Daemon(m, m.Option(nfs.PATH), kit.Simple(JAVA, kit.Simple(arg, func(k, v string) string { return "-D" + k + mdb.EQ + v }),
		"-jar", kit.Format("target/%s-0.0.1-SNAPSHOT.jar", m.Option(mdb.NAME)))...)
}
func (s spring) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		s.Hash.List(m, arg...).PushAction(s.Start, s.Build)
	} else {
		m.Cmd(cli.DAEMON).Table(func(value ice.Maps, index int, head []string) {
			if strings.Contains(value[ice.CMD], "target/"+arg[0]+"-0.0.1-SNAPSHOT.jar") {
				m.PushRecord(value, head...)
			}
		})
	}
}
func init() { ice.CodeCtxCmd(spring{}) }
