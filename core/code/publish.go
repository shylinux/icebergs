package code

import (
	"os"
	"path"
	"runtime"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _publish_bin_list(m *ice.Message, p string) *ice.Message {
	m.Option(cli.CMD_DIR, p)
	defer m.SortTimeR(mdb.TIME)
	for _, ls := range strings.Split(cli.SystemCmds(m, "ls |xargs file |grep executable"), ice.NL) {
		if file := strings.TrimSpace(strings.Split(ls, ice.DF)[0]); file != "" {
			if s, e := nfs.StatFile(m, path.Join(p, file)); e == nil {
				m.Push(mdb.TIME, s.ModTime()).Push(nfs.SIZE, kit.FmtSize(s.Size())).Push(nfs.PATH, file)
				m.PushDownload(mdb.LINK, file, path.Join(p, file)).PushButton(nfs.TRASH)
			}
		}
	}
	return m
}
func _publish_list(m *ice.Message, arg ...string) *ice.Message {
	m.Option(nfs.DIR_REG, kit.Select("", arg, 0))
	return nfs.DirDeepAll(m, ice.USR_PUBLISH, nfs.PWD, nil, kit.Select(nfs.DIR_WEB_FIELDS, arg, 1))
}
func _publish_file(m *ice.Message, file string, arg ...string) string {
	if strings.HasSuffix(file, ice.ICE_BIN) {
		file, arg = cli.SystemFind(m, os.Args[0]), kit.Simple(kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH))
	} else if s, e := nfs.StatFile(m, file); m.Assert(e) && s.IsDir() {
		file = m.Cmdx(nfs.TAR, mdb.IMPORT, path.Base(file), file)
		defer func() { nfs.Remove(m, file) }()
	}
	target := path.Join(ice.USR_PUBLISH, kit.Select(path.Base(file), arg, 0))
	return m.Logs(mdb.EXPORT, PUBLISH, target, nfs.FROM, file).Cmdx(nfs.LINK, target, file)
}
func _publish_contexts(m *ice.Message, arg ...string) {
	for _, k := range kit.Default(arg, ice.MISC) {
		switch k {
		case ice.BASE:
			m.Option(web.DOMAIN, m.Cmd(web.SPIDE, ice.SHY).Append(web.CLIENT_ORIGIN))
		case ice.CORE:
			m.Option(web.DOMAIN, m.Cmd(web.SPIDE, ice.DEV).Append(web.CLIENT_ORIGIN))
		default:
			m.Options(web.DOMAIN, m.Option(ice.MSG_USERHOST), cli.CTX_ENV, kit.Select("", ice.SP+kit.JoinKV(ice.EQ, ice.SP, cli.CTX_POD, m.Option(ice.MSG_USERPOD)), m.Option(ice.MSG_USERPOD) != ""))
			_publish_file(m, ice.ICE_BIN)
		}
		if s := strings.TrimSpace(kit.Renders(m.Config(kit.Keys(ice.CONTEXTS, k)), m)); k == INSTALL {
			m.Echo(s)
		} else {
			m.EchoScript(s)
		}
	}
}

const PUBLISH = "publish"

func init() {
	Index.MergeCommands(ice.Commands{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell", Help: "发布", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Config(ice.CONTEXTS, _contexts) }},
			ice.VOLCANOS: {Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				_publish_list(m, kit.ExtReg(`(html|css|js)`)).Cmdy("", ice.CONTEXTS, ice.MISC).Echo(ice.NL).EchoQRCode(m.Option(ice.MSG_USERWEB))
			}},
			ice.ICEBERGS: {Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				_publish_bin_list(m, ice.USR_PUBLISH).Cmdy("", ice.CONTEXTS, ice.CORE)
			}},
			ice.INTSHELL: {Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				_publish_list(m, kit.ExtReg(`(sh|vim|conf)`)).Cmdy("", ice.CONTEXTS, ice.BASE)
			}},
			ice.CONTEXTS: {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, arg...) }},
			mdb.INPUTS:   {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg[1:]).Cut("path,size,time") }},
			mdb.CREATE:   {Name: "create file", Hand: func(m *ice.Message, arg ...string) { _publish_file(m, m.Option(nfs.FILE)) }},
			nfs.TRASH:    {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, path.Join(ice.USR_PUBLISH, m.Option(nfs.PATH))) }},
		}, ctx.ConfAction(ice.CONTEXTS, _contexts), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.DIR, kit.Select("", arg, 0), nfs.DIR_WEB_FIELDS, kit.Dict(nfs.DIR_ROOT, ice.USR_PUBLISH)).SortTimeR(mdb.TIME)
		}},
	})
}

var _contexts = kit.Dict(
	ice.MISC, `export ctx_dev={{.Option "domain"}}{{.Option "ctx_env"}}; temp=$(mktemp); if curl -h &>/dev/null; then curl -o $temp -fsSL $ctx_dev; else wget -O $temp -q $ctx_dev; fi; source $temp app username {{.Option "user.name"}} usernick "{{.Option "user.nick"}}"`,
	ice.CORE, `temp=$(mktemp); if curl -h &>/dev/null; then curl -o $temp -fsSL {{.Option "domain"}}; else wget -O $temp -q {{.Option "domain"}}; fi; source $temp binary`,
	ice.BASE, `temp=$(mktemp); if curl -h &>/dev/null; then curl -o $temp -fsSL {{.Option "domain"}}; else wget -O $temp -q {{.Option "domain"}}; fi; source $temp source`,
)
