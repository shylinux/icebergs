package bash

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/configs": {Name: "/configs", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("web.code.git.configs").Table(func(index int, value map[string]string, head []string) {
				if strings.HasPrefix(value[mdb.NAME], "url") {
					m.Echo(`git config --global "%s" "%s"`, value[mdb.NAME], value[mdb.VALUE])
					m.Echo(ice.NL)
				}
			})
		}},
	}})
}
