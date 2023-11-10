package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	TAGID = "tagid"
)
const TAGS = "tags"

func init() {
	Index.MergeCommands(ice.Commands{
		TAGS: {Name: "tags access id openid auto", Help: "标签", Meta: Meta(), Actions: ice.Actions{
			mdb.CREATE: {Name: "create name*", Hand: func(m *ice.Message, arg ...string) {
				SpidePost(m, TAGS_CREATE, "tag.name", m.Option(mdb.NAME))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				SpidePost(m, TAGS_DELETE, "tag.id", m.Option(mdb.ID))
			}},
			mdb.UPDATE: {Name: "update name*", Hand: func(m *ice.Message, arg ...string) {
				SpidePost(m, TAGS_UPDATE, "tag.id", m.Option(mdb.ID), "tag.name", m.Option(mdb.NAME))
			}},
			mdb.DELETE: {Hand: func(m *ice.Message, arg ...string) {
				SpidePost(m, TAGS_MEMBERS_BATCHUNTAGGING, TAGID, m.Option(mdb.ID), "openid_list.0", m.Option(OPENID))
			}},
			REMARK: {Name: "remark remark", Help: "备注", Hand: func(m *ice.Message, arg ...string) {
				SpidePost(m, USER_REMARK, m.OptionSimple(OPENID, REMARK))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction("").Option(ice.MSG_ACTION, "")
			} else if m.Options(ACCESS, arg[0]); len(arg) == 1 {
				res := SpideGet(m, TAGS_GET)
				kit.For(kit.Value(res, TAGS), func(value ice.Map) {
					m.Push("", value, []string{mdb.ID, mdb.NAME, mdb.COUNT})
				})
				m.PushAction(mdb.UPDATE, mdb.REMOVE).Action(mdb.CREATE)
			} else if len(arg) == 2 {
				_user_list(m, SpidePost(m, USER_TAG_GET, TAGID, arg[1])).PushAction(REMARK, mdb.DELETE)
			} else {
				_user_info(m, arg[2])
			}
		}},
	})
}
