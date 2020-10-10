package code

import (
	"strings"

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
			COMPILE: {Name: COMPILE, Help: "编译", Value: kit.Data(
				kit.MDB_PATH, "usr/publish", "env", kit.Dict(
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
			COMPILE: {Name: "compile os=linux,darwin,windows arch=amd64,386,arm src=src/main.go 执行:button", Help: "编译", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					// 目录列表
					m.Cmdy(nfs.DIR, m.Conf(COMPILE, kit.META_PATH), "time size path")
					return
				}

				main := "src/main.go"
				goos := m.Conf(cli.RUNTIME, "host.GOOS")
				arch := m.Conf(cli.RUNTIME, "host.GOARCH")
				for _, k := range arg {
					switch k {
					case "linux", "darwin", "windows":
						goos = k
					case "amd64", "386", "arm":
						arch = k
					default:
						main = k
					}
				}

				// 编译目标
				file := path.Join(m.Conf(cmd, "meta.path"), kit.Keys(kit.Select("ice", path.Base(strings.TrimSuffix(main, ".go")), main != "src/main.go"), goos, arch))

				// 编译参数
				m.Optionv(cli.CMD_ENV, kit.Simple(m.Confv(COMPILE, "meta.env"), "GOARCH", arch, "GOOS", goos))
				if msg := m.Cmd(cli.SYSTEM, m.Confv(COMPILE, "meta.go"), file, main); msg.Append(cli.CMD_CODE) != "0" {
					m.Copy(msg)
				} else {
					m.Log_EXPORT("source", main, "target", file)
					m.PushRender(kit.MDB_LINK, "download", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/publish/"+path.Base(file)))
					m.Echo(file)
				}
			}},
		},
	}, nil)
}
