package java

import "shylinux.com/x/ice"

type maven struct {
	ice.Code
	linux string `data:"https://mirrors.tencent.com/macports/distfiles/maven3/apache-maven-3.8.5-bin.tar.gz"`
	list  string `name:"list path auto order install" help:"打包构建"`
}

func (s maven) List(m *ice.Message, arg ...string) {
	s.Code.Source(m, "", arg...)
}
func init() { ice.CodeCtxCmd(maven{}) }
