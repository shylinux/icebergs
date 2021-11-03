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
	kit "shylinux.com/x/toolkits"
)

const PARSE = "parse"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		PARSE: {Name: "parse type=auto,base64,json,http,form,time,list auto text", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) < 2 {
				return
			}

			if arg[1] = strings.TrimSpace(arg[1]); arg[0] == "auto" {
				if strings.HasPrefix(arg[1], "{") || strings.HasPrefix(arg[1], "[") {
					arg[0] = "json"
				} else if strings.HasPrefix(arg[1], "http") {
					arg[0] = "http"
				} else if strings.Contains(arg[1], "=") {
					arg[0] = "form"
				} else if _, e := strconv.ParseInt(arg[1], 10, 64); e == nil {
					arg[0] = "time"
				} else {
					arg[0] = "list"
				}
			}

			switch m.OptionFields(mdb.DETAIL); arg[0] {
			case "base64":
				buf, err := base64.StdEncoding.DecodeString(arg[1])
				if err == nil {
					m.Echo(hex.EncodeToString(buf))
				}

			case "json":
				m.Echo(kit.Formats(kit.UnMarshal(arg[1])))

			case "http":
				u, _ := url.Parse(arg[1])
				m.Push("proto", u.Scheme)
				m.Push("host", u.Host)
				m.Push("path", u.Path)
				for k, v := range u.Query() {
					for _, v := range v {
						m.Push(k, v)
					}
				}
				m.EchoQRCode(arg[1])

			case "form":
				for _, v := range kit.Split(arg[1], "&", "&", "&") {
					ls := kit.Split(v, "=", "=", "=")
					key, _ := url.QueryUnescape(ls[0])
					value, _ := url.QueryUnescape(kit.Select("", ls, 1))
					m.Push(key, value)
				}

			case "time":
				if i, e := strconv.ParseInt(arg[1], 10, 64); e == nil {
					m.Echo(time.Unix(i, 0).Format(ice.MOD_TIME))
				}

			case "list":
				for i, v := range kit.Split(arg[1]) {
					m.Push(kit.Format(i), v)
				}
			}
		}},
	}})
}
