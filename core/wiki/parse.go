package wiki

import (
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
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
				} else if strings.Contains(arg[1], "=") {
					arg[0] = web.FORM
				} else if _, e := strconv.ParseInt(arg[1], 10, 64); e == nil {
					arg[0] = mdb.TIME
				} else {
					arg[0] = mdb.LIST
				}
			}

			switch m.OptionFields(mdb.DETAIL); arg[0] {
			case "base64":
				if buf, err := base64.StdEncoding.DecodeString(arg[1]); err == nil {
					m.Echo(hex.EncodeToString(buf))
				}
			case nfs.JSON:
				m.Echo(kit.Formats(kit.UnMarshal(arg[1])))
			case web.HTTP:
				u, _ := url.Parse(arg[1])
				m.Push(tcp.PROTO, u.Scheme)
				m.Push(tcp.HOST, u.Host)
				m.Push(nfs.PATH, u.Path)
				for k, v := range u.Query() {
					for _, v := range v {
						m.Push(k, v)
					}
				}
				m.EchoQRCode(arg[1])
			case web.FORM:
				for _, v := range strings.Split(arg[1], "&") {
					ls := strings.Split(v, ice.EQ)
					m.Push(kit.QueryUnescape(ls[0]), kit.QueryUnescape(kit.Select("", ls, 1)))
				}
			case mdb.TIME:
				if i, e := strconv.ParseInt(arg[1], 10, 64); e == nil {
					m.Echo(time.Unix(i, 0).Format(ice.MOD_TIME))
				}
			case mdb.LIST:
				for i, v := range kit.Split(arg[1]) {
					m.Push(kit.Format(i), v)
				}
			}
		}},
	})
}
