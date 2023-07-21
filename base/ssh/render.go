package ssh

import (
	"fmt"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	kit "shylinux.com/x/toolkits"
)

func Render(m *ice.Message, cmd string, arg ...ice.Any) (res string) {
	switch args := kit.Simple(arg...); cmd {
	case ice.RENDER_RESULT:
		kit.If(len(args) > 0, func() { m.Resultv(args) })
		res = m.Result()
	case ice.RENDER_VOID:
		return res
	default:
		if res = m.Result(); res == "" {
			res = m.TableEchoWithStatus().Result()
		}
	}
	if fmt.Fprint(m.O, res); !strings.HasSuffix(res, lex.NL) {
		fmt.Fprint(m.O, lex.NL)
	}
	return res
}
