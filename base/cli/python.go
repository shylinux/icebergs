package cli

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"fmt"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PYTHON: {Name: "python", Help: "脚本命令", Value: kit.Data(
				PYTHON, "python", "pip", "pip",
				"qrcode", `import pyqrcode; print(pyqrcode.create('%s').terminal(module_color='%s', quiet_zone=1))`,
			)},
		},
		Commands: map[string]*ice.Command{
			PYTHON: {Name: "python cmd arg...", Help: "脚本命令", Action: map[string]*ice.Action{
				"install": {Name: "install arg...", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SYSTEM, m.Conf(PYTHON, "meta.pip"), "install", arg)
				}},
				"qrcode": {Name: "qrcode text color", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
					prefix := []string{SYSTEM, m.Conf(PYTHON, kit.Keys(kit.MDB_META, PYTHON))}
					m.Cmdy(prefix, "-c", fmt.Sprintf(m.Conf(PYTHON, "meta.qrcode"),
						kit.Select("hello world", arg, 0), kit.Select("blue", arg, 1)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				prefix := []string{SYSTEM, m.Conf(PYTHON, kit.Keys(kit.MDB_META, PYTHON))}
				m.Cmdy(prefix, arg)
			}},
		},
	}, nil)
}
