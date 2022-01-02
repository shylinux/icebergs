package chrome

import (
	"shylinux.com/x/ice"
)

type spide struct {
	cache

	download string `name:"download" help:"下载"`
	list     string `name:"list wid tid url auto insert" help:"节点"`
}

func (s spide) Download(m *ice.Message, arg ...string) {
	m.Cmdy(s.cache.Create, arg).ProcessHold()
}
func (s spide) List(m *ice.Message, arg ...string) {
	if s.Spide(m, arg...); len(arg) > 1 {
		m.PushAction(s.Download)
	}
}

func init() { ice.CodeCtxCmd(spide{}) }
