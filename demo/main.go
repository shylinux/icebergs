package main

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/core/chat"
	_ "github.com/shylinux/icebergs/core/wiki"
)

func main() {
	println(ice.Run())
}
