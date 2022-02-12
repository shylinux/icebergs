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

func _bin_list(m *ice.Message, dir string) {
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

func _publish_file(m *ice.Message, file string, arg ...string) string {
	if strings.HasSuffix(file, "ice.bin") { // 打包应用
		arg = kit.Simple(kit.Keys(ice.ICE, runtime.GOOS, runtime.GOARCH))

	} else if s, e := os.Stat(file); m.Assert(e) && s.IsDir() {
		p := path.Base(file) + ".tar.gz"
		m.Cmd(nfs.TAR, p, file)
		defer func() { os.Remove(p) }()
		file = p // 打包目录
	}

	// 发布文件
	target := path.Join(m.Config(nfs.PATH), kit.Select(path.Base(file), arg, 0))
	m.Log_EXPORT(PUBLISH, target, "from", file)
	m.Cmd(nfs.LINK, target, file)
	return target
}
func _publish_list(m *ice.Message, arg ...string) {
	m.Option(nfs.DIR_DEEP, ice.TRUE)
	m.Option(nfs.DIR_REG, kit.Select("", arg, 0))
	m.Option(nfs.DIR_ROOT, m.Config(nfs.PATH))
	m.Cmdy(nfs.DIR, nfs.PWD, kit.Select("time,size,line,path,link", arg, 1))
}

const PUBLISH = "publish"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		PUBLISH: {Name: PUBLISH, Help: "发布", Value: kit.Data(
			nfs.PATH, ice.USR_PUBLISH, ice.CONTEXTS, _contexts,
			SH, `#! /bin/sh
echo "hello world"
`,
			JS, `Volcanos("onengine", {})
`,
		)},
	}, Commands: map[string]*ice.Command{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell export", Help: "发布", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.Config(nfs.PATH))
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.PrefixKey())
				m.Config(ice.CONTEXTS, _contexts)
			}},
			ice.VOLCANOS: {Name: "volcanos", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.EchoQRCode(m.Option(ice.MSG_USERWEB)) }()
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.CORE) }()
				m.Cmd(PUBLISH, mdb.CREATE, nfs.FILE, ice.ETC_MISS_SH)
				m.Cmd(PUBLISH, mdb.CREATE, nfs.FILE, ice.GO_MOD)

				m.Cmd(nfs.DEFS, path.Join(m.Config(nfs.PATH), ice.ORDER_JS), m.Config(JS))
				m.Cmd(nfs.DEFS, path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, nfs.PATH)), PAGE_CACHE_JS), "")
				m.Cmd(nfs.DEFS, path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, nfs.PATH)), PAGE_CACHE_CSS), "")
				_publish_list(m, `.*\.(html|css|js)$`)
			}},
			ice.ICEBERGS: {Name: "icebergs", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.BASE) }()
				m.Cmd(PUBLISH, mdb.CREATE, nfs.FILE, ice.BIN_ICE_BIN)
				m.Cmd(PUBLISH, mdb.CREATE, nfs.FILE, ice.BIN_ICE_SH)
				_bin_list(m, m.Config(nfs.PATH))
			}},
			ice.INTSHELL: {Name: "intshell", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.MISC) }()
				m.Cmd(nfs.DEFS, path.Join(m.Config(nfs.PATH), ice.ORDER_SH), m.Config(SH))
				_publish_list(m, ".*\\.(sh|vim|conf)$")
			}},
			ice.CONTEXTS: {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
				u := kit.ParseURL(tcp.ReplaceLocalhost(m, m.Option(ice.MSG_USERWEB)))
				m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, strings.Split(u.Host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ":"), 1)))

				if len(arg) == 0 {
					arg = append(arg, "core", "binary")
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
				var list = []string{}
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
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "vim.tar.gz"), ".vim/plugged", kit.Dict(nfs.DIR_ROOT, os.Getenv(cli.HOME)))
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "contexts.lib.tar.gz"), ice.USR_LOCAL_LIB)
				m.Cmd(nfs.TAR, kit.Path(ice.USR_PUBLISH, "contexts.bin.tar.gz"), list)
				m.Cmd(PUBLISH, mdb.CREATE, ice.ETC_PATH)

				m.Cmd(PUBLISH, mdb.CREATE, ice.ETC_MISS_SH)
				m.Cmd(PUBLISH, mdb.CREATE, ice.GO_MOD)
				m.Cmd(PUBLISH, mdb.CREATE, ice.GO_SUM)

				m.Cmd("web.code.git.server", mdb.IMPORT)
				m.ToastSuccess()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(nfs.DIR_ROOT, m.Config(nfs.PATH))
			m.Cmdy(nfs.DIR, kit.Select("", arg, 0), "time,size,path,action,link")
		}},
	}})
}

var _contexts = kit.Dict(
	"binary", `# 官方启动
ctx_temp=$(mktemp); curl -fsSL https://shylinux.com -o $ctx_temp; source $ctx_temp binary
`,
	"core", `# 脚本启动
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp app
`,

	"source", `# 下载源码
ctx_temp=$(mktemp); curl -fsSL https://shylinux.com -o $ctx_temp; source $ctx_temp source
`,
	"project", `# 创建项目
ctx_temp=$(mktemp); curl -fsSL https://shylinux.com -o $ctx_temp; source $ctx_temp project
`,
	"base", `# 开发环境
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp dev
`,
	"misc", `# 终端环境
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp
`,
)
