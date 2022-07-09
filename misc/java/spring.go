package java

import (
	"path"
	"strings"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

type spring struct {
	ice.Code
	linux string `data:"https://mirrors.tencent.com/macports/distfiles/spring-boot-cli/spring-boot-cli-2.7.0-bin.tar.gz"`

	ice.Hash
	short string `data:"name"`
	field string `data:"time,name,text"`

	create string `name:"create name text" help:"创建"`
	start  string `name:"start server.port" help:"启动"`

	list string `name:"list name auto create order install" help:"服务框架"`
}

func (s spring) Order(m *ice.Message) {
	s.Code.Order(m, "", ice.BIN)
}
func (s spring) Create(m *ice.Message, arg ...string) {
	s.Hash.Create(m, arg...)
}
func (s spring) Build(m *ice.Message) {
	s.Code.Stream(m, path.Join(ice.USR, m.Option(mdb.NAME)), "mvn", "package")
}
func (s spring) Start(m *ice.Message, arg ...string) {
	args := []string{}
	for i := 0; i < len(arg)-1; i += 2 {
		args = append(args, "-D"+arg[i]+"="+arg[i+1])
	}
	s.Code.Daemon(m, path.Join(ice.USR, m.Option(mdb.NAME)), kit.Simple("java", args, "-jar", kit.Format("target/%s-0.0.1-SNAPSHOT.jar", m.Option(mdb.NAME)))...)
}
func (s spring) List(m *ice.Message, arg ...string) {
	if len(arg) == 0 { // 项目列表
		s.Hash.List(m, arg...).PushAction(s.Start, s.Build)

	} else { // 服务列表
		m.Cmd(cli.DAEMON).Table(func(index int, value ice.Maps, head []string) {
			if strings.Contains(value[ice.CMD], "target/"+arg[0]+"-0.0.1-SNAPSHOT.jar") {
				m.PushRecord(value, head...)
			}
		})
	}
}

func init() { ice.CodeCtxCmd(spring{}) }
