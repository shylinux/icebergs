package totp

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/toolkits"

	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"strings"
	"time"
)

func gen() string {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, time.Now().Unix()/30)
	b := hmac.New(sha1.New, buf.Bytes()).Sum(nil)
	return strings.ToUpper(base32.StdEncoding.EncodeToString(b[:]))
}
func get(key string) string {
	buf := []byte{}
	now := kit.Int64(time.Now().Unix() / 30)
	for i := 0; i < 8; i++ {
		buf = append(buf, byte((now >> ((7 - i) * 8))))
	}

	s, _ := base32.StdEncoding.DecodeString(strings.ToUpper(key))

	hm := hmac.New(sha1.New, s)
	hm.Write(buf)
	b := hm.Sum(nil)

	o := b[len(b)-1] & 0x0F
	p := b[o : o+4]
	p[0] = p[0] & 0x7f
	res := int64(p[0])<<24 + int64(p[1])<<16 + int64(p[2])<<8 + int64(p[3])
	return kit.Format("%06d", res%1000000)
}

var Index = &ice.Context{Name: "totp", Help: "动态码",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"totp": {Name: "totp", Help: "动态码", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("totp")
		}},

		"new": {Name: "new user [secret]", Help: "创建密钥", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 密钥列表
				m.Richs("totp", nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"time", "name"})
				})
				return
			}

			if m.Richs("totp", nil, arg[0], func(key string, value map[string]interface{}) {
				// 密钥详情
				m.Push("detail", value)
			}) != nil {
				return
			}

			if len(arg) == 1 {
				// 创建密钥
				arg = append(arg, gen())
			}

			// 添加密钥
			m.Log(ice.LOG_CREATE, "%s: %s", arg[0], m.Rich("totp", nil, kit.Dict(
				kit.MDB_NAME, arg[0], kit.MDB_TEXT, arg[1], kit.MDB_EXTRA, kit.Dict(arg[2:]),
			)))

			// 创建共享
			defer m.Cmdy(ice.WEB_SHARE, "optauth", arg[0], kit.Format("otpauth://totp/%s?secret=%s", arg[0], arg[1]))
		}},
		"get": {Name: "get user [secret]", Help: "获取密码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 密码列表
				m.Richs("totp", nil, "*", func(key string, value map[string]interface{}) {
					m.Push("name", value["name"])
					m.Push("code", m.Cmdx("get", value["name"], value["text"]))
				})
				return
			}

			if len(arg) == 1 {
				// 获取密钥
				m.Richs("totp", nil, arg[0], func(key string, value map[string]interface{}) {
					arg = append(arg, kit.Format(value["text"]))
				})
			}
			if len(arg) == 1 {
				// 创建密钥
				arg = append(arg, m.Cmdx("new", arg[0]))
			}

			// 获取密码
			m.Echo(get(arg[1]))
		}},
	},
}

func init() { aaa.Index.Register(Index, nil) }
