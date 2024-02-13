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
		UPGRADE: {Value: kit.Dict(
			mdb.META, kit.Dict(mdb.FIELDS, "type,file,path"),
			mdb.HASH, kit.Dict(
				nfs.TARGET, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, ice.BIN, nfs.FILE, ice.ICE_BIN)),
				ctx.CONFIG, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.SHY, nfs.FILE, ice.ETC_LOCAL_SHY)),
				nfs.BINARY, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.TAR, nfs.FILE, "contexts.bin.tar.gz")),
				nfs.SOURCE, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.TAR, nfs.FILE, "contexts.src.tar.gz")),
				COMPILE, kit.Dict(mdb.LIST, kit.List(mdb.TYPE, nfs.TAR, nfs.FILE, "go1.15.5", nfs.PATH, ice.USR_LOCAL)),
			)),
		},
	}, Commands: ice.Commands{
		UPGRADE: {Name: "upgrade item=target,config,binary,source,compile run restart", Help: "升级", Actions: ice.Actions{
			cli.RESTART: {Hand: func(m *ice.Message, arg ...string) { m.Go(func() { m.Sleep30ms(ice.EXIT, 1) }) }},
		}, Hand: func(m *ice.Message, arg ...string) {
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
				m.Debug("what %v", uri)
				m.Debug("what %v", ice.Info.Make.Domain)
				kit.If(m.Spawn().Options(ice.MSG_USERPOD, "").ParseLink(ice.Info.Make.Domain).Option(ice.MSG_USERPOD), func(p string) {
					m.Debug("what %v", p)
					uri = kit.MergeURL(uri, ice.POD, p)
				})
				m.Debug("what %v", uri)
				dir := path.Join(kit.Format(value[nfs.PATH]), kit.Format(value[nfs.FILE]))
				web.GoToast(m, web.DOWNLOAD, func(toast func(name string, count, total int)) []string {
					switch web.SpideSave(m, dir, uri, func(count, total, value int) {
						toast(dir, count, total)
					}); value[mdb.TYPE] {
					case nfs.TAR:
						m.Cmd(cli.SYSTEM, nfs.TAR, "xf", dir, "-C", path.Dir(dir))
						// m.Cmd(nfs.TAR, mdb.EXPORT, dir, "-C", path.Dir(dir))
					case ice.BIN:
						os.Chmod(dir, ice.MOD_DIR)
					}
					return nil
				})
				m.Cmdy(nfs.DIR, dir, "time,size,path,hash").Push(web.ORIGIN, kit.MergeURL2(web.SpideOrigin(m, ice.DEV_IP), uri))
			})
			if web.ToastSuccess(m); m.Option(ice.EXIT) == ice.TRUE {
				web.Toast(m, cli.RESTART)
				m.Cmd("", cli.RESTART)
			}
		}},
	}})
}
