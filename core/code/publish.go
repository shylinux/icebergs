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

	} else if s, e := os.Stat(file); m.Assert(e) && s.IsDir() {
		p := path.Base(file) + ".tar.gz"
		m.Cmd(nfs.TAR, mdb.IMPORT, p, file)
		defer func() { os.Remove(p) }()
		file = p // 打包目录
	}

	// 发布文件
	target := path.Join(m.Config(nfs.PATH), kit.Select(path.Base(file), arg, 0))
	m.Log_EXPORT(PUBLISH, target, cli.FROM, file)
	m.Cmd(nfs.LINK, target, file)
	return target
}
func _publish_list(m *ice.Message, arg ...string) {
	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_ROOT, m.Config(nfs.PATH))
	m.Option(nfs.DIR_REG, kit.Select("", arg, 0))
	m.Cmdy(nfs.DIR, nfs.PWD, kit.Select(nfs.DIR_WEB_FIELDS, arg, 1))
}
func _publish_bin_list(m *ice.Message, dir string) {
	p := m.Option(cli.CMD_DIR, dir)
	for _, ls := range strings.Split(strings.TrimSpace(m.Cmd(cli.SYSTEM, "bash", "-c", "ls |xargs file |grep executable").Append(cli.CMD_OUT)), ice.NL) {
		if file := strings.TrimSpace(strings.Split(ls, ":")[0]); file != "" {
			if s, e := os.Stat(path.Join(p, file)); e == nil {
				m.Push(mdb.TIME, s.ModTime())
				m.Push(nfs.SIZE, kit.FmtSize(s.Size()))
				m.Push(nfs.FILE, file)
				m.PushDownload(mdb.LINK, file, path.Join(p, file))
			}
		}
	}
	m.SortTimeR(mdb.TIME)
}

const PUBLISH = "publish"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		PUBLISH: {Name: PUBLISH, Help: "发布", Value: kit.Data(nfs.PATH, ice.USR_PUBLISH, ice.CONTEXTS, _contexts)},
	}, Commands: map[string]*ice.Command{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell export", Help: "发布", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.Config(nfs.PATH))
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.PrefixKey())
				m.Config(ice.CONTEXTS, _contexts)
			}},
			ice.VOLCANOS: {Name: "volcanos", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.EchoQRCode(m.Option(ice.MSG_USERWEB)) }()
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS) }()
				_publish_list(m, `.*\.(html|css|js)$`)
			}},
			ice.ICEBERGS: {Name: "icebergs", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS) }()
				m.Cmd(PUBLISH, mdb.CREATE, nfs.FILE, ice.BIN_ICE_BIN)
				_publish_bin_list(m, m.Config(nfs.PATH))
			}},
			ice.INTSHELL: {Name: "intshell", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS) }()
				_publish_list(m, ".*\\.(sh|vim|conf)$")
			}},
			ice.CONTEXTS: {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
				u := kit.ParseURL(tcp.ReplaceLocalhost(m, m.Option(ice.MSG_USERWEB)))
				m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, strings.Split(u.Host, ice.DF)[0],
					kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ice.DF), 1)))

				if len(arg) == 0 {
					arg = append(arg, ice.CORE, ice.BASE)
				}
				for _, k := range arg {
					if buf, err := kit.Render(m.Config(kit.Keys(ice.CONTEXTS, k)), m); m.Assert(err) {
						m.EchoScript(strings.TrimSpace(string(buf)))
					}
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, kit.Select(nfs.PWD, arg, 1))
				m.ProcessAgain()
			}},
			mdb.CREATE: {Name: "create file", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_publish_file(m, m.Option(nfs.FILE))
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				p := m.Option(cli.CMD_DIR, m.Config(nfs.PATH))
				os.Remove(path.Join(p, m.Option(nfs.PATH)))
			}},
			mdb.EXPORT: {Name: "export", Help: "工具链", Hand: func(m *ice.Message, arg ...string) {
				var list = []string{ice.USR_LOCAL_LIB}
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
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "vim.tar.gz"), ".vim/plugged", kit.Dict(nfs.DIR_ROOT, kit.Env(cli.HOME)))
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "contexts.bin.tar.gz"), list)
				m.Cmd(PUBLISH, mdb.CREATE, ice.ETC_PATH)

				m.Cmd(PUBLISH, mdb.CREATE, ice.MAKEFILE)
				m.Cmd(PUBLISH, mdb.CREATE, ice.ETC_MISS_SH)
				m.Cmd(PUBLISH, mdb.CREATE, ice.SRC_MAIN_GO)
				m.Cmd(PUBLISH, mdb.CREATE, ice.GO_MOD)
				m.Cmd(PUBLISH, mdb.CREATE, ice.GO_SUM)

				m.Cmd("web.code.git.server", mdb.IMPORT)
				m.ToastSuccess()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(nfs.DIR_ROOT, m.Config(nfs.PATH))
			m.Cmdy(nfs.DIR, kit.Select("", arg, 0), nfs.DIR_WEB_FIELDS)
		}},
	}})
}

var _contexts = kit.Dict(
	"misc", `# 终端环境
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp dev
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); wget -O $ctx_temp $ctx_dev; source $ctx_temp dev
`,
	"core", `# 定制版
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL $ctx_dev; source $ctx_temp app
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); wget -O $ctx_temp $ctx_dev; source $ctx_temp app
`,
	"base", `# 官方版
ctx_temp=$(mktemp); curl -o $ctx_temp -fsSL https://shylinux.com; source $ctx_temp binary
ctx_temp=$(mktemp); wget -O $ctx_temp https://shylinux.com; source $ctx_temp binary
`,
)
