package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
)

const PUBLISH = "publish"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PUBLISH: {Name: "publish", Help: "发布", Value: kit.Data("path", "usr/publish")},
		},
		Commands: map[string]*ice.Command{
			PUBLISH: {Name: "publish [source]", Help: "发布", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 目录列表
					m.Cmdy(nfs.DIR, m.Conf(cmd, "meta.path"), "time size path")
					return
				}

				if s, e := os.Stat(arg[0]); m.Assert(e) && s.IsDir() {
					// 打包目录
					p := path.Base(arg[0]) + ".tar.gz"
					m.Cmd(cli.SYSTEM, "tar", "-zcf", p, arg[0])
					defer func() { os.Remove(p) }()
					arg[0] = p
				}

				// 发布文件
				target := path.Join(m.Conf(cmd, "meta.path"), path.Base(arg[0]))
				m.Cmd(nfs.LINK, target, arg[0])

				// 发布记录
				m.Cmdy(web.STORY, web.CATCH, "bin", target)
				m.Log_EXPORT("source", arg[0], "target", target)
			}},
		},
	}, nil)
}
