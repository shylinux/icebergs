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
			COMPILE: {Name: "compile os=linux,darwin,windows arch=amd64,386,arm src=src/main.go@key 执行:button", Help: "编译", Action: map[string]*ice.Action{
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
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
				args := []string{"-ldflags"}
				list := []string{
					fmt.Sprintf(`-X main.Time="%s"`, m.Time()),
					fmt.Sprintf(`-X main.Version="%s"`, m.Cmdx(cli.SYSTEM, "git", "describe", "--tags")),
					fmt.Sprintf(`-X main.HostName="%s"`, m.Conf(cli.RUNTIME, "boot.hostname")),
					fmt.Sprintf(`-X main.UserName="%s"`, m.Conf(cli.RUNTIME, "boot.username")),
				}

				// 编译参数
				m.Optionv(cli.CMD_ENV, kit.Simple(m.Confv(COMPILE, "meta.env"), "GOARCH", arch, "GOOS", goos))
				if msg := m.Cmd(cli.SYSTEM, m.Confv(COMPILE, "meta.go"), args, "'"+strings.Join(list, " ")+"'", "-o", file, main); msg.Append(cli.CMD_CODE) != "0" {
					m.Copy(msg)
				} else {
					m.Log_EXPORT("source", main, "target", file)
					m.PushRender(kit.MDB_LINK, "download", kit.MergeURL2(m.Option(ice.MSG_USERWEB), "/publish/"+path.Base(file)))
					m.Echo(file)
				}
			}},
		},
	})
}
