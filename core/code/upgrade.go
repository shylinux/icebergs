package code

import (
	"os"
	"path"
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const UPGRADE = "upgrade"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		UPGRADE: {Name: UPGRADE, Help: "升级", Value: kit.Dict(mdb.HASH, kit.Dict(
			cli.SYSTEM, kit.Dict(mdb.LIST, kit.List(
				mdb.TYPE, "bin", nfs.FILE, "ice.bin", nfs.PATH, ice.BIN_ICE_BIN,
			)),
			nfs.BINARY, kit.Dict(mdb.LIST, kit.List(
				mdb.TYPE, "txt", nfs.FILE, "path", nfs.PATH, ice.ETC_PATH,
				mdb.TYPE, "tar", nfs.FILE, "contexts.bin.tar.gz",
			)),
			nfs.SOURCE, kit.Dict(mdb.LIST, kit.List(
				mdb.TYPE, "sh", nfs.FILE, "miss.sh", nfs.PATH, ice.ETC_MISS_SH,
				mdb.TYPE, "txt", nfs.FILE, "main.go", nfs.PATH, ice.SRC_MAIN_GO,
				mdb.TYPE, "txt", nfs.FILE, "go.mod",
				mdb.TYPE, "txt", nfs.FILE, "go.sum",
			)),
		))},
	}, Commands: map[string]*ice.Command{
		UPGRADE: {Name: "upgrade item=system,binary,source run", Help: "升级", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Grows(cmd, kit.Keys(mdb.HASH, kit.Select(cli.SYSTEM, arg, 0)), "", "", func(index int, value map[string]interface{}) {
				if value[nfs.FILE] == ice.ICE_BIN { // 程序文件
					value[nfs.FILE] = kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH)
					m.Option(ice.EXIT, ice.TRUE)
				}

				// 下载文件
				dir := kit.Select(kit.Format(value[nfs.FILE]), value[nfs.PATH])
				msg := m.Cmd(web.SPIDE, ice.DEV, web.SPIDE_CACHE, web.SPIDE_GET, "/publish/"+kit.Format(value[nfs.FILE]))
				m.Cmd(web.STORY, web.WATCH, msg.Append(nfs.FILE), dir)
				switch value[mdb.TYPE] {
				case "sh", "bin":
					os.Chmod(kit.Format(dir), 0770)
				case "tar":
					m.Cmd(nfs.TAR, mdb.EXPORT, dir, "-C", path.Dir(dir))
				}
			})
			if m.Option(ice.EXIT) == ice.TRUE {
				m.Sleep("1s").Go(func() { m.Cmd(ice.EXIT, 1) })
			}
		}},
	}})
}
