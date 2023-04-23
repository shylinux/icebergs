package bash

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/web"
)

const CONFIGS = "configs"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(CONFIGS): {Hand: func(m *ice.Message, arg ...string) {
			if strings.Contains(m.Option(cli.RELEASE), cli.ALPINE) {
				m.Echo("sed -i 's/dl-cdn.alpinelinux.org/mirrors.tencent.com/g' /etc/apk/repositories && apk update").Echo(lex.NL)
				m.Echo("TZ=Asia/Shanghai; apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone").Echo(lex.NL)
			}
		}},
	})
}
