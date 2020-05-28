package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
)

func _inner_save(m *ice.Message, name, text string) {
	if f, e := os.Create(name); m.Assert(e) {
		defer f.Close()
		if n, e := f.WriteString(text); m.Assert(e) {
			m.Logs(ice.LOG_EXPORT, "file", name, "size", n)
		}
	}
}

func init() {
	const (
		INNER  = "inner"
		SAVE   = "save"
		COMMIT = "commit"
	)

	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			INNER: {Name: "inner", Help: "编辑器", Value: kit.Data(kit.MDB_SHORT, INNER)},
		},
		Commands: map[string]*ice.Command{
			INNER: {Name: "inner path=auto auto", Help: "编辑器", Action: map[string]*ice.Action{
				SAVE: {Name: "save name content", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_inner_save(m, arg[0], kit.Select(m.Option("content"), arg, 1))
				}},
				COMMIT: {Name: "commit name", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					web.StoryCatch(m, "", arg[0])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy("nfs.dir", arg)
			}},
		},
	}, nil)
}
