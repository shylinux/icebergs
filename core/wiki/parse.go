package wiki

import (
	"net/url"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const PARSE = "parse"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			PARSE: {Name: "parse type=auto,json,http,form,list auto text:textarea", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					return
				}
				if arg[0] == "auto" && (strings.HasPrefix(arg[1], "{") || strings.HasPrefix(arg[1], "[")) {
					arg[0] = "json"
				} else if strings.HasPrefix(arg[1], "http") {
					arg[0] = "http"
				} else if strings.Contains(arg[1], "=") {
					arg[0] = "form"
				} else {
					arg[0] = "list"
				}

				m.Option(mdb.FIELDS, mdb.DETAIL)
				switch arg[0] {
				case "json":
					m.Echo(kit.Formats(kit.UnMarshal(arg[1])))
				case "http":
					u, _ := url.Parse(arg[1])
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
				case "list":
					for i, v := range kit.Split(arg[1]) {
						m.Push(kit.Format(i), v)
					}
				}
			}},
		},
		Configs: map[string]*ice.Config{},
	})
}
