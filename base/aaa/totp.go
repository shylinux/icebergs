package aaa

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
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

func TOTP_GET(key string, num int, per int64) string { return _totp_get(key, num, per) }

const (
	SECRET = "secret"
	NUMBER = "number"
	PERIOD = "period"
)
const TOTP = "totp"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TOTP: {Name: TOTP, Help: "动态令牌", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_LINK, "otpauth://totp/%s?secret=%s",
			)},
		},
		Commands: map[string]*ice.Command{
			TOTP: {Name: "totp name auto create", Help: "动态令牌", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create user secret period=30 number=6", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(SECRET) == "" { // 创建密钥
						m.Option(SECRET, _totp_gen(kit.Int64(m.Option(PERIOD))))
					}

					// 添加密钥
					m.Cmd(mdb.INSERT, TOTP, "", mdb.HASH, kit.MDB_NAME, m.Option("user"),
						SECRET, m.Option(SECRET), PERIOD, m.Option(PERIOD), NUMBER, m.Option(NUMBER),
					)
					m.Cmdy("web.wiki.image", "qrcode", kit.Format(m.Conf(TOTP, "meta.link"), m.Option("user"), m.Option(SECRET)))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, TOTP, "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 密码列表
					m.Option(ice.MSG_PROCESS, "_refresh")
					m.Option(mdb.FIELDS, "time,name,secret,period,number")
					m.Cmd(mdb.SELECT, TOTP, "", mdb.HASH).Table(func(index int, value map[string]string, head []string) {
						per := kit.Int64(value[PERIOD])
						m.Push("time", m.Time())
						m.Push("rest", per-time.Now().Unix()%per)
						m.Push("name", value["name"])
						m.Push("code", _totp_get(value[SECRET], kit.Int(value[NUMBER]), per))
					})
					m.PushAction(mdb.REMOVE)
					m.Sort(kit.MDB_NAME)
					return
				}

				// 获取密码
				m.Cmd(mdb.SELECT, TOTP, "", mdb.HASH, kit.MDB_NAME, arg[0]).Table(func(index int, value map[string]string, head []string) {
					m.Echo(_totp_get(value[SECRET], kit.Int(value[NUMBER]), kit.Int64(value[PERIOD])))
				})
			}},
		},
	}, nil)
}
