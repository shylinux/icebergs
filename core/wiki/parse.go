package wiki

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const PARSE = "parse"

func init() {
	Index.MergeCommands(ice.Commands{
		PARSE: {Name: "parse type=auto,base64,json,http,form,time,list auto text", Help: "解析", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 {
				return
			}
			if arg[1] = strings.TrimSpace(arg[1]); arg[0] == ice.AUTO {
				if strings.HasPrefix(arg[1], "{") || strings.HasPrefix(arg[1], "[") {
					arg[0] = nfs.JSON
				} else if strings.HasPrefix(arg[1], web.HTTP) {
					arg[0] = web.HTTP
				} else if strings.Contains(arg[1], mdb.EQ) {
					arg[0] = web.FORM
				} else if _, e := strconv.ParseInt(arg[1], 10, 64); e == nil {
					arg[0] = mdb.TIME
				} else {
					arg[0] = mdb.LIST
				}
			}
			switch m.OptionFields(mdb.DETAIL); arg[0] {
			case nfs.JSON:
				m.Echo(kit.Formats(kit.UnMarshal(arg[1])))
				ctx.DisplayStoryJSON(m)
			case web.HTTP:
				u := kit.ParseURL(arg[1])
				m.Push(tcp.PROTO, u.Scheme).Push(tcp.HOST, u.Host).Push(nfs.PATH, u.Path)
				kit.For(u, func(k string, v []string) { m.Push(k, v) })
				m.EchoQRCode(arg[1])
			case web.FORM:
				kit.SplitKV("=", "&", arg[1], func(k string, v []string) {
					kit.For(v, func(v string) { m.Push(kit.QueryUnescape(k), kit.QueryUnescape(v)) })
				})
			case mdb.TIME:
				if i, e := strconv.ParseInt(arg[1], 10, 64); e == nil {
					m.Echo(time.Unix(i, 0).Format(ice.MOD_TIME))
				}
			case mdb.LIST:
				kit.For(kit.Split(arg[1]), func(i int, v string) { m.Push(kit.Format(i), v) })
			case "base64":
				if buf, err := base64.StdEncoding.DecodeString(arg[1]); err == nil {
					m.Echo(hex.EncodeToString(buf))
				}
			}
		}},
	})
}
