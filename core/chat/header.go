package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"fmt"
)

const (
	CHECK = "check"
	LOGIN = "login"
	TITLE = "title"
)
const HEADER = "header"
const _pack = `
<!DOCTYPE html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=0.7,user-scalable=no">
    <title>volcanos</title>
    <link rel="shortcut icon" type="image/ico" href="favicon.ico">
    <style type="text/css">%s</style>
    <style type="text/css">%s</style>
</head>
<body>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>Volcanos.meta.webpack = true</script>
</body>
`

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			HEADER: {Name: "header", Help: "标题栏", Value: kit.Dict(
				TITLE, "github.com/shylinux/contexts",
			)},
		},
		Commands: map[string]*ice.Command{
			"/" + HEADER: {Name: "/header", Help: "标题栏", Action: map[string]*ice.Action{
				"userrole": {Name: "userrole", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(aaa.UserRole(m, m.Option("who")))
				}},

				CHECK: {Name: "check", Help: "登录检查", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},
				LOGIN: {Name: "login", Help: "用户登录", Hand: func(m *ice.Message, arg ...string) {
					m.Echo(m.Option(ice.MSG_USERNAME))
				}},

				"pack": {Name: "pack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.webpack", "pack")

					if f, _, e := kit.Create("usr/publish/webpack/" + m.Option("name") + ".js"); m.Assert(e) {
						defer f.Close()

						f.WriteString(`Volcanos.meta.pack = ` + kit.Formats(kit.UnMarshal(m.Option("content"))))
					}

					m.Option(nfs.DIR_ROOT, "")
					if f, p, e := kit.Create("usr/publish/webpack/" + m.Option("name") + ".html"); m.Assert(e) {
						f.WriteString(fmt.Sprintf(_pack,
							m.Cmdx(nfs.CAT, "usr/volcanos/cache.css"),
							m.Cmdx(nfs.CAT, "usr/volcanos/index.css"),

							m.Cmdx(nfs.CAT, "usr/volcanos/proto.js"),
							m.Cmdx(nfs.CAT, "usr/volcanos/cache.js"),
							m.Cmdx(nfs.CAT, "usr/publish/webpack/"+m.Option("name")+".js"),
							m.Cmdx(nfs.CAT, "usr/volcanos/index.js"),
						))
						m.Echo(p)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Echo(m.Conf(HEADER, TITLE))
			}},
		},
	}, nil)
}
