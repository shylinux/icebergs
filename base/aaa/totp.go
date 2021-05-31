package aaa

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"math"
	"strings"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
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
		buf = append(buf, byte((uint64(now) >> uint64(((7 - i) * 8)))))
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
	PERIOD = "period"
	NUMBER = "number"
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
				mdb.CREATE: {Name: "create name secret period=30 number=6", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(SECRET) == "" { // 创建密钥
						m.Option(SECRET, _totp_gen(kit.Int64(m.Option(PERIOD))))
					}

					m.Cmd(mdb.INSERT, TOTP, "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME),
						SECRET, m.Option(SECRET), PERIOD, m.Option(PERIOD), NUMBER, m.Option(NUMBER),
					)
					m.ProcessRefresh("1ms")
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, TOTP, "", mdb.HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) == 0, "time,name,secret,period,number")
				m.Cmd(mdb.SELECT, TOTP, "", mdb.HASH, kit.MDB_NAME, arg).Table(func(index int, value map[string]string, head []string) {
					m.Push(kit.MDB_TIME, m.Time())
					m.Push(kit.MDB_NAME, value[kit.MDB_NAME])

					period := kit.Int64(value[PERIOD])
					m.Push("rest", period-time.Now().Unix()%period)
					m.Push("code", _totp_get(value[SECRET], kit.Int(value[NUMBER]), period))

					if len(arg) > 0 {
						m.PushQRCode("scan", kit.Format(m.Conf(TOTP, kit.Keym(kit.MDB_LINK)), value[kit.MDB_NAME], value[SECRET]))
						m.Echo(_totp_get(value[SECRET], kit.Int(value[NUMBER]), kit.Int64(value[PERIOD])))
					}
				})
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
