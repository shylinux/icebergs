package code

import (
	"os"
	"path"
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
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
			ctx.CONFIG, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.SHY, nfs.FILE, ice.ETC_LOCAL_SHY)),
			COMPILE, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.TAR, nfs.FILE, "go1.15.5", nfs.PATH, ice.USR_LOCAL)),
		), mdb.META, kit.Dict(mdb.FIELD, "type,file,path"))},
	}, Commands: ice.Commands{
		UPGRADE: {Name: "upgrade item=target,config,binary,source,compile run restart", Help: "升级", Actions: ice.MergeActions(ice.Actions{
			cli.RESTART: {Hand: func(m *ice.Message, arg ...string) { m.Go(func() { m.Sleep300ms(ice.EXIT, 1) }) }},
		}), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelect(m.Spawn(), kit.Select(nfs.TARGET, arg, 0)).Table(func(value ice.Maps) {
				if kit.Select("", arg, 0) == COMPILE {
					value[nfs.FILE] = kit.Keys(kit.Format(value[nfs.FILE]), runtime.GOOS+"-"+runtime.GOARCH, kit.Select("tar.gz", "zip", runtime.GOOS == cli.WINDOWS))
				}
				if value[nfs.FILE] == ice.ICE_BIN && os.Getenv(cli.CTX_POD) == "" {
					value[nfs.FILE] = kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH)
					defer nfs.Rename(m, value[nfs.FILE], ice.BIN_ICE_BIN)
					m.Option(ice.EXIT, ice.TRUE)
				}
				uri := "/publish/" + kit.Format(value[nfs.FILE])
				if os.Getenv(cli.CTX_POD) != "" {
					uri = kit.MergeURL2(os.Getenv(cli.CTX_DEV), "/chat/pod/"+os.Getenv(cli.CTX_POD), cli.GOOS, runtime.GOOS, cli.GOARCH, runtime.GOARCH)
				}
				dir := path.Join(kit.Format(value[nfs.PATH]), kit.Format(value[nfs.FILE]))
				switch web.SpideSave(m, dir, uri, nil); value[mdb.TYPE] {
				case nfs.TAR:
					m.Cmd(cli.SYSTEM, nfs.TAR, "xf", dir, "-C", path.Dir(dir))
					// m.Cmd(nfs.TAR, mdb.EXPORT, dir, "-C", path.Dir(dir))
				case ice.BIN:
					os.Chmod(dir, ice.MOD_DIR)
				}
				m.Cmdy(nfs.DIR, dir, "time,size,path,hash")
			})
			if web.ToastSuccess(m); m.Option(ice.EXIT) == ice.TRUE {
				m.Cmd("", cli.RESTART)
				web.Toast(m, cli.RESTART)
			}
		}},
	}})
}
