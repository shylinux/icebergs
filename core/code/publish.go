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
			)},
		},
		Commands: map[string]*ice.Command{
			PUBLISH: {Name: "publish path=auto auto 火山架 冰山架 神农架", Help: "发布", Action: map[string]*ice.Action{
				"ish": {Name: "ish", Help: "神农架", Hand: func(m *ice.Message, arg ...string) {
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
