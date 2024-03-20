package ssh

import (
	"fmt"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	psh "shylinux.com/x/icebergs/base/ssh"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const AUTH = "auth"
	psh.Index.MergeCommands(ice.Commands{
		AUTH: {Name: "auth name auto", Help: "授权", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create authkey*:textarea", Hand: func(m *ice.Message, arg ...string) {
				if ls := kit.Split(m.Option(AUTHKEY)); len(ls) > 2 {
					mdb.HashCreate(m, mdb.TYPE, ls[0], mdb.NAME, ls[len(ls)-1], mdb.TEXT, strings.Join(ls[1:len(ls)-1], "+"))
				}
			}},
			nfs.LOAD: {Name: "load authkey*=.ssh/authorized_keys", Icon: "bi bi-folder-plus", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.CAT, kit.HomePath(m.Option(AUTHKEY)), func(pub string) { m.Cmd("", mdb.CREATE, pub) })
			}},
			nfs.SAVE: {Name: "save authkey*=.ssh/authorized_keys", Hand: func(m *ice.Message, arg ...string) {
				list := []string{}
				m.Cmds("").Table(func(value ice.Maps) {
					list = append(list, fmt.Sprintf("%s %s %s", value[mdb.TYPE], value[mdb.TEXT], value[mdb.NAME]))
				})
				if len(list) > 0 {
					m.Cmd(nfs.SAVE, kit.HomePath(m.Option(AUTHKEY)), strings.Join(list, lex.NL)+lex.NL)
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE, nfs.LOAD, nfs.SAVE)
			}
		}},
	})
}
