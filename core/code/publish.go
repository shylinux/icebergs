package code

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _publish_file(m *ice.Message, file string, arg ...string) string {
	if strings.HasSuffix(file, "ice.bin") {
		// 打包应用
		arg = append(arg, kit.Keys("ice", runtime.GOOS, runtime.GOARCH))
		if _, e := os.Stat(path.Join(m.Conf(PUBLISH, kit.META_PATH), kit.Select(path.Base(file), arg, 0))); e == nil {
			return ""
		}

	} else if s, e := os.Stat(file); m.Assert(e) && s.IsDir() {
		// 打包目录
		p := path.Base(file) + ".tar.gz"
		m.Cmd(cli.SYSTEM, "tar", "-zcf", p, file)
		defer func() { os.Remove(p) }()
		file = p
	}

	// 发布文件
	target := path.Join(m.Conf(PUBLISH, kit.META_PATH), kit.Select(path.Base(file), arg, 0))
	m.Log_EXPORT(PUBLISH, target, kit.MDB_FROM, file)
	m.Cmd(nfs.LINK, target, file)
	return target
}

const PUBLISH = "publish"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PUBLISH: {Name: PUBLISH, Help: "发布", Value: kit.Data(
				kit.MDB_PATH, "usr/publish", ice.CONTEXTS, _contexts,
				SH, `#!/bin/bash
echo "hello world"
`,
				JS, `Volcanos("onengine", {_init: function(can, sub) {
    can.base.Log("hello volcanos world")
}, river: {

}})
`,
			)},
		},
		Commands: map[string]*ice.Command{
			PUBLISH: {Name: "publish path auto dream volcanos icebergs intshell", Help: "发布", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create file", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_publish_file(m, m.Option(kit.MDB_FILE))
				}},
				ice.VOLCANOS: {Name: "volcanos", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
					defer func() { m.EchoQRCode(m.Option(ice.MSG_USERWEB)) }()
					defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, "miss") }()
					m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.ETC_MISS)
					m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.GO_MOD)
					m.Cmd(nfs.DEFS, path.Join(m.Conf(PUBLISH, kit.META_PATH), ice.ORDER_JS), m.Conf(PUBLISH, kit.Keym(JS)))
					m.Cmd(nfs.DEFS, "usr/volcanos/page/cache.css", "")
					m.Cmd(nfs.DEFS, "usr/volcanos/page/cache.js", "")

					m.Option(nfs.DIR_DEEP, true)
					m.Option(nfs.DIR_REG, `.*\.(html|css|js)$`)
					m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
					m.Cmdy(nfs.DIR, "./", "time,size,line,path,link")
				}},
				ice.ICEBERGS: {Name: "icebergs", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
					defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, "base") }()
					m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE_BIN)
					m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, ice.BIN_ICE)

					p := m.Option(cli.CMD_DIR, m.Conf(PUBLISH, kit.META_PATH))
					ls := strings.Split(strings.TrimSpace(m.Cmd(cli.SYSTEM, "bash", "-c", "ls |xargs file |grep executable").Append(cli.CMD_OUT)), "\n")
					for _, ls := range ls {
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
				}},
				ice.INTSHELL: {Name: "intshell", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
					defer func() { m.Cmdy(PUBLISH, ice.CONTEXTS, "tmux") }()
					m.Cmd(nfs.DEFS, path.Join(m.Conf(PUBLISH, kit.META_PATH), ice.ORDER_SH), m.Conf(PUBLISH, kit.Keym(SH)))

					m.Option(nfs.DIR_DEEP, true)
					m.Option(nfs.DIR_REG, ".*\\.(sh|vim|conf)$")
					m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
					m.Cmdy(nfs.DIR, "./", "time,size,line,path,link")
				}},
				ice.CONTEXTS: {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					host := u.Host
					if strings.Contains(host, "localhost") {
						host = strings.Replace(host, "localhost", m.Cmd(tcp.HOST).Append(tcp.IP), 1)
					}

					m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, strings.Split(host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(host, ":"), 1)))
					m.Option("hostport", fmt.Sprintf("%s:%s", strings.Split(host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(host, ":"), 1)))
					m.Option("hostname", strings.Split(host, ":")[0])

					m.Option("userhost", fmt.Sprintf("%s@%s", m.Option(ice.MSG_USERNAME), strings.Split(host, ":")[0]))
					m.Option("hostpath", kit.Path("./.ish/pluged"))

					if len(arg) == 0 {
						arg = append(arg, "base")
					}
					for _, k := range arg {
						if buf, err := kit.Render(m.Conf(PUBLISH, kit.Keym(ice.CONTEXTS, k)), m); m.Assert(err) {
							m.EchoScript(strings.TrimSpace(string(buf)))
						}
					}
				}},
				web.DREAM: {Name: "dream name=hi repos", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.DREAM, tcp.START, arg)
					m.Process(ice.PROCESS_OPEN, kit.MergeURL(m.Option(ice.MSG_USERWEB),
						kit.SSH_POD, kit.Keys(m.Option(ice.MSG_USERPOD), m.Option(kit.MDB_NAME))))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.DREAM, mdb.INPUTS, arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					p := m.Option(cli.CMD_DIR, m.Conf(PUBLISH, kit.META_PATH))
					os.Remove(path.Join(p, m.Option(kit.MDB_PATH)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
				m.Cmdy(nfs.DIR, kit.Select("", arg, 0), "time,size,path,link")
			}},
		},
	})
}

var _contexts = kit.Dict(
	"module", `# 创建模块
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp module
`,
	"source", `# 源码安装
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp source
`,
	"binary", `# 应用安装
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp binary
`,
	"tmux", `# 终端环境
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp
`,
	"base", `# 生产环境
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp app
`,
	"miss", `# 开发环境
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp dev
`,
	"tool", `# 群组环境
mkdir contexts; cd contexts
export ctx_log=/dev/stdout ctx_dev={{.Option "httphost"}} ctx_river={{.Option "sess.river"}} ctx_share={{.Option "share"}} ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp ice
`,
)
