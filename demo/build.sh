#! /bin/sh

export ice_app=${ice_app:="ice.app"}
export ice_err=${ice_err:="boot.log"}
export ice_serve=${ice_serve:="web.serve"}
export ice_can=${ice_can:="https://github.com/shylinux/volcanos"}

prepare() {
    [ -f main.go ] || cat >> main.go <<END
package main

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	_ "github.com/shylinux/icebergs/core"
	_ "github.com/shylinux/icebergs/misc"
)

func main() {
	println(ice.Run())
}
END

    [ -f go.mod ] || go mod init ${PWD##**/}

    [ -f Makefile ] || cat >> Makefile <<END
all:
	sh build.sh build && sh build.sh restart
END
}
build() {
    prepare && go build -o bin/shy main.go
}

start() {
    [ -z "$@" ] && ( [ -d usr/volcanos ] || git clone $ice_can usr/volcanos )
    while true; do
        date && $ice_app $@ 2>$ice_err && echo -e "\n\nrestarting..." || break
    done
}
restart() {
    kill -2 `cat var/run/shy.pid`
}
shutdown() {
    kill -3 `cat var/run/shy.pid`
}

cmd=$1 && shift
[ -z "$cmd" ] && cmd=start
$cmd $*
