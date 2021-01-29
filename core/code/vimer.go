package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"path"
)

const VIMER = "vimer"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 刷新:button=auto save", Help: "编辑器", Meta: kit.Dict(
			"display", "/plugin/local/code/vimer.js", "style", "editor",
		), Action: map[string]*ice.Action{
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.SAVE, path.Join(m.Option(kit.MDB_PATH), m.Option(kit.MDB_FILE)))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Cmdy(INNER, arg) }},
	}})
}
