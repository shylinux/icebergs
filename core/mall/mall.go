package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const MALL = "mall"

var Index = &ice.Context{Name: MALL, Help: "贸易中心"}

func init() { web.Index.Register(Index, nil, ASSET, SALARY) }

func Prefix(arg ...ice.Any) string { return web.Prefix(MALL, kit.Keys(arg...)) }
