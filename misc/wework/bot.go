package wework

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const BOT = "bot"

func init() {
	Index.MergeCommands(ice.Commands{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) {}},
		"/bot/": {Name: "/bot/", Help: "机器人", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(BOT, arg[0])

			check := kit.Sort([]string{msg.Append("token"), m.Option("nonce"), m.Option("timestamp"), m.Option("echostr")})
			sig := kit.Format(sha1.Sum([]byte(strings.Join(check, ""))))
			if m.WarnNotRight(sig != m.Option("msg_signature"), check, sig) {
				// return
			}

			aeskey, err := base64.StdEncoding.DecodeString(msg.Append("ekey"))
			m.Assert(err)

			en_msg, err := base64.StdEncoding.DecodeString(m.Option("echostr"))
			m.Assert(err)

			block, err := aes.NewCipher(aeskey)
			m.Assert(err)

			mode := cipher.NewCBCDecrypter(block, aeskey[:aes.BlockSize])
			mode.CryptBlocks(en_msg, en_msg)
			m.RenderResult(en_msg)
		}},
		BOT: {Name: "bot name chat text:textarea auto create", Help: "机器人", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name token ekey hook", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, m.Option("hook"), m.Option("name"))
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,token,ekey,hook")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 2 {
				m.Cmdy(web.SPIDE, arg[0], "", kit.Format(kit.Dict(
					"chatid", arg[1], "msgtype", "text", "text.content", arg[2],
				)))
			} else {
				m.PushAction(mdb.REMOVE)
			}
		}},
	})
}
