package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
)

const COMPILE = "compile"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			COMPILE: {Name: "compile", Help: "编译", Value: kit.Data(
				"path", "usr/publish", "env", kit.Dict(
					"PATH", os.Getenv("PATH"),
					"HOME", os.Getenv("HOME"),
					"GOCACHE", os.Getenv("GOCACHE"),
					"GOPROXY", "https://goproxy.cn,direct",
					"GOPRIVATE", "github.com",
					"CGO_ENABLED", "0",
				), "go", []interface{}{"go", "build", "-o"},
			)},
		},
		Commands: map[string]*ice.Command{
			COMPILE: {Name: "compile [os [arch [main]]]", Help: "编译", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 目录列表
					m.Cmdy(nfs.DIR, m.Conf(cmd, "meta.path"), "time size path")
					return
				}

				// 编译目标
				main := kit.Select("src/main.go", arg, 2)
				arch := kit.Select(m.Conf(cli.RUNTIME, "host.GOARCH"), arg, 1)
				goos := kit.Select(m.Conf(cli.RUNTIME, "host.GOOS"), arg, 0)
				file := ""
				if m.Option(cli.CMD_DIR) == "" {
					file = path.Join(m.Conf(cmd, "meta.path"), kit.Keys(kit.Select("ice", m.Option("name")), goos, arch))
				} else {
					file = kit.Keys(kit.Select("ice", m.Option("name")), goos, arch)
				}
				if goos == "windows" {
					file += ".exe"
				}

				// 编译参数
				m.Optionv(cli.CMD_ENV, kit.Simple(m.Confv(COMPILE, "meta.env"), "GOARCH", arch, "GOOS", goos))
				if msg := m.Cmd(cli.SYSTEM, m.Confv(COMPILE, "meta.go"), file, main); msg.Append(cli.CMD_CODE) != "0" {
					m.Copy(msg)
					return
				}

				// 编译记录
				// m.Cmdy(web.STORY, web.CATCH, "bin", file)
				m.Log_EXPORT("source", main, "target", file)
			}},
		},
	}, nil)
}
