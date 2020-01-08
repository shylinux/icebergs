#! /bin/sh

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
	@echo && date
	go build -o ice.bin main.go && chmod u+x ice.bin && ./ice.sh restart
END

    [ -f ice.sh ] || cat >> ice.sh <<END
#! /bin/sh

export PATH=\${PWD}:\$PATH
export ctx_pid=\${ctx_pid:=var/run/ice.pid}
export ctx_log=\${ctx_log:=boot.log}

prepare() {
    [ -e ice.sh ] || curl \$ctx_dev/publish/ice.sh -o ice.sh && chmod u+x ice.sh
    [ -e ice.bin ] && chmod u+x ice.bin && return

    bin="ice"
    case \`uname -s\` in
        Darwin) bin=\${bin}.darwin ;;
        Linux) bin=\${bin}.linux ;;
        *) bin=\${bin}.windows ;;
    esac
    case \`uname -m\` in
        x86_64) bin=\${bin}.amd64 ;;
        i686) bin=\${bin}.386 ;;
        arm*) bin=\${bin}.arm ;;
    esac
    curl \$ctx_dev/publish/\${bin} -o ice.bin && chmod u+x ice.bin
 }
start() {
    trap HUP hup && while true; do
        date && ice.bin \$@ 2>\$ctx_log && echo -e "\n\nrestarting..." || break
    done
}
serve() {
    prepare && shutdown && start \$@
}
restart() {
    [ -e \$ctx_pid ] && kill -2 \`cat \$ctx_pid\` || echo
}
shutdown() {
    [ -e \$ctx_pid ] && kill -3 \`cat \$ctx_pid\` || echo
}

cmd=\$1 && [ -n "\$cmd" ] && shift || cmd=serve
\$cmd \$*
END
    chmod u+x ice.sh
}

build() {
    miss=./ && [ "$1" != "" ] && miss=$1 && shift && mkdir $miss
    cd $miss && prepare && go build -o ice.bin main.go && chmod u+x ice.bin && ./ice.sh start serve dev
}

tutor() {
    mkdir $1
    [ -f "$1/$1.go" ] || cat >> "$1/$1.go" <<END
package $1

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "$1", Help: "$1",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"$1": {Name: "$1", Help: "$1", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"$1": {Name: "$1", Help: "$1", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
            m.Echo("hello world")
		}},
	},
}

func init() { cli.Index.Register(Index, nil) }

END
}

cmd=build && [ "$1" != "" ] && cmd=$1 && shift
$cmd $*
