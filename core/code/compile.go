package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

const COMPILE = "compile"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			COMPILE: {Name: COMPILE, Help: "编译", Value: kit.Data(
				kit.MDB_PATH, "usr/publish", kit.SSH_ENV, kit.Dict(
					"CGO_ENABLED", "0", "GOCACHE", os.Getenv("GOCACHE"),
					"HOME", os.Getenv("HOME"), "PATH", os.Getenv("PATH"),
					"GOPROXY", "https://goproxy.cn,direct", "GOPRIVATE", "github.com",
				), "go", []interface{}{"go", "build"},
			)},
		},
		Commands: map[string]*ice.Command{
			COMPILE: {Name: "compile arch=amd64,386,arm os=linux,darwin,windows src=src/main.go@key 执行:button", Help: "编译", Action: map[string]*ice.Action{
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Option(nfs.DIR_REG, `.*\.go$`)
					m.Cmdy(nfs.DIR, "src", "path,size,time")
					m.Sort(kit.MDB_PATH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Cmdy(nfs.DIR, m.Conf(COMPILE, kit.META_PATH), "time,size,path")
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
				file := path.Join(kit.Select("", m.Conf(cmd, kit.META_PATH), m.Option(cli.CMD_DIR) == ""), kit.Keys(kit.Select("ice", path.Base(strings.TrimSuffix(main, ".go")), main != "src/main.go"), goos, arch))

				// 编译参数
				m.Optionv(cli.CMD_ENV, kit.Simple(m.Confv(COMPILE, "meta.env"), "GOARCH", arch, "GOOS", goos))
				if msg := m.Cmd(cli.SYSTEM, m.Confv(COMPILE, "meta.go"), "-o", file, main, "src/version.go", "src/binpack.go"); msg.Append(cli.CMD_CODE) != "0" {
					m.Copy(msg)
					return
				}

				m.Log_EXPORT("source", main, "target", file)
				m.Push(kit.MDB_TIME, m.Time())
				m.PushDownload(kit.MDB_LINK, file)
				m.Echo(file)
			}},
		},
	})
}
