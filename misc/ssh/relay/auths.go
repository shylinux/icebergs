package relay

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/misc/ssh"
	kit "shylinux.com/x/toolkits"
)

const (
	SSH_AUTHS = "ssh.auths"
)

type auths struct {
	list string `name:"list auto"`
}

func (s auths) List(m *ice.Message, arg ...string) {
	list := map[string]map[string]bool{}
	head := []string{}
	m.AdminCmd(web.DREAM, web.ORIGIN).Table(func(val ice.Maps) {
		head = append(head, val[mdb.NAME])
		m.AdminCmd(web.SPACE, val[mdb.NAME], ssh.RSA, ssh.AUTHS).Table(func(value ice.Maps) {
			if _, ok := list[value[mdb.NAME]]; !ok {
				list[value[mdb.NAME]] = map[string]bool{}
			}
			list[value[mdb.NAME]][val[mdb.NAME]] = true
		})
	})
	m.AdminCmd(web.DREAM, web.SERVER).Table(func(value ice.Maps) {
		kit.For(head, func(key string) {
			if data, ok := list[value[mdb.NAME]]; ok && data[key] {
				m.Push(key, "ok")
			} else {
				m.Push(key, "")
			}
		})
	})
}

func init() { ice.Cmd(SSH_AUTHS, auths{}) }
