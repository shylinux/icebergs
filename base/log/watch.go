package log

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const WATCH = "watch"

func init() {
	Index.MergeCommands(ice.Commands{
		WATCH: {Name: "watch auto", Help: "记录", Hand: func(m *ice.Message, arg ...string) {
			operate := map[string]int{}
			for _, line := range strings.Split(m.Cmdx(nfs.CAT, path.Join(ice.VAR_LOG, "watch.log")), ice.NL) {
				ls := kit.Split(line, "", " ", " ")
				if len(ls) < 5 {
					continue
				}
				m.Push(mdb.TIME, ls[0]+ice.SP+ls[1])
				m.Push("order", ls[2])
				m.Push("ship", ls[3])
				m.Push("source", kit.Slice(ls, -1)[0])
				m.Push("operate", ls[4])
				m.Push("content", kit.Join(kit.Slice(ls, 5, -1), ice.SP))
				operate[ls[4]]++
			}
			m.StatusTimeCount(operate)
		}},
	})
}
