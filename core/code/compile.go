package code

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const COMPILE = "compile"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			COMPILE: {Name: COMPILE, Help: "编译", Value: kit.Data(
				kit.MDB_PATH, ice.USR_PUBLISH, cli.ENV, kit.Dict(
					"CGO_ENABLED", "0", "GOCACHE", os.Getenv("GOCACHE"),
					cli.HOME, os.Getenv(cli.HOME), cli.PATH, os.Getenv(cli.PATH),
					"GOPROXY", "https://goproxy.cn,direct", "GOPRIVATE", "github.com",
				), GO, []interface{}{GO, cli.BUILD},
			)},
		},
		Commands: map[string]*ice.Command{
			COMPILE: {Name: "compile arch=amd64,386,arm os=linux,darwin,windows src=src/main.go@key run:button", Help: "编译", Action: map[string]*ice.Action{
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(nfs.DIR, "src", "path,size,time", ice.Option{nfs.DIR_REG, `.*\.go$`})
					m.Sort(kit.MDB_PATH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Cmdy(nfs.DIR, m.Conf(COMPILE, kit.META_PATH), "time,size,path")
					return
				}

				main := ice.SRC_MAIN_GO
				goos := m.Conf(cli.RUNTIME, "host.GOOS")
				arch := m.Conf(cli.RUNTIME, "host.GOARCH")
				for _, k := range arg {
					switch k {
					case cli.LINUX, cli.DARWIN, cli.WINDOWS:
						goos = k
					case "amd64", "386", "arm":
						arch = k
					default:
						main = k
					}
				}
				_autogen_version(m.Spawn())

				// 编译目标
				file := path.Join(kit.Select("", m.Conf(COMPILE, kit.META_PATH), m.Option(cli.CMD_DIR) == ""),
					kit.Keys(kit.Select("ice", path.Base(strings.TrimSuffix(main, ".go")), main != ice.SRC_MAIN_GO), goos, arch))

				// 编译参数
				m.Optionv(cli.CMD_ENV, kit.Simple(m.Confv(COMPILE, kit.Keym(cli.ENV)), "GOARCH", arch, "GOOS", goos))
				if msg := m.Cmd(cli.SYSTEM, m.Confv(COMPILE, kit.Keym(GO)),
					"-o", file, main, ice.SRC_VERSION, ice.SRC_BINPACK); msg.Append(cli.CMD_CODE) != "0" {
					m.Copy(msg)
					return
				}

				m.Log_EXPORT(cli.SOURCE, main, cli.TARGET, file)
				m.Cmdy(nfs.DIR, file)
				m.EchoDownload(file)
				m.StatusTimeCount()
			}},
		},
	})
}
