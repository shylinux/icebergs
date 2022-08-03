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
	Index.Merge(&ice.Context{Configs: ice.Configs{
		UPGRADE: {Name: UPGRADE, Help: "升级", Value: kit.Dict(mdb.HASH, kit.Dict(
			nfs.TARGET, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, "bin", nfs.FILE, "ice.bin")),
			nfs.SOURCE, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, "tar", nfs.FILE, "contexts.src.tar.gz")),
			nfs.BINARY, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, "tar", nfs.FILE, "contexts.bin.tar.gz")),
		))},
	}, Commands: ice.Commands{
		UPGRADE: {Name: "upgrade item=target,source,binary run restart", Help: "升级", Actions: ice.Actions{
			cli.RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
				m.Sleep("1s").Go(func() { m.Cmd(ice.EXIT, 1) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			mdb.Grows(m, m.CommandKey(), kit.Keys(mdb.HASH, kit.Select(cli.SYSTEM, arg, 0)), "", "", func(index int, value ice.Map) {
				if value[nfs.FILE] == ice.ICE_BIN { // 程序文件
					value[nfs.FILE] = kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH)
					defer m.Cmd(cli.SYSTEM, "mv", value[nfs.FILE], ice.BIN_ICE_BIN)
					m.Option(ice.EXIT, ice.TRUE)
				}

				// 下载文件
				dir := kit.Select(kit.Format(value[nfs.FILE]), value[nfs.PATH])
				m.Cmd(web.SPIDE, ice.DEV, web.SPIDE_SAVE, dir, web.SPIDE_GET, "/publish/"+kit.Format(value[nfs.FILE]))
				switch value[mdb.TYPE] {
				case "bin":
					os.Chmod(dir, 0755)
				case "tar":
					m.Cmd(nfs.TAR, mdb.EXPORT, dir, "-C", path.Dir(dir))
				}
			})
			if web.ToastSuccess(m); m.Option(ice.EXIT) == ice.TRUE {
				m.Sleep("1s").Go(func() { m.Cmd(ice.EXIT, 1) })
				web.ToastRestart(m)
			}
		}},
	}})
}
