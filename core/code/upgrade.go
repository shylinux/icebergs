package code

import (
	"os"
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const UPGRADE = "upgrade"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			UPGRADE: {Name: UPGRADE, Help: "升级", Value: kit.Dict(kit.MDB_HASH, kit.Dict(
				cli.SYSTEM, kit.Dict(kit.MDB_LIST, kit.List(
					kit.MDB_TYPE, "bin", kit.MDB_FILE, "ice.sh", kit.MDB_PATH, ice.BIN_ICE_SH,
					kit.MDB_TYPE, "bin", kit.MDB_FILE, "ice.bin", kit.MDB_PATH, ice.BIN_ICE_BIN,
				)),
				cli.SOURCE, kit.Dict(kit.MDB_LIST, kit.List(
					kit.MDB_TYPE, "txt", kit.MDB_FILE, "main.go", kit.MDB_PATH, ice.SRC_MAIN_GO,
					kit.MDB_TYPE, "txt", kit.MDB_FILE, "miss.sh", kit.MDB_PATH, ice.ETC_MISS_SH,
					kit.MDB_TYPE, "txt", kit.MDB_FILE, "go.mod", kit.MDB_PATH, ice.GO_MOD,
				)),
			))},
		},
		Commands: map[string]*ice.Command{
			UPGRADE: {Name: "upgrade item=system,source run:button", Help: "升级", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Grows(cmd, kit.Keys(kit.MDB_HASH, kit.Select(cli.SYSTEM, arg, 0)), "", "", func(index int, value map[string]interface{}) {
					if value[kit.MDB_PATH] == ice.BIN_ICE_BIN { // 程序文件
						value[kit.MDB_FILE] = kit.Keys("ice", runtime.GOOS, runtime.GOARCH)
						m.Option("exit", ice.TRUE)
					}

					// 下载文件
					msg := m.Cmd(web.SPIDE, ice.DEV, web.SPIDE_CACHE, web.SPIDE_GET, "/publish/"+kit.Format(value[kit.MDB_FILE]))
					m.Cmd(web.STORY, web.WATCH, msg.Append(kit.MDB_FILE), value[kit.MDB_PATH])
					os.Chmod(kit.Format(value[kit.MDB_PATH]), 0770)
				})
				if m.Option("exit") == ice.TRUE {
					m.Sleep("1s").Go(func() { m.Cmd("exit", 1) })
				}
			}},
		},
	})
}
