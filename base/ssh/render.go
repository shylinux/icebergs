package ssh

import (
	"fmt"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func Render(msg *ice.Message, cmd string, arg ...ice.Any) (res string) {
	switch args := kit.Simple(arg...); cmd {
	case ice.RENDER_RESULT:
		if len(args) > 0 {
			msg.Resultv(args)
		}
		res = msg.Result()
	case ice.RENDER_VOID:
		return res
	default:
		if res = msg.Result(); res == "" {
			res = msg.TableEcho().Result()
		}
	}
	if fmt.Fprint(msg.O, res); !strings.HasSuffix(res, ice.NL) {
		fmt.Fprint(msg.O, ice.NL)
	}
	return res
}
