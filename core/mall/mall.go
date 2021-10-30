package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
)

const MALL = "mall"

var Index = &ice.Context{Name: MALL, Help: "贸易中心"}

func init() { web.Index.Register(Index, nil, ASSET, SALARY) }
