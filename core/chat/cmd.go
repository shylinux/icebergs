package chat

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const CMD = "cmd"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CMD: {Name: CMD, Help: "命令", Value: kit.Data(kit.MDB_PATH, "./")},
		},
		Commands: map[string]*ice.Command{
			"/cmd/": {Name: "/cmd/", Help: "命令", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) == 0 {
						m.Push("index", "cmd")
						m.Push("args", "")
						return
					}
					m.Cmdy(ctx.COMMAND, arg[0])
				}},
				cli.RUN: {Name: "command", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if strings.HasSuffix(m.R.URL.Path, "/") {
					m.RenderDownload(path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, kit.MDB_PATH)), "page/cmd.html"))
					return
				}
				m.RenderDownload(path.Join(m.Conf(CMD, kit.META_PATH), path.Join(arg...)))
			}},
			"cmd": {Name: "cmd path auto up", Help: "命令", Action: map[string]*ice.Action{
				"up": {Name: "up", Help: "上一级", Hand: func(m *ice.Message, arg ...string) {
					if strings.TrimPrefix(m.R.URL.Path, "/cmd") == "/" {
						m.Cmdy("cmd")
						return
					}
					if strings.HasSuffix(m.R.URL.Path, "/") {
						m.Process("_location", "../")
					} else {
						m.Process("_location", "./")
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Process("_location", arg[0])
					return
				}
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(CMD, kit.META_PATH), strings.TrimPrefix(path.Dir(m.R.URL.Path), "/cmd")))
				m.Cmdy(nfs.DIR, arg)
			}},
		}})
}
