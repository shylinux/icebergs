package totp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"

	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"math"
	"strings"
	"time"
)

func _totp_gen(per int64) string {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, time.Now().Unix()/per)
	b := hmac.New(sha1.New, buf.Bytes()).Sum(nil)
	return strings.ToUpper(base32.StdEncoding.EncodeToString(b[:]))
}
func _totp_get(key string, num int, per int64) string {
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

const TOTP = "totp"
const (
	NEW = "new"
	GET = "get"
)

var Index = &ice.Context{Name: TOTP, Help: "动态码",
	Configs: map[string]*ice.Config{
		TOTP: {Name: TOTP, Help: "动态码", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_LINK, "otpauth://totp/%s?secret=%s",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		NEW: {Name: "new user [secret]", Help: "创建密钥", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 密钥列表
				m.Richs(TOTP, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_NAME})
				})
				return
			}

			if m.Richs(TOTP, nil, arg[0], func(key string, value map[string]interface{}) {
				// 密钥详情
				if len(arg) > 1 {
					m.Render(ice.RENDER_QRCODE, kit.Format(m.Conf(TOTP, "meta.link"), value[kit.MDB_NAME], value[kit.MDB_TEXT]))
				} else {
					m.Push("detail", value)
				}
			}) != nil {
				return
			}

			if len(arg) == 1 {
				// 创建密钥
				arg = append(arg, _totp_gen(30))
			}

			// 添加密钥
			m.Log(ice.LOG_CREATE, "%s: %s", arg[0], m.Rich(TOTP, nil, kit.Dict(
				kit.MDB_NAME, arg[0], kit.MDB_TEXT, arg[1], kit.MDB_EXTRA, kit.Dict(arg[2:]),
			)))
		}},
		GET: {Name: "get [name [number [period]]] auto", Help: "获取密码", Meta: kit.Dict(
			"_refresh", "1000",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				// 密码列表
				m.Richs(TOTP, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
					per := kit.Int64(kit.Select("30", value["period"]))
					m.Push("time", m.Time())
					m.Push("rest", per-time.Now().Unix()%per)
					m.Push("name", value["name"])
					m.Push("code", _totp_get(kit.Format(value["text"]), kit.Int(kit.Select("6", value["number"])), per))

				})
				m.Sort(kit.MDB_NAME)
				return
			}

			m.Richs(TOTP, nil, arg[0], func(key string, value map[string]interface{}) {
				// 获取密码
				m.Echo(_totp_get(kit.Format(value[kit.MDB_TEXT]), kit.Int(kit.Select("6", arg, 1)), kit.Int64(kit.Select("30", arg, 2))))
			})
		}},
	},
}

func init() { aaa.Index.Register(Index, nil) }
