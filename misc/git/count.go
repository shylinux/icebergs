package git

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

const COUNT = "count"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		COUNT: {Name: "count path auto count", Help: "统计", Action: map[string]*ice.Action{
			COUNT: {Name: "count", Help: "计数", Hand: func(m *ice.Message, arg ...string) {
				files := map[string]int{}
				lines := map[string]int{}
				m.Option(nfs.DIR_DEEP, ice.TRUE)
				m.Option(nfs.DIR_TYPE, nfs.TYPE_CAT)
				m.Cmdy(nfs.DIR, arg, func(file string) {
					if strings.Contains(file, "bin/") {
						return
					}
					switch kit.Ext(file) {
					case "sum", "log":
						return
					}

					files["total"]++
					files[kit.Ext(file)]++
					m.Cmdy(nfs.CAT, file, func(text string, line int) {
						if kit.Ext(file) == "go" {
							switch {
							case strings.HasPrefix(text, "func"):
								lines["_func"]++
							case strings.HasPrefix(text, "type"):
								lines["_type"]++
							}
						}

						lines["total"]++
						lines[kit.Ext(file)]++
					})
				})
				for k := range lines {
					m.Push("type", k)
					m.Push("files", files[k])
					m.Push("lines", lines[k])
				}
				m.SortIntR("lines")
				m.StatusTime()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(nfs.DIR, arg)
		}},
	}})
}
