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

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _totp_gen(per int64) string {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, time.Now().Unix()/per)
	b := hmac.New(sha1.New, buf.Bytes()).Sum(nil)
	return strings.ToUpper(base32.StdEncoding.EncodeToString(b[:]))
}
func _totp_get(key string, per int64, num int) string {
	buf, now := []byte{}, kit.Int64(time.Now().Unix()/per)
	kit.For(8, func(i int) { buf = append(buf, byte((uint64(now) >> uint64(((7 - i) * 8))))) })
	kit.If(len(key)%8, func(l int) { key += strings.Repeat(mdb.EQ, 8-l) })
	s, _ := base32.StdEncoding.DecodeString(strings.ToUpper(key))
	hm := hmac.New(sha1.New, s)
	hm.Write(buf)
	b := hm.Sum(nil)
	n := b[len(b)-1] & 0x0F
	res := int64(b[n]&0x7F)<<24 | int64(b[n+1]&0xFF)<<16 | int64(b[n+2]&0xFF)<<8 | int64(b[n+3]&0xFF)
	return kit.Format(kit.Format("%%0%dd", num), res%int64(math.Pow10(num)))
}

const (
	TOKEN = "token"
)
const TOTP = "totp"

func init() {
	const (
		NUMBER = "number"
		PERIOD = "period"
		SECRET = "secret"
	)
	Index.MergeCommands(ice.Commands{
		TOTP: {Name: "totp name auto", Help: "令牌", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name*=hi number*=6 period*=30 secret", Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(SECRET) == "", func() { m.Option(SECRET, _totp_gen(kit.Int64(m.Option(PERIOD)))) })
				mdb.HashCreate(m, m.OptionSimple(mdb.NAME, NUMBER, PERIOD, SECRET))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,number,period,secret", mdb.LINK, "otpauth://totp/%s?secret=%s")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m.Spawn(), arg...).Table(func(value ice.Maps) {
				kit.If(len(arg) > 0, func() { m.OptionFields(ice.FIELDS_DETAIL) })
				m.Push(mdb.TIME, m.Time()).Push(mdb.NAME, value[mdb.NAME])
				period := kit.Int64(value[PERIOD])
				m.Push(mdb.EXPIRE, period-time.Now().Unix()%period)
				m.Push(mdb.VALUE, _totp_get(value[SECRET], period, kit.Int(value[NUMBER])))
				if len(arg) > 0 {
					m.PushQRCode(mdb.SCAN, kit.Format(mdb.Config(m, mdb.LINK), value[mdb.NAME], value[SECRET]))
					m.Echo(m.Append(mdb.VALUE))
				}
			})
			if len(arg) == 0 {
				m.PushAction(mdb.REMOVE).Action(mdb.CREATE, mdb.PRUNES).StatusTimeCount()
			}
		}},
	})
}

func TOTP_GET(key string, per int64, num int) string { return _totp_get(key, per, num) }
