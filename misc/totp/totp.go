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
	"math"
	"strings"
	"time"
)

func gen(per int64) string {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, time.Now().Unix()/per)
	b := hmac.New(sha1.New, buf.Bytes()).Sum(nil)
	return strings.ToUpper(base32.StdEncoding.EncodeToString(b[:]))
}
func get(key string, num int, per int64) string {
	now := kit.Int64(time.Now().Unix() / per)

	buf := []byte{}
	for i := 0; i < 8; i++ {
		buf = append(buf, byte((now >> ((7 - i) * 8))))
	}

	if l := len(key) % 8; l != 0 {
		key += strings.Repeat("=", 8-l)
	}
	s, _ := base32.StdEncoding.DecodeString(strings.ToUpper(key))

	hm := hmac.New(sha1.New, s)
	hm.Write(buf)
	b := hm.Sum(nil)

	n := b[len(b)-1] & 0x0F
	res := int64(b[n]&0x7F)<<24 | int64(b[n+1]&0xFF)<<16 | int64(b[n+2]&0xFF)<<8 | int64(b[n+3]&0xFF)
	return kit.Format(kit.Format("%%0%dd", num), res%int64(math.Pow10(num)))
}

var Index = &ice.Context{Name: "totp", Help: "动态码",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"totp": {Name: "totp", Help: "动态码", Value: kit.Data(kit.MDB_SHORT, "name", "share", "otpauth://totp/%s?secret=%s")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

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
				if len(arg) > 1 {
					m.Render(ice.RENDER_QRCODE, kit.Format(m.Conf("totp", "meta.share"), value["name"], value["text"]))
				} else {
					m.Push("detail", value)
				}
			}) != nil {
				return
			}

			if len(arg) == 1 {
				// 创建密钥
				arg = append(arg, gen(30))
			}

			// 添加密钥
			m.Log(ice.LOG_CREATE, "%s: %s", arg[0], m.Rich("totp", nil, kit.Dict(
				kit.MDB_NAME, arg[0], kit.MDB_TEXT, arg[1], kit.MDB_EXTRA, kit.Dict(arg[2:]),
			)))
		}},
		"get": {Name: "get [user [number [period]]]", Help: "获取密码", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 密码列表
				m.Richs("totp", nil, "*", func(key string, value map[string]interface{}) {
					m.Push("name", value["name"])
					m.Push("code", get(kit.Format(value["text"]), kit.Int(kit.Select("6", value["number"])), kit.Int64(kit.Select("30", value["period"]))))
				})
				return
			}

			m.Richs("totp", nil, arg[0], func(key string, value map[string]interface{}) {
				// 获取密码
				m.Echo(get(kit.Format(value["text"]), kit.Int(kit.Select("6", arg, 1)), kit.Int64(kit.Select("30", arg, 2))))
			})
		}},
	},
}

func init() { aaa.Index.Register(Index, nil) }