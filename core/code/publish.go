package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
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

	// 发布文件
	target := path.Join(m.Conf(PUBLISH, kit.META_PATH), kit.Select(path.Base(file), arg, 0))
	m.Cmd(nfs.LINK, target, file)

	// 发布记录
	// m.Cmdy(web.STORY, web.CATCH, "bin", target)
	m.Log_EXPORT(PUBLISH, target, "from", file)
	return target
}
func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PUBLISH: {Name: PUBLISH, Help: "发布", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_PATH, "usr/publish",
				"contexts", _contexts,
			)},
		},
		Commands: map[string]*ice.Command{
			PUBLISH: {Name: "publish path=auto auto 火山架 冰山架 神农架", Help: "发布", Action: map[string]*ice.Action{
				"contexts": {Name: "contexts", Help: "环境", Hand: func(m *ice.Message, arg ...string) {
					u := kit.ParseURL(m.Option(ice.MSG_USERWEB))
					m.Option("httphost", fmt.Sprintf("%s://%s:%s", u.Scheme, strings.Split(u.Host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ":"), 1)))
					m.Option("hostport", fmt.Sprintf("%s:%s", strings.Split(u.Host, ":")[0], kit.Select(kit.Select("80", "443", u.Scheme == "https"), strings.Split(u.Host, ":"), 1)))
					m.Option("hostname", strings.Split(u.Host, ":")[0])

					m.Option("userhost", fmt.Sprintf("%s@%s", m.Option(ice.MSG_USERNAME), strings.Split(u.Host, ":")[0]))
					m.Option("hostpath", kit.Path("./.ish/pluged"))

					if buf, err := kit.Render(m.Conf(PUBLISH, kit.Keys("meta.contexts", kit.Select("base", arg, 0))), m); m.Assert(err) {
						m.Cmdy("web.wiki.spark", "shell", string(buf))
					}
				}},
				"ish": {Name: "ish", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
					_publish_file(m, "etc/conf/tmux.conf")
					m.Option(nfs.DIR_REG, ".*\\.(sh|vim|conf)")
					m.Cmdy(nfs.DIR, m.Conf(PUBLISH, kit.META_PATH), "time size line path link")
				}},
				"ice": {Name: "ice", Help: "冰山架", Hand: func(m *ice.Message, arg ...string) {
					_publish_file(m, "bin/ice.bin", fmt.Sprintf("ice.%s.%s", runtime.GOOS, runtime.GOARCH))
					_publish_file(m, "bin/ice.sh")
					m.Option(nfs.DIR_REG, "ice.*")
					m.Cmdy(nfs.DIR, m.Conf(PUBLISH, kit.META_PATH), "time size path link")
				}},
				"can": {Name: "can", Help: "火山架", Hand: func(m *ice.Message, arg ...string) {
					m.Option(nfs.DIR_DEEP, true)
					m.Option(nfs.DIR_REG, ".*\\.(js|css|html)")
					m.Cmdy(nfs.DIR, m.Conf(PUBLISH, kit.META_PATH), "time size line path link")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Conf(cmd, kit.META_PATH))
				m.Cmdy(nfs.DIR, kit.Select("", arg, 0), "time size path")
			}},
		},
	}, nil)
}

var _contexts = kit.Dict(
	"tmux", `
# 终端环境
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp
`,
	"base", `
# 生产环境
mkdir contexts; cd contexts
export ctx_log=/dev/stdout ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp ice
`,
	"miss", `
# 开发环境
mkdir contexts; cd contexts
export ctx_dev={{.Option "httphost"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp dev
`,
	"tool", `
# 生产环境
mkdir contexts; cd contexts
export ctx_dev={{.Option "httphost"}} ctx_river={{.Option "sess.river"}} ctx_share={{.Option "share"}} ctx_temp=$(mktemp); curl -sL $ctx_dev >$ctx_temp; source $ctx_temp ice
`,
)

/*
yum install -y make git vim go
mkdir ~/.ssh &>/dev/null; touch ~/.ssh/config; [ -z "$(cat ~/.ssh/config|grep 'HOST {{.Option "hostname"}}')" ] && echo -e "HOST {{.Option "hostname"}}\n    Port 9030" >> ~/.ssh/config
export ISH_CONF_HUB_PROXY={{.Option "userhost"}}:.ish/pluged/
git clone $ISH_CONF_HUB_PROXY/github.com/shylinux/contexts && cd contexts
source etc/miss.sh

touch ~/.gitconfig; [ -z "$(cat ~/.gitconfig|grep '\[url \"{{.Option "userhost"}}')" ] && echo -e "[url \"{{.Option "userhost"}}:ish/pluged/\"]\n    insteadOf=\"https://github.com/\"\n" >> ~/.gitconfig
git clone https://github.com/shylinux/contexts && cd contexts
source etc/miss.sh
*/
