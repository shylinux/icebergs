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

func _cmd_render(m *ice.Message, cmd string, args ...interface{}) {
	list := []interface{}{kit.Dict("index", cmd, "args", args)}
	m.RenderResult(kit.Format(m.Conf(CMD, kit.Keym(kit.MDB_TEMPLATE)), kit.Format(list)))
}

const CMD = "cmd"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/cmd/": {Name: "/cmd/", Help: "命令", Action: map[string]*ice.Action{
				ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) == 0 {
						m.Push("index", CMD)
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
					m.RenderDownload(path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, kit.MDB_PATH)), m.Conf(CMD, kit.Keym(kit.MDB_INDEX))))
					return
				}

				switch p := path.Join(m.Conf(CMD, kit.META_PATH), path.Join(arg...)); kit.Ext(p) {
				case "svg":
					_cmd_render(m, "web.wiki.draw", path.Dir(p)+"/", path.Base(p))
				case "json":
					_cmd_render(m, "web.wiki.json", p)
				case "shy":
					_cmd_render(m, "web.wiki.word", p)
				case "csv":
					_cmd_render(m, "web.wiki.data", p)
				case "go", "mod", "sum":
					_cmd_render(m, "web.code.inner", path.Dir(p)+"/", path.Base(p))
				default:
					m.RenderDownload(p)
				}
			}},
			CMD: {Name: "cmd path auto upload up home", Help: "命令", Action: map[string]*ice.Action{
				web.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					_action_upload(m)
					m.Upload(path.Join(m.Conf(CMD, kit.META_PATH), strings.TrimPrefix(path.Dir(m.R.URL.Path), "/cmd")))
				}},
				"home": {Name: "home", Help: "根目录", Hand: func(m *ice.Message, arg ...string) {
					m.ProcessLocation("/chat/cmd/")
				}},
				"up": {Name: "up", Help: "上一级", Hand: func(m *ice.Message, arg ...string) {
					if strings.TrimPrefix(m.R.URL.Path, "/cmd") == "/" {
						m.Cmdy(CMD)
						return
					}
					if strings.HasSuffix(m.R.URL.Path, "/") {
						m.ProcessLocation("../")
					} else {
						m.ProcessLocation("./")
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.ProcessLocation(arg[0])
					return
				}
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(CMD, kit.META_PATH), strings.TrimPrefix(path.Dir(m.R.URL.Path), "/cmd")))
				m.Cmdy(nfs.DIR, arg)
			}},
		},
		Configs: map[string]*ice.Config{
			CMD: {Name: CMD, Help: "命令", Value: kit.Data(
				kit.MDB_PATH, "./", kit.MDB_INDEX, "page/cmd.html", kit.MDB_TEMPLATE, `
<!DOCTYPE html>
<body><script src="/page/can.js"></script>
	<script>app("/chat/", %s)</script>
</body>
`,
			)},
		},
	})
}
