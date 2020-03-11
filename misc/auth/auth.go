package auth

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

var Index = &ice.Context{Name: "auth", Help: "auth",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"auth": {Name: "auth", Help: "auth", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("auth")
		}},

		"new": {Name: "new user [secrete]", Help: "创建动态码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Richs("auth", nil, "*", func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{"name", "username", "website"})
				})
				return
			}

			m.Append("_output", "qrcode")
			if len(arg) == 1 {
				if m.Richs("auth", nil, arg[0], func(key string, value map[string]interface{}) {
					m.Echo("otpauth://totp/%s?secret=%s", value["name"], value["secrete"])
				}) != nil {
					return
				}
				arg = append(arg, gen())
			}

			data := kit.Dict("name", arg[0], "secrete", arg[1])
			for i := 2; i < len(arg)-1; i += 2 {
				kit.Value(data, arg[i], arg[i+1])
			}

			n := m.Rich("auth", nil, data)
			m.Log(ice.LOG_CREATE, "%s", n)

			m.Cmdy(ice.WEB_SHARE, "optauth", arg[0], kit.Format("otpauth://totp/%s?secret=%s", arg[0], arg[1]))
		}},
		"get": {Name: "get user [secrete]", Help: "获取动态码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Richs("auth", nil, "*", func(key string, value map[string]interface{}) {
					m.Push("name", value["name"])
					m.Push("code", m.Cmdx("get", value["name"], value["secrete"]))
				})
				return
			}

			if len(arg) == 1 {
				m.Richs("auth", nil, arg[0], func(key string, value map[string]interface{}) {
					arg = append(arg, kit.Format(value["secrete"]))
				})
			}
			if len(arg) == 1 {
				arg = append(arg, m.Cmdx("new", arg[0]))
			}

			m.Echo(get(arg[1]))
		}},
	},
}

func init() { aaa.Index.Register(Index, nil) }
