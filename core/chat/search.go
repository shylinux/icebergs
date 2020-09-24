package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/search": {Name: "/search", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] == "render" {
					m.Cmdy(m.Space(m.Option("pod")), mdb.RENDER, arg[1:])
					return
				}
				m.Cmdy(m.Space(m.Option("pod")), mdb.SEARCH, arg)
			}},
		},
	}, nil)
}
