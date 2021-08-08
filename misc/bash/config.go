package bash

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/config": {Name: "/config", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd("web.code.git.config").Table(func(index int, value map[string]string, head []string) {
				if strings.HasPrefix(value[kit.MDB_NAME], "url") {
					m.Echo(`git config --global "%s" "%s"`, value[kit.MDB_NAME], value[kit.MDB_VALUE])
					m.Echo(ice.NL)
				}
			})
		}},
	}})
}
