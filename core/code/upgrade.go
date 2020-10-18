package code

import (
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
			UPGRADE: {Name: UPGRADE, Help: "升级", Value: kit.Dict(kit.MDB_HASH, kit.Dict(
				kit.MDB_PATH, "usr/upgrade", "system", kit.Dict(kit.MDB_LIST, kit.List(
					kit.MDB_INPUT, "bin", "file", "ice.bin", "path", "bin/ice.bin",
					kit.MDB_INPUT, "bin", "file", "ice.sh", "path", "bin/ice.sh",
				)),
			))},
		},
		Commands: map[string]*ice.Command{
			UPGRADE: {Name: "upgrade item:select=system 执行:button", Help: "升级", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				exit := false
				m.Grows(cmd, kit.Keys(kit.MDB_HASH, kit.Select("system", arg, 0)), "", "", func(index int, value map[string]interface{}) {
					if value[kit.MDB_FILE] == "ice.bin" {
						// 程序文件
						value[kit.MDB_FILE] = kit.Keys("ice", m.Conf(cli.RUNTIME, "host.GOOS"), m.Conf(cli.RUNTIME, "host.GOARCH"))
						exit = true
					}

					// 下载文件
					msg := m.Cmd(web.SPIDE, web.SPIDE_DEV, web.SPIDE_CACHE, web.SPIDE_GET, "/publish/"+kit.Format(value[kit.MDB_FILE]))
					m.Cmd(web.STORY, web.WATCH, msg.Append(kit.MDB_FILE), value[kit.MDB_PATH])
					os.Chmod(kit.Format(value[kit.MDB_PATH]), 0770)
				})
				if exit {
					m.Sleep("1s").Go(func() { m.Cmd("exit", 1) })
				}
			}},
		},
	}, nil)
}
