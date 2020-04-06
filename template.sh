#! /bin/sh

ice_sh="bin/ice.sh"
ice_bin="bin/ice.bin"
ice_mod="${PWD##**/}"
init_shy="etc/init.shy"
local_shy="etc/local.shy"
exit_shy="etc/exit.shy"
main_go="src/main.go"
main_js="src/main.js"
readme="README.md"
shy="src/main.shy"

prepare() {
    [ -d src ] || mkdir src

    [ -f ${main_go} ] || cat >> ${main_go} <<END
package main

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	_ "github.com/shylinux/icebergs/core"
	_ "github.com/shylinux/icebergs/misc"
)

func main() { println(ice.Run()) }
END

    [ -f ${shy} ] || cat >> ${shy} <<END
title "${ice_mod}"

field "自动化" favor 

field "启动流程" favor args "启动流程"
field "请求响应" favor args "请求响应"
field "服务集群" favor args "服务集群"

field "数据结构" favor args "数据结构"
field "系统架构" favor args "系统架构"
field "编译原理" favor args "编译原理"
END

    [ -f ${main_js} ] || cat >> ${main_js} <<END
Volcanos("onimport", {help: "导入数据", list: [],
    "init": function(can, msg, cb, output, action, option) {},
})
Volcanos("onaction", {help: "控件菜单", list: []})
Volcanos("onchoice", {help: "控件交互", list: ["刷新"]
    "刷新": function(event, can, value, cmd, target) {},
})
Volcanos("ondetail", {help: "控件详情", list: []})
Volcanos("onexport", {help: "导出数据", list: []})
END

    [ -f go.mod ] || go mod init ${ice_mod}

    [ -f Makefile ] || cat >> Makefile <<END
export GOPROXY=https://goproxy.cn
export GOPRIVATE=github.com
export CGO_ENABLED=0
all:
	@echo && date
	go build -o ${ice_bin} ${main_go} && chmod u+x ${ice_bin} && ./${ice_sh} restart
END

    [ -d etc ] || mkdir etc
    [ -f ${init_shy} ] || cat >> ${init_shy} <<END
~web

~ssh
    source etc/local.shy
END
    [ -f ${local_shy} ] || cat >> ${local_shy} <<END
~aaa

~web

END
    [ -f ${exit_shy} ] || cat >> "${exit_shy}" <<END
~web

END

    [ -d bin ] || mkdir bin
    [ -f ${ice_sh} ] || cat >> ${ice_sh} <<END
#! /bin/sh

export PATH=\${PWD}/bin:\${PWD}:\$PATH
export ctx_pid=\${ctx_pid:=var/run/ice.pid}
export ctx_log=\${ctx_log:=bin/boot.log}
export ctx_mod=\${ctx_mod:="gdb,log,ssh,ctx"}

prepare() {
    [ -d bin ] || mkdir bin
    [ -e ${ice_sh} ] || curl \$ctx_dev/publish/ice.sh -o ${ice_sh} && chmod u+x ${ice_sh}
    [ -e ${ice_bin} ] && chmod u+x ${ice_bin} && return

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
    curl \$ctx_dev/publish/\${bin} -o ${ice_bin} && chmod u+x ${ice_bin}
 }
start() {
    trap HUP hup && while true; do
        date && ice.bin \$@ 2>\$ctx_log && echo -e "\n\nrestarting..." || break
    done
}
restart() {
    [ -e \$ctx_pid ] && kill -2 \`cat \$ctx_pid\` || echo
}
shutdown() {
    [ -e \$ctx_pid ] && kill -3 \`cat \$ctx_pid\` || echo
}
serve() {
    prepare && shutdown && start \$@
}

cmd=\$1 && [ -n "\$cmd" ] && shift || cmd=serve
\$cmd \$*
END
    chmod u+x ${ice_sh}
}

build() {
	export CGO_ENABLED=0
	export GOPRIVATE=github.com
	export GOPROXY=https://goproxy.cn
    miss=./ && [ "$1" != "" ] && miss=$1 && shift && mkdir $miss
    cd $miss && prepare && go build -o ${ice_bin} ${main_go} && chmod u+x ${ice_bin} && ./${ice_sh} start serve
}

tutor() {
    [ -d $1 ] || mkdir $1

    [ -f "$1/$1.js" ] || cat >> "$1/$1.js" <<END
Volcanos("onimport", {help: "导入数据", list: [],
    "init": function(can, msg, cb, output, action, option) {},
})
Volcanos("onaction", {help: "控件菜单", list: []})
Volcanos("onchoice", {help: "控件交互", list: ["刷新"]
    "刷新": function(event, can, value, cmd, target) {},
})
Volcanos("ondetail", {help: "控件详情", list: []})
Volcanos("onexport", {help: "导出数据", list: []})
END
    [ -f "$1/$1.shy" ] || cat >> "$1/$1.shy" <<END
title "$1"
END
    [ -f "$1/$1.go" ] || cat >> "$1/$1.go" <<END
package $1

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/wiki"
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

func init() { wiki.Index.Register(Index, nil) }

END
    ls -l $1
}

cmd=build && [ "$1" != "" ] && cmd=$1 && shift
$cmd $*
