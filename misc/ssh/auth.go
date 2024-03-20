package ssh

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	psh "shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const (
		AUTH = "auth"
	)
	psh.Index.MergeCommands(ice.Commands{
		AUTH: {Name: "auth list", Help: "连接", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(nfs.CAT, kit.HomePath(m.Option(AUTHKEY)), func(pub string) {
				if ls := kit.Split(pub); len(ls) > 2 {
					m.Push(mdb.TYPE, ls[0])
					m.Push(mdb.NAME, ls[len(ls)-1])
					m.Push(mdb.TEXT, strings.Join(ls[1:len(ls)-1], "+"))
				}
			})
		}},
	})
}
