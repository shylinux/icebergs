package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/" + mdb.SEARCH: {Name: "/search", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 2 {
					m.Cmdy(m.Space(m.Option("pod")), mdb.RENDER, arg)
					return
				}
				m.Cmdy(m.Space(m.Option("pod")), mdb.SEARCH, arg)
			}},
		},
	}, nil)
}
