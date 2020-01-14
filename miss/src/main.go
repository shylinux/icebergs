package main

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	_ "github.com/shylinux/icebergs/core"
	_ "github.com/shylinux/icebergs/misc"

	_ "github.com/shylinux/icebergs/misc/mp"
	_ "github.com/shylinux/icebergs/misc/wx"
)

func main() {
	println(ice.Run())
}
