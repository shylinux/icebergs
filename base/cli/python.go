package cli

import (
	"path"
	"runtime"
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const PYTHON = "python"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PYTHON: {Name: "python", Help: "脚本命令", Value: kit.Data(
				"windows", "http://mirrors.sohu.com/python/3.5.2/Python-3.5.2.tar.xz",
				"darwin", "http://mirrors.sohu.com/python/3.5.2/Python-3.5.2.tar.xz",
				"linux", "http://mirrors.sohu.com/python/3.5.2/Python-3.5.2.tar.xz",

				"qrcode", `import pyqrcode; print(pyqrcode.create('%s').terminal(module_color='%s', quiet_zone=1))`,
				PYTHON, "python", "pip", "pip",
			)},
		},
		Commands: map[string]*ice.Command{
			PYTHON: {Name: "python 编译:button 下载:button", Help: "脚本命令", Action: map[string]*ice.Action{
				"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.install", "download", m.Conf(PYTHON, kit.Keys(kit.MDB_META, runtime.GOOS)))
				}},
				"compile": {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
					name := path.Base(strings.TrimSuffix(strings.TrimSuffix(m.Conf(PYTHON, kit.Keys(kit.MDB_META, runtime.GOOS)), ".tar.xz"), "zip"))
					m.Option(CMD_DIR, path.Join(m.Conf("web.code.install", kit.META_PATH), name))
					m.Cmdy(SYSTEM, "./configure", "--prefix="+kit.Path("usr/local"))
					m.Cmdy(SYSTEM, "make", "-j8")
					m.Cmdy(SYSTEM, "make", "install")
				}},

				"install": {Name: "install arg...", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SYSTEM, m.Conf(PYTHON, "meta.pip"), "install", arg)
				}},
				"qrcode": {Name: "qrcode text color", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
					prefix := []string{SYSTEM, m.Conf(PYTHON, kit.Keys(kit.MDB_META, PYTHON))}
					m.Cmdy(prefix, "-c", kit.Format(m.Conf(PYTHON, "meta.qrcode"),
						kit.Select("hello world", arg, 0), kit.Select("blue", arg, 1)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				prefix := []string{SYSTEM, m.Conf(PYTHON, kit.Keys(kit.MDB_META, PYTHON))}
				m.Cmdy(prefix, arg)
			}},
		},
	}, nil)
}
