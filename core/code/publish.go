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
	"shylinux.com/x/icebergs/base/gdb"
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
	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_ROOT, ice.USR_PUBLISH)
	m.Option(nfs.DIR_REG, kit.Select("", arg, 0))
	m.Cmdy(nfs.DIR, nfs.PWD, kit.Select(nfs.DIR_WEB_FIELDS, arg, 1))
}
func _publish_bin_list(m *ice.Message, dir string) {
	p := m.Option(cli.CMD_DIR, dir)
	for _, ls := range strings.Split(strings.TrimSpace(m.Cmd(cli.SYSTEM, "bash", "-c", "ls |xargs file |grep executable").Append(cli.CMD_OUT)), ice.NL) {
		if file := strings.TrimSpace(strings.Split(ls, ":")[0]); file != "" {
			if s, e := nfs.StatFile(m, path.Join(p, file)); e == nil {
				m.Push(mdb.TIME, s.ModTime())
				m.Push(nfs.SIZE, kit.FmtSize(s.Size()))
				m.Push(nfs.FILE, file)
				m.PushDownload(mdb.LINK, file, path.Join(p, file))
			}
		}
	}
	m.SortTimeR(mdb.TIME)
}

const (
	GIT = "git"
)
const PUBLISH = "publish"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		PUBLISH: {Name: PUBLISH, Help: "发布", Value: kit.Data(ice.CONTEXTS, _contexts)},
	}, Commands: ice.Commands{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell", Help: "发布", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, ice.USR_PUBLISH)
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.PrefixKey())
				gdb.Watch(m, web.SERVE_START, m.PrefixKey())
				m.Config(ice.CONTEXTS, _contexts)
			}},
			web.SERVE_START: {Name: "serve.start", Help: "服务启动", Hand: func(m *ice.Message, arg ...string) {
				_publish_file(m, ice.ICE_BIN)
			}},
			ice.VOLCANOS: {Name: "volcanos", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.EchoQRCode(m.Option(ice.MSG_USERWEB)) }()
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS) }()
				_publish_list(m, `.*\.(html|css|js)$`)
			}},
			ice.ICEBERGS: {Name: "icebergs", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS) }()
				_publish_bin_list(m, ice.USR_PUBLISH)
			}},
			ice.INTSHELL: {Name: "intshell", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS) }()
				_publish_list(m, `.*\.(sh|vim|conf)$`)
			}},
			ice.CONTEXTS: {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
				u := web.OptionUserWeb(m)
				host := strings.Split(u.Host, ice.DF)[0]
				if host == tcp.LOCALHOST {
					host = m.Cmd(tcp.HOST).Append(aaa.IP)
				}
				m.Option("ctx_env", kit.Select("", " "+kit.JoinKV("=", " ", "ctx_pod", m.Option(ice.MSG_USERPOD)), m.Option(ice.MSG_USERPOD) != ""))
				m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, host, kit.Select(kit.Select("443", "80", u.Scheme == ice.HTTP), strings.Split(u.Host, ice.DF), 1)))

				if len(arg) == 0 {
					arg = append(arg, ice.MISC, ice.CORE, ice.BASE)
				}
				for _, k := range arg {
					switch k {
					case ice.MISC:
						if bin := path.Join(ice.USR_PUBLISH, kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH)); !nfs.ExistsFile(m, bin) {
							m.Cmd(nfs.LINK, bin, m.Cmdx(cli.RUNTIME, "boot.bin"))
						}

					case ice.CORE:
						if !nfs.ExistsFile(m, ".git") {
							repos := web.MergeURL2(m, "/x/"+kit.Select(ice.Info.PathName, m.Option(ice.MSG_USERPOD)))
							m.Cmd(cli.SYSTEM, "git", "init")
							m.Cmd(cli.SYSTEM, "git", "remote", "add", "origin", repos)
							m.Cmd("web.code.git.repos", mdb.CREATE, repos, "master", "", nfs.PWD)
						}
						m.Option("remote", kit.Select(ice.Info.Make.Remote, strings.TrimSpace(m.Cmdx(cli.SYSTEM, "git", "config", "remote.origin.url"))))
						m.Option("pathname", strings.TrimSuffix(path.Base(m.Option("remote")), ".git"))
					case ice.BASE:
					}
					if buf, err := kit.Render(m.Config(kit.Keys(ice.CONTEXTS, k)), m); m.Assert(err) {
						m.EchoScript(strings.TrimSpace(string(buf)))
					}
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, kit.Select(nfs.PWD, arg, 1)).ProcessAgain()
			}},
			mdb.CREATE: {Name: "create file", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_publish_file(m, m.Option(nfs.FILE))
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.TRASH, path.Join(ice.USR_PUBLISH, m.Option(nfs.PATH)))
			}},
			mdb.EXPORT: {Name: "export", Help: "工具链", Hand: func(m *ice.Message, arg ...string) {
				var list = []string{ice.ETC_PATH}
				m.Cmd(nfs.CAT, ice.ETC_PATH, func(text string) {
					if strings.HasPrefix(text, ice.USR_PUBLISH) {
						return
					}
					if strings.HasPrefix(text, ice.BIN) {
						return
					}
					if strings.HasPrefix(text, ice.PS) {
						return
					}
					list = append(list, text)
				})

				web.PushStream(m)
				defer m.ProcessHold()
				defer m.StatusTimeCount()
				defer web.ToastSuccess(m)
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "contexts.bin.tar.gz"), list)
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "contexts.src.tar.gz"), ice.MAKEFILE, ice.ETC_MISS_SH, ice.SRC_MAIN_GO, ice.GO_MOD, ice.GO_SUM)
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "contexts.home.tar.gz"), ".vim/plugged", kit.Dict(nfs.DIR_ROOT, kit.Env(cli.HOME)))
				m.Cmd("web.code.git.server", mdb.IMPORT)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Option(nfs.DIR_ROOT, ice.USR_PUBLISH)
			m.Cmdy(nfs.DIR, kit.Select("", arg, 0), nfs.DIR_WEB_FIELDS)
		}},
	}})
}

var _contexts = kit.Dict(
	ice.MISC, `# 下载命令 curl 或 wget
export ctx_dev={{.Option "httphost"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL $ctx_dev; source $ctx_temp app username {{.Option "user.name"}}
export ctx_dev={{.Option "httphost"}}{{.Option "ctx_env"}}; ctx_temp=$(mktemp); wget -O $ctx_temp -q $ctx_dev; source $ctx_temp app username {{.Option "user.name"}}
`,
	ice.CORE, `# 下载工具 curl 或 wget
ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL {{.Cmdx "spide" "shy" "client.url"}}; source $ctx_temp binary
ctx_temp=$(mktemp); wget -O $ctx_temp -q {{.Cmdx "spide" "shy" "client.url"}}; source $ctx_temp binary
`,
	ice.BASE, `# 下载源码
git clone {{.Option "remote"}}; cd {{.Option "pathname"}} && source etc/miss.sh
`,
)
