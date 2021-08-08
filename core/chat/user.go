package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const USERS = "users"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		USERS: {Name: "users username auto insert invite", Help: "用户", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert username", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, USERS), mdb.HASH, arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, USERS), mdb.HASH, m.OptionSimple(aaa.USERNAME))
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(aaa.USER, ice.OptionFields(aaa.USERNAME, aaa.USERZONE, aaa.USERNICK))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), "time,username")
			m.Cmdy(mdb.SELECT, RIVER, _river_key(m, USERS), mdb.HASH, aaa.USERNAME, arg)
			m.Table(func(index int, value map[string]string, head []string) {
				m.Richs(USERS, nil, value[aaa.USERNAME], func(key string, val map[string]interface{}) {
					val = kit.GetMeta(val)
					m.Push(aaa.USERNICK, val[aaa.USERNICK])
					m.PushImages(aaa.AVATAR, kit.Format(val[aaa.AVATAR]), kit.Select("60", "240", m.Option(mdb.FIELDS) == mdb.DETAIL))
				})
			})
			m.PushAction(mdb.REMOVE)
		}},
	}})
}
