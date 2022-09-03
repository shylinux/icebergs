package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: {Name: "count path auto count", Help: "代码行", Actions: ice.Actions{
			COUNT: {Name: "count", Help: "计数", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 || arg[0] == "usr/" {
					m.Echo("to many file, please choice sub dir")
					return
				}
				files := map[string]int{}
				lines := map[string]int{}
				m.Option(nfs.DIR_DEEP, ice.TRUE)
				m.Option(nfs.DIR_TYPE, nfs.TYPE_CAT)
				m.Cmdy(nfs.DIR, arg, func(file string) {
					if strings.Contains(file, "node_modules/") {
						return
					}
					if strings.Contains(file, "bin/") {
						return
					}
					if strings.Contains(file, "var/") {
						return
					}
					if strings.Contains(file, "tags") {
						return
					}
					switch kit.Ext(file) {
					case "sum", "log":
						return
					}

					files[mdb.TOTAL]++
					files[kit.Ext(file)]++
					m.Cmdy(nfs.CAT, file, func(text string, line int) {
						if kit.Ext(file) == code.GO {
							switch {
							case strings.HasPrefix(text, "func"):
								lines["_func"]++
							case strings.HasPrefix(text, "type"):
								lines["_type"]++
							}
						}

						lines[mdb.TOTAL]++
						lines[kit.Ext(file)]++
					})
				})
				for k := range lines {
					m.Push(mdb.TYPE, k)
					m.Push("files", files[k])
					m.Push("lines", lines[k])
				}
				m.SortIntR("lines")
				m.StatusTime()
			}},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg) }},
	})
}
