package cli

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			QRCODE: {Name: "qrcode", Help: "二维码", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			QRCODE: {Name: "qrcode", Help: "二维码", Action: map[string]*ice.Action{}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			}},
		},
	})
}
