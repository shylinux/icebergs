package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
)

const PUBLISH = "publish"

func _publish_file(m *ice.Message, file string, arg ...string) string {
	if s, e := os.Stat(file); m.Assert(e) && s.IsDir() {
		// 打包目录
		p := path.Base(file) + ".tar.gz"
		m.Cmd(cli.SYSTEM, "tar", "-zcf", p, file)
		defer func() { os.Remove(p) }()
		file = p
	}
	if strings.HasSuffix(file, "ice.bin") {
		arg = append(arg, kit.Keys("ice", runtime.GOOS, runtime.GOARCH))
	}

	// 发布文件
	target := path.Join(m.Conf(PUBLISH, kit.META_PATH), kit.Select(path.Base(file), arg, 0))
	m.Cmd(nfs.LINK, target, file)

	// 发布记录
	m.Log_EXPORT(PUBLISH, target, "from", file)
	return target
}
func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PUBLISH: {Name: PUBLISH, Help: "发布", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_PATH, "usr/publish", "contexts", _contexts,
			)},
		},
		Commands: map[string]*ice.Command{
			PUBLISH: {Name: "publish path auto create can ice ish", Help: "发布", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create file", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_publish_file(m, m.Option(kit.MDB_FILE))
				}},
				"can": {Name: "can", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
					m.Option(nfs.DIR_DEEP, true)
					m.Option(nfs.DIR_REG, `.*\.(html|css|js)$`)
					m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
					m.Cmdy(nfs.DIR, "./", "time,size,line,path,link")
					m.Cmdy(PUBLISH, "contexts", "miss")
				}},
				"ice": {Name: "ice", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
					defer func() { m.Cmdy(PUBLISH, "contexts", "base") }()
					m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, "bin/ice.sh")
					m.Cmd(PUBLISH, mdb.CREATE, kit.MDB_FILE, "bin/ice.bin")

					p := m.Option(cli.CMD_DIR, m.Conf(PUBLISH, kit.META_PATH))
					ls := strings.Split(strings.TrimSpace(m.Cmd(cli.SYSTEM, "bash", "-c", "ls |xargs file |grep executable").Append(cli.CMD_OUT)), "\n")
					for _, ls := range ls {
						if file := strings.TrimSpace(strings.Split(ls, ":")[0]); file != "" {
							if s, e := os.Stat(path.Join(p, file)); e == nil {
								m.Push(kit.MDB_TIME, s.ModTime())
								m.Push(kit.MDB_SIZE, kit.FmtSize(s.Size()))
								m.Push(kit.MDB_FILE, file)
								m.PushDownload(file, path.Join(p, file))
							}
						}
					}
					m.SortTimeR(kit.MDB_TIME)
				}},
				"ish": {Name: "ish", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
					m.Option(nfs.DIR_DEEP, true)
					m.Option(nfs.DIR_REG, ".*\\.(sh|vim|conf)$")
					m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
					m.Cmdy(nfs.DIR, "./", "time,size,line,path,link")
					m.Cmdy(PUBLISH, "contexts", "tmux")
				}},
				"contexts": {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, strings.Split(u.Host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ":"), 1)))
					m.Option("hostport", fmt.Sprintf("%s:%s", strings.Split(u.Host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ":"), 1)))
					m.Option("hostname", strings.Split(u.Host, ":")[0])

					m.Option("userhost", fmt.Sprintf("%s@%s", m.Option(ice.MSG_USERNAME), strings.Split(u.Host, ":")[0]))
					m.Option("hostpath", kit.Path("./.ish/pluged"))

					if len(arg) == 0 {
						arg = append(arg, "base")
					}
					for _, k := range arg {
						if buf, err := kit.Render(m.Conf(PUBLISH, kit.Keym("contexts", k)), m); m.Assert(err) {
							m.EchoScript(string(buf))
						}
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					p := m.Option(cli.CMD_DIR, m.Conf(PUBLISH, kit.META_PATH))
					os.Remove(path.Join(p, m.Option(kit.MDB_PATH)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Conf(PUBLISH, kit.META_PATH))
				m.Cmdy(nfs.DIR, kit.Select("", arg, 0), "time size path link")
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}

var _contexts = kit.Dict(
	"tmux", `
# 终端环境
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp
`,
	"base", `
# 生产环境
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp ice
`,
	"miss", `
# 开发环境
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp dev
`,
	"tool", `
# 群组环境
mkdir contexts; cd contexts
export ctx_log=/dev/stdout ctx_dev={{.Option "httphost"}} ctx_river={{.Option "sess.river"}} ctx_share={{.Option "share"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp ice
`,
)
