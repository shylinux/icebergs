package wework

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const BOT = "bot"

type Bot struct {
	short string `data:"name"`
	field string `data:"time,name,token,ekey,hook"`

	create string `name:"list name token ekey hook" help:"创建"`
	list   string `name:"list name chat text:textarea auto create" help:"机器人"`
}

func (b Bot) Create(m *ice.Message, arg ...string) {
	m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)
}
func (b Bot) List(m *ice.Message, arg ...string) {
	m.Fields(len(arg), m.Config(kit.MDB_FIELD))
	m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, m.Config(kit.MDB_SHORT), arg)
	if len(arg) > 2 {
		m.Cmd(web.SPIDE, mdb.CREATE, arg[0], m.Append("hook"))
		m.SetAppend()
		m.Cmdy(web.SPIDE, arg[0], "", kit.Format(kit.Dict(
			"chatid", arg[1],
			"msgtype", "text", "text", kit.Dict(
				"content", arg[2],
			),
		)))
	}
}

func init() { ice.Cmd("web.chat.wework.bot", Bot{}) }
