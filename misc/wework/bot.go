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
	Index.Merge(&ice.Context{Configs: ice.Configs{
		BOT: {Name: "bot", Help: "机器人", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,token,ekey,hook",
		)},
	}, Commands: ice.Commands{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, arg ...string) {}},
		"/bot/": {Name: "/bot/", Help: "机器人", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(BOT, arg[0])

			check := kit.Sort([]string{msg.Append("token"), m.Option("nonce"), m.Option("timestamp"), m.Option("echostr")})
			sig := kit.Format(sha1.Sum([]byte(strings.Join(check, ""))))
			if m.Warn(sig != m.Option("msg_signature"), ice.ErrNotRight, check, sig) {
				// return
			}

			aeskey, err := base64.StdEncoding.DecodeString(msg.Append("ekey"))
			m.Debug("what %v %v", msg.Append("ekey"), aeskey)
			m.Assert(err)

			en_msg, err := base64.StdEncoding.DecodeString(m.Option("echostr"))
			m.Debug("what %v", en_msg)
			m.Assert(err)

			block, err := aes.NewCipher(aeskey)
			m.Assert(err)

			mode := cipher.NewCBCDecrypter(block, aeskey[:aes.BlockSize])
			mode.CryptBlocks(en_msg, en_msg)
			m.Debug("what %v", en_msg)
			m.RenderResult(en_msg)
		}},
		BOT: {Name: "bot name chat text:textarea auto create", Help: "机器人", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create name token ekey hook", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, m.Option("name"), m.Option("hook"))
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, arg)
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...); len(arg) > 2 {
				m.Cmdy(web.SPIDE, arg[0], "", kit.Format(kit.Dict(
					"chatid", arg[1], "msgtype", "text", "text.content", arg[2],
				)))
			} else {
				m.PushAction(mdb.REMOVE)
			}
		}},
	}})
}
