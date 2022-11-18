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
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _publish_file(m *ice.Message, file string, arg ...string) string {
	if strings.HasSuffix(file, ice.ICE_BIN) { // 打包应用
		arg = kit.Simple(kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH))
		file = cli.SystemFind(m, os.Args[0])

	} else if s, e := nfs.StatFile(m, file); m.Assert(e) && s.IsDir() {
		file = m.Cmdx(nfs.TAR, mdb.IMPORT, path.Base(file), file)
		defer func() { nfs.Remove(m, file) }()
	}

	// 发布文件
	target := path.Join(ice.USR_PUBLISH, kit.Select(path.Base(file), arg, 0))
	m.Logs(mdb.EXPORT, PUBLISH, target, cli.FROM, file)
	m.Cmd(nfs.LINK, target, file)
	return target
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
				m.Push(mdb.TIME, s.ModTime())
				m.Push(nfs.SIZE, kit.FmtSize(s.Size()))
				m.Push(nfs.PATH, file)
				m.PushDownload(mdb.LINK, file, path.Join(p, file))
				m.PushButton(nfs.TRASH)
			}
		}
	}
	m.SortTimeR(mdb.TIME)
}
func _publish_contexts(m *ice.Message, arg ...string) {
	u := web.OptionUserWeb(m)
	host := strings.Split(u.Host, ice.DF)[0]
	if host == tcp.LOCALHOST {
		host = m.Cmd(tcp.HOST).Append(aaa.IP)
	}
	m.Option(cli.CTX_ENV, kit.Select("", ice.SP+kit.JoinKV(ice.EQ, ice.SP, cli.CTX_POD, m.Option(ice.MSG_USERPOD)), m.Option(ice.MSG_USERPOD) != ""))
	m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, host, kit.Select(kit.Select("443", "80", u.Scheme == ice.HTTP), strings.Split(u.Host, ice.DF), 1)))

	if len(arg) == 0 {
		arg = append(arg, ice.MISC)
	}
	for _, k := range arg {
		switch k {
		case INSTALL:
			m.Echo(kit.Renders(`export ctx_dev={{.Option "httphost"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); wget -O $ctx_temp -q $ctx_dev; source $ctx_temp app username {{.Option "user.name"}}`, m))
			return

		case ice.MISC:
			_publish_file(m, ice.ICE_BIN)

		case ice.BASE:
			m.Option("remote", kit.Select(ice.Info.Make.Remote, cli.SystemExec(m, "git", "config", "remote.origin.url")))
			m.Option("pathname", strings.TrimSuffix(path.Base(m.Option("remote")), ".git"))
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
	Index.Merge(&ice.Context{Commands: ice.Commands{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell", Help: "发布", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Config(ice.CONTEXTS, _contexts)
			}},
			web.SERVE_START: {Name: "serve.start", Help: "服务启动", Hand: func(m *ice.Message, arg ...string) {
				_publish_file(m, ice.ICE_BIN)
			}},
			ice.VOLCANOS: {Name: "volcanos", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.EchoQRCode(m.Option(ice.MSG_USERWEB)) }()
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.MISC) }()
				_publish_list(m, `.*\.(html|css|js)$`)
			}},
			ice.ICEBERGS: {Name: "icebergs", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.CORE) }()
				_publish_bin_list(m, ice.USR_PUBLISH)
			}},
			ice.INTSHELL: {Name: "intshell", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.BASE) }()
				_publish_list(m, `.*\.(sh|vim|conf)$`)
			}},
			ice.CONTEXTS: {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
				_publish_contexts(m, arg...)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, kit.Select(nfs.PWD, arg, 1)).ProcessAgain()
			}},
			mdb.CREATE: {Name: "create file", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_publish_file(m, m.Option(nfs.FILE))
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.TRASH, path.Join(ice.USR_PUBLISH, m.Option(nfs.PATH)))
			}},
		}, aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Option(nfs.DIR_ROOT, ice.USR_PUBLISH)
			m.Cmdy(nfs.DIR, kit.Select("", arg, 0), nfs.DIR_WEB_FIELDS)
		}},
	}, Configs: ice.Configs{PUBLISH: {Value: kit.Data(ice.CONTEXTS, _contexts)}}})
}

var _contexts = kit.Dict(
	ice.MISC, `# 下载命令 curl 或 wget
export ctx_dev={{.Option "httphost"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL $ctx_dev; source $ctx_temp app username {{.Option "user.name"}} usernick {{.Option "user.nick"}}
export ctx_dev={{.Option "httphost"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); wget -O $ctx_temp -q $ctx_dev; source $ctx_temp app username {{.Option "user.name"}} usernick {{.Option "user.nick"}}
`,
	ice.CORE, `# 下载工具 curl 或 wget
ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL {{.Cmdx "spide" "shy" "client.url"}}; source $ctx_temp binary
ctx_temp=$(mktemp); wget -O $ctx_temp -q {{.Cmdx "spide" "shy" "client.url"}}; source $ctx_temp binary
`,
	ice.BASE, `# 下载源码
git clone {{.Option "remote"}}; cd {{.Option "pathname"}} && source etc/miss.sh
`,
)
