package code

import (
	"fmt"
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
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _publish_file(m *ice.Message, file string, arg ...string) string {
	if strings.HasSuffix(file, ice.ICE_BIN) {
		file, arg = cli.SystemFind(m, os.Args[0]), kit.Simple(kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH))
	} else if s, e := nfs.StatFile(m, file); m.Assert(e) && s.IsDir() {
		file = m.Cmdx(nfs.TAR, mdb.IMPORT, path.Base(file), file)
		defer func() { nfs.Remove(m, file) }()
	}
	target := path.Join(ice.USR_PUBLISH, kit.Select(path.Base(file), arg, 0))
	m.Logs(mdb.EXPORT, PUBLISH, target, cli.FROM, file)
	return m.Cmdx(nfs.LINK, target, file)
}
func _publish_list(m *ice.Message, arg ...string) {
	m.Option(nfs.DIR_REG, kit.Select("", arg, 0))
	nfs.DirDeepAll(m, ice.USR_PUBLISH, nfs.PWD, nil, kit.Select(nfs.DIR_WEB_FIELDS, arg, 1))
}
func _publish_bin_list(m *ice.Message, dir string) {
	p := m.Option(cli.CMD_DIR, dir)
	for _, ls := range strings.Split(cli.SystemCmds(m, "ls |xargs file |grep executable"), ice.NL) {
		if file := strings.TrimSpace(strings.Split(ls, ice.DF)[0]); file != "" {
			if s, e := nfs.StatFile(m, path.Join(p, file)); e == nil {
				m.Push(mdb.TIME, s.ModTime()).Push(nfs.SIZE, kit.FmtSize(s.Size())).Push(nfs.PATH, file)
				m.PushDownload(mdb.LINK, file, path.Join(p, file)).PushButton(nfs.TRASH)
			}
		}
	}
	m.SortTimeR(mdb.TIME)
}
func PublishScript(m *ice.Message, arg ...string) {
	u := web.OptionUserWeb(m)
	host := tcp.PublishLocalhost(m, strings.Split(u.Host, ice.DF)[0])
	m.Option(cli.CTX_ENV, kit.Select("", ice.SP+kit.JoinKV(ice.EQ, ice.SP, cli.CTX_POD, m.Option(ice.MSG_USERPOD)), m.Option(ice.MSG_USERPOD) != ""))
	m.Option(web.DOMAIN, fmt.Sprintf("%s://%s:%s", u.Scheme, host, kit.Select(kit.Select("443", "80", u.Scheme == ice.HTTP), strings.Split(u.Host, ice.DF), 1)))
	for _, v := range arg {
		m.EchoScript(kit.Renders(v, m))
	}
}
func _publish_contexts(m *ice.Message, arg ...string) {
	u := web.OptionUserWeb(m)
	host := tcp.PublishLocalhost(m, strings.Split(u.Host, ice.DF)[0])
	m.Option(cli.CTX_ENV, kit.Select("", ice.SP+kit.JoinKV(ice.EQ, ice.SP, cli.CTX_POD, m.Option(ice.MSG_USERPOD)), m.Option(ice.MSG_USERPOD) != ""))
	m.Option(web.DOMAIN, fmt.Sprintf("%s://%s:%s", u.Scheme, host, kit.Select(kit.Select("443", "80", u.Scheme == ice.HTTP), strings.Split(u.Host, ice.DF), 1)))
	for _, k := range kit.Default(arg, ice.MISC) {
		switch k {
		case INSTALL:
			m.Echo(kit.Renders(`export ctx_dev={{.Option "domain"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); wget -O $ctx_temp -q $ctx_dev; source $ctx_temp app username {{.Option "user.name"}}`, m))
			return
		case ice.MISC:
			_publish_file(m, ice.ICE_BIN)
		case ice.CORE:
			m.Option(web.DOMAIN, m.Cmdx(web.SPIDE, ice.SHY, "client.origin"))
		case ice.BASE:
			m.Option(web.DOMAIN, m.Cmdx(web.SPIDE, ice.SHY, "client.origin"))
			m.Option(nfs.REMOTE, kit.Select(ice.Info.Make.Remote, cli.SystemExec(m, "git", "config", "remote.origin.url")))
		}
		if buf, err := kit.Render(m.Config(kit.Keys(ice.CONTEXTS, k)), m); m.Assert(err) {
			m.EchoScript(strings.TrimSpace(string(buf)))
		}
	}
}

const (
	GIT = "git"
)
const PUBLISH = "publish"

func init() {
	Index.MergeCommands(ice.Commands{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell", Help: "发布", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Config(ice.CONTEXTS, _contexts) }},
			web.SERVE_START: {Hand: func(m *ice.Message, arg ...string) {
				if runtime.GOOS == cli.WINDOWS {
					return
				}
				_publish_file(m, ice.ICE_BIN)
			}},
			ice.VOLCANOS: {Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.EchoQRCode(m.Option(ice.MSG_USERWEB)) }()
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.MISC) }()
				_publish_list(m, kit.ExtReg(`(html|css|js)`))
			}},
			ice.ICEBERGS: {Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.CORE) }()
				_publish_bin_list(m, ice.USR_PUBLISH)
			}},
			ice.INTSHELL: {Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.BASE) }()
				_publish_list(m, kit.ExtReg(`(sh|vim|conf)`))
			}},
			ice.CONTEXTS: {Hand: func(m *ice.Message, arg ...string) { _publish_contexts(m, arg...) }},
			mdb.INPUTS:   {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(nfs.DIR, arg[1:]).Cut("path,size,time").ProcessAgain() }},
			mdb.CREATE:   {Name: "create file", Help: "添加", Hand: func(m *ice.Message, arg ...string) { _publish_file(m, m.Option(nfs.FILE)) }},
			nfs.TRASH:    {Hand: func(m *ice.Message, arg ...string) { nfs.Trash(m, path.Join(ice.USR_PUBLISH, m.Option(nfs.PATH))) }},
		}, ctx.ConfAction(ice.CONTEXTS, _contexts), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(nfs.DIR, kit.Select("", arg, 0), nfs.DIR_WEB_FIELDS, kit.Dict(nfs.DIR_ROOT, ice.USR_PUBLISH)).SortTimeR(mdb.TIME)
		}},
	})
}

var _contexts = kit.Dict(
	ice.MISC, `
# 下载工具 wget Alpine
export ctx_dev={{.Option "domain"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); wget -O $ctx_temp -q $ctx_dev; source $ctx_temp app username {{.Option "user.name"}} usernick {{.Option "user.nick"}}

# 下载工具 curl Centos / MacOS
export ctx_dev={{.Option "domain"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL $ctx_dev; source $ctx_temp app username {{.Option "user.name"}} usernick {{.Option "user.nick"}}
`,
	ice.CORE, `
# 下载命令 wget Busybox
ctx_temp=$(mktemp); wget -O $ctx_temp -q http://shylinux.com; source $ctx_temp binary

# 下载命令 wget Alpine
ctx_temp=$(mktemp); wget -O $ctx_temp -q {{.Option "domain"}}; source $ctx_temp binary

# 下载命令 curl Centos / MacOS
ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL {{.Option "domain"}}; source $ctx_temp binary
`,
	ice.BASE, `
# 下载源码 wget Alpine
ctx_temp=$(mktemp); wget -O $ctx_temp -q {{.Option "domain"}}; source $ctx_temp source

# 下载源码 curl Centos / MacOS
ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL {{.Option "domain"}}; source $ctx_temp source
`,
)
