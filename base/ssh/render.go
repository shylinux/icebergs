package ssh

import (
	"fmt"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	kit "shylinux.com/x/toolkits"
)

func Render(msg *ice.Message, cmd string, arg ...ice.Any) (res string) {
	switch args := kit.Simple(arg...); cmd {
	case ice.RENDER_RESULT:
		kit.If(len(args) > 0, func() { msg.Resultv(args) })
		res = msg.Result()
	case ice.RENDER_VOID:
		return res
	default:
		if res = msg.Result(); res == "" {
			res = msg.TableEcho().Result()
		}
	}
	if fmt.Fprint(msg.O, res); !strings.HasSuffix(res, lex.NL) {
		fmt.Fprint(msg.O, lex.NL)
	}
	return res
}
