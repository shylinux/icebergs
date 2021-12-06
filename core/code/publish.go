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
				m.Push(kit.MDB_TIME, s.ModTime())
				m.Push(kit.MDB_SIZE, kit.FmtSize(s.Size()))
				m.Push(kit.MDB_FILE, file)
				m.PushDownload(kit.MDB_LINK, file, path.Join(p, file))
			}
		}
	}
	m.SortTimeR(kit.MDB_TIME)
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
	target := path.Join(m.Config(kit.MDB_PATH), kit.Select(path.Base(file), arg, 0))
	m.Log_EXPORT(PUBLISH, target, kit.MDB_FROM, file)
	m.Cmd(nfs.LINK, target, file)
	return target
}
func _publish_list(m *ice.Message, arg ...string) {
	m.Option(nfs.DIR_DEEP, ice.TRUE)
	m.Option(nfs.DIR_REG, kit.Select("", arg, 0))
	m.Option(nfs.DIR_ROOT, m.Config(kit.MDB_PATH))
	m.Cmdy(nfs.DIR, ice.PWD, kit.Select("time,size,line,path,link", arg, 1))
}

const PUBLISH = "publish"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		PUBLISH: {Name: PUBLISH, Help: "发布", Value: kit.Data(
			kit.MDB_PATH, ice.USR_PUBLISH, ice.CONTEXTS, _contexts,
			SH, `#!/bin/bash
echo "hello world"
`,
			JS, `Volcanos("onengine", {})
`,
		)},
	}, Commands: map[string]*ice.Command{
		PUBLISH: {Name: "publish path auto create volcanos icebergs intshell export", Help: "发布", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.Config(kit.MDB_PATH))
				m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, m.PrefixKey())
				m.Config(ice.CONTEXTS, _contexts)
			}},
			ice.VOLCANOS: {Name: "volcanos", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.EchoQRCode(m.Option(ice.MSG_USERWEB)) }()
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.CORE) }()
				m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.ETC_MISS_SH)
				m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.GO_MOD)

				m.Cmd(nfs.DEFS, path.Join(m.Config(kit.MDB_PATH), ice.ORDER_JS), m.Config(JS))
				m.Cmd(nfs.DEFS, path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, nfs.PATH)), PAGE_CACHE_JS), "")
				m.Cmd(nfs.DEFS, path.Join(m.Conf(web.SERVE, kit.Keym(ice.VOLCANOS, nfs.PATH)), PAGE_CACHE_CSS), "")
				_publish_list(m, `.*\.(html|css|js)$`)
			}},
			ice.ICEBERGS: {Name: "icebergs", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.BASE) }()
				m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE_BIN)
				m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE_SH)
				_bin_list(m, m.Config(kit.MDB_PATH))
			}},
			ice.INTSHELL: {Name: "intshell", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
				defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, ice.MISC) }()
				m.Cmd(nfs.DEFS, path.Join(m.Config(kit.MDB_PATH), ice.ORDER_SH), m.Config(SH))
				_publish_list(m, ".*\\.(sh|vim|conf)$")
			}},
			ice.CONTEXTS: {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
				u := kit.ParseURL(tcp.ReplaceLocalhost(m, m.Option(ice.MSG_USERWEB)))
				m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, strings.Split(u.Host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ":"), 1)))

				if len(arg) == 0 {
					arg = append(arg, "misc", "core", "base", "binary", "source", "project")
				}
				for _, k := range arg {
					if buf, err := kit.Render(m.Config(kit.Keys(ice.CONTEXTS, k)), m); m.Assert(err) {
						m.EchoScript(strings.TrimSpace(string(buf)))
					}
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.DIR, kit.Select(ice.PWD, arg, 1))
				m.ProcessAgain()
			}},
			mdb.CREATE: {Name: "create file", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_publish_file(m, m.Option(kit.MDB_FILE))
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				p := m.Option(cli.CMD_DIR, m.Config(kit.MDB_PATH))
				os.Remove(path.Join(p, m.Option(kit.MDB_PATH)))
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
			m.Option(nfs.DIR_ROOT, m.Config(kit.MDB_PATH))
			m.Cmdy(nfs.DIR, kit.Select("", arg, 0), "time,size,path,action,link")
		}},
	}})
}

var _contexts = kit.Dict(
	"project", `# 创建项目
ctx_temp=$(mktemp); curl -fsSL https://shylinux.com -o $ctx_temp; source $ctx_temp project
`,
	"source", `# 下载源码
ctx_temp=$(mktemp); curl -fsSL https://shylinux.com -o $ctx_temp; source $ctx_temp source
`,
	"binary", `# 安装应用
ctx_temp=$(mktemp); curl -fsSL https://shylinux.com -o $ctx_temp; source $ctx_temp binary
`,
	"base", `# 生产环境
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp app
`,
	"core", `# 开发环境
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp dev
`,
	"misc", `# 终端环境
export ctx_dev={{.Option "httphost"}}; ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp
`,
	"tool", `# 群组环境
export ctx_dev={{.Option "httphost"}} ctx_share={{.Option "share"}} ctx_river={{.Option "sess.river"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp app
`,
)
