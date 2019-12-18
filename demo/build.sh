#! /bin/sh

export ice_app=${ice_app:="ice.app"}
export ice_err=${ice_err:="boot.log"}
export ice_conf=${ice_app:="var/conf"}
export ice_serve=${ice_serve:="web.serve"}

prepare() {
    [ -f main.go ] || cat >> main.go <<END
package main

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/core/chat"
)

func main() {
	println(ice.Run())
}
END

    [ -f go.mod ] || go mod init

    [ -f Makefile ] || cat >> Makefile <<END
all:
	sh build.sh build && sh build.sh restart
END
    mkdir -p usr/template/wiki
    [ -f usr/template/common.tmpl ] || cat >> usr/template/common.tmpl <<END

END
    [ -f usr/template/wiki/common.tmpl ] || cat >> usr/template/wiki/common.tmpl <<END

END
}
build() {
    prepare && go build -o bin/shy main.go
}
start() {
    [ -d usr/volcanos ] || git clone https://github.com/shylinux/volcanos usr/volcanos
    while true; do
        date && $ice_app $* 2>$ice_err && log "\n\nrestarting..." || break
    done
}
log() { echo -e $*; }
restart() {
    kill -2 `cat var/run/shy.pid`
}
shutdown() {
    kill -3 `cat var/run/shy.pid`
}
help() {
    echo "usage: $0 cmd arg"
}

cmd=$1 && shift
[ -z "$cmd" ] && cmd=start
$cmd $*
