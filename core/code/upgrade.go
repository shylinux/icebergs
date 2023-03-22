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
		UPGRADE: {Value: kit.Dict(mdb.HASH, kit.Dict(
			nfs.TARGET, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, ice.BIN, nfs.FILE, ice.ICE_BIN)),
			nfs.BINARY, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.TAR, nfs.FILE, "contexts.bin.tar.gz")),
			nfs.SOURCE, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.TAR, nfs.FILE, "contexts.src.tar.gz")),
			COMPILE, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.TAR, nfs.FILE, "go1.15.5", nfs.PATH, ice.USR_LOCAL)),
		), mdb.META, kit.Dict(mdb.FIELD, "type,file,path"))},
	}, Commands: ice.Commands{
		UPGRADE: {Name: "upgrade item=target,binary,source,compile run restart", Help: "升级", Actions: ice.MergeActions(ice.Actions{
			cli.RESTART: {Hand: func(m *ice.Message, arg ...string) { m.Go(func() { m.Sleep300ms(ice.EXIT, 1) }) }},
		}), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelect(m.Spawn(), kit.Select(nfs.TARGET, arg, 0)).Tables(func(value ice.Maps) {
				if value[nfs.FILE] == ice.ICE_BIN {
					value[nfs.FILE] = kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH)
					defer nfs.Rename(m, value[nfs.FILE], ice.BIN_ICE_BIN)
					m.Option(ice.EXIT, ice.TRUE)
				}
				if kit.Select("", arg, 0) == COMPILE {
					value[nfs.FILE] = kit.Keys(kit.Format(value[nfs.FILE]), runtime.GOOS+"-"+runtime.GOARCH, kit.Select("tar.gz", "zip", runtime.GOOS == cli.WINDOWS))
				}
				dir := path.Join(kit.Format(value[nfs.PATH]), kit.Format(value[nfs.FILE]))
				switch web.SpideSave(m, dir, "/publish/"+kit.Format(value[nfs.FILE]), nil); value[mdb.TYPE] {
				case nfs.TAR:
					m.Cmd(cli.SYSTEM, nfs.TAR, "xf", dir, "-C", path.Dir(dir))
					// m.Cmd(nfs.TAR, mdb.EXPORT, dir, "-C", path.Dir(dir))
				case ice.BIN:
					os.Chmod(dir, 0755)
				}
				m.Cmdy(nfs.DIR, dir, "time,size,path,hash")
			})
			if web.ToastSuccess(m); m.Option(ice.EXIT) == ice.TRUE {
				m.Cmd("", cli.RESTART)
				web.ToastRestart(m)
			}
		}},
	}})
}
