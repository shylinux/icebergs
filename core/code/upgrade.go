package code

import (
	"net/http"
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
)

const UPGRADE = "upgrade"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			UPGRADE: {Name: "upgrade", Help: "升级", Value: kit.Dict(kit.MDB_HASH, kit.Dict(
				"path", "usr/upgrade", "system", kit.Dict(kit.MDB_LIST, kit.List(
					kit.MDB_INPUT, "bin", "file", "ice.bin", "path", "bin/ice.bin",
					kit.MDB_INPUT, "bin", "file", "ice.sh", "path", "bin/ice.sh",
				)),
			))},
		},
		Commands: map[string]*ice.Command{
			UPGRADE: {Name: "upgrade item", Help: "升级", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				exit := false
				m.Grows(cmd, kit.Keys(kit.MDB_HASH, kit.Select("system", arg, 0)), "", "", func(index int, value map[string]interface{}) {
					if value[kit.MDB_FILE] == "ice.bin" {
						// 程序文件
						value[kit.MDB_FILE] = kit.Keys("ice", m.Conf(cli.RUNTIME, "host.GOOS"), m.Conf(cli.RUNTIME, "host.GOARCH"))
						exit = true
					}

					// 下载文件
					h := m.Cmdx(web.SPIDE, "dev", web.CACHE, http.MethodGet, "/publish/"+kit.Format(value[kit.MDB_FILE]))
					if h == "" {
						exit = false
						return
					}

					// 升级记录
					m.Cmd(web.STORY, web.CATCH, "bin", value[kit.MDB_PATH], h)
					m.Cmd(web.STORY, web.WATCH, h, path.Join(m.Conf(UPGRADE, "meta.path"), kit.Format(value[kit.MDB_PATH])))
					m.Cmd(web.STORY, web.WATCH, h, value[kit.MDB_PATH])
					os.Chmod(kit.Format(value[kit.MDB_PATH]), 0770)
				})
				if exit {
					m.Sleep("1s").Gos(m, func(m *ice.Message) { m.Cmd("exit", 1) })
				}
			}},
		},
	}, nil)
}
