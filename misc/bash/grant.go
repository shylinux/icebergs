package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const GRANT = "grant"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"grant": {Name: "grant sid auto", Help: "授权", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 {
				if m.Confirm("grant "+arg[0]+"?") == ice.TRUE {
					m.Cmdy(SESS, mdb.MODIFY, GRANT, m.Option(ice.MSG_USERNAME), ice.Option{kit.MDB_HASH, arg[0]})
				}
			}
			m.Cmdy(SESS, arg)
		}},
	}})
}
