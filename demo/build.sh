#! /bin/sh

ice_sh=${ice_sh:="ice.sh"}

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

    [ -f ${ice_sh} ] || cat >> ${ice_sh} <<END
#! /bin/sh

export PATH=\${PWD}/bin:\$PATH
prepare() {
    which ice.bin && return
    curl -s https://shylinux.com/publish/ice.bin -o bin/ice.bin
 }
start() {
    prepare && while true; do
        date && ice.bin \$@ 2>boot.log && break || echo -e "\n\nrestarting..."
    done
}
restart() {
    kill -2 \`cat var/run/shy.pid\`
}
shutdown() {
    kill -3 \`cat var/run/shy.pid\`
}

cmd=\$1 && shift
[ -z "\$cmd" ] && cmd=start
\$cmd \$*
END
    chmod u+x ${ice_sh}

    [ -f Makefile ] || cat >> Makefile <<END
all:
	go build -o bin/ice.bin main.go && chmod u+x bin/ice.bin && ./${ice_sh} restart
END
}

build() {
    [ "$1" != "" ] && mdkir $1 && cd $1
    prepare && go build -o bin/ice.bin main.go && chmod u+x bin/ice.bin && ./${ice_sh}
}

cmd=$1 && shift
[ -z "$cmd" ] && cmd=build
$cmd $*
