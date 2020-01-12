{{title "ICEBERGS"}}
{{brief `icebergs是一个后端框架，通过模块化、集群化实现资源的无限的扩展与自由的组合。`}}

{{chain "icebergs" `
icebergs
    type.go
        msg.Detail
        msg.Option
        msg.Append
        msg.Result
        msg.Travel
        msg.Search
        msg.Conf
        msg.Cmd
        msg.Cap
    base.go bg blue
        Begin
        _init
        Start bg red
            code
            wiki
            chat
                ocean
                river
                action
                storm
                steam
            team
            mall
        _exit
        Close
    conf.go
        init
        host
        boot
        node
        user
        work
        auth
        data
        file
` "" "" 16}}

{{shell "一键创建项目" "usr" "install" `mkdir miss; cd miss && curl -s https://shylinux.com/publish/build.sh | sh`}}

{{shell "一键启动项目" "usr" "install" `mkdir miss; cd miss && curl -s https://shylinux.com/publish/ice.sh | sh`}}

{{chapter "配置模块 base/ctx"}}
{{chapter "命令模块 base/cli"}}
cli模块用于与系统进行交互。

{{order "命令" `
系统信息 ice.CLI_RUNTIME
系统命令 ice.CLI_SYSTEM
`}}

{{chapter "通信模块 base/tcp"}}
tcp模块用于管理网络的读写

{{chapter "存储模块 base/nfs"}}
nfs模块用于管理文件的读写。

{{chapter "终端模块 base/ssh"}}
ssh模块用于与终端交互。

{{chapter "数据模块 base/mdb"}}
mdb模块用于管理数据的读写。

{{chapter "词法模块 base/lex"}}

{{chapter "语法模块 base/yac"}}

{{chapter "日志模块 base/log"}}
log模块负责输出日志。

{{chapter "事件模块 base/gdb"}}
gdb模块会根据各种触发条件，择机执行各种命令。

{{order "命令" `
信号器 ice.SIGNAL
定时器 ice.TIMER
触发器 ice.EVENT
`}}

{{chapter "认证模块 base/aaa"}}
aaa模块用于各种权限管理与身份认证。

{{order "命令" `
角色 ice.AAA_ROLE
用户 ice.AAA_USER
会话 ice.AAA_SESS
`}}

{{chapter "网络模块 base/web"}}
web模块用于组织网络节点，与生成前端网页，

{{section "网络爬虫 ice.WEB_SPIDE"}}
WEB_SPIDE功能，用于发送网络请求获取相关数据。

{{section "网络服务 ice.WEB_SERVE"}}
WEB_SERVE功能，用于启动网络服务器接收网络请求。

{{section "网络节点 ice.WEB_SPACE"}}
WEB_SPACE功能，用于与相连网络节点进行通信。

{{section "网络任务 ice.WEB_DREAM"}}
WEB_DREAM功能，用于启动本地节点，管理各种任务的相关资源。

{{section "网络收藏 ice.WEB_FAVOR"}}
WEB_FAVOR功能，用于收藏各种实时数据，进行分类管理。

{{section "网络缓存 ice.WEB_CACHE"}}
WEB_CACHE功能，用于管理缓存数据，自动存储与传输。

{{section "网络存储 ice.WEB_STORY"}}
WEB_STORY功能，用于记录数据的历史变化，可以查看任意历史版本。

{{section "网络共享 ice.WEB_SHARE"}}
WEB_SHARE功能，用于数据与应用的共享，可以查到所有数据流通记录。

{{section "网络路由 ice.WEB_ROUTE"}}
{{section "网络代理 ice.WEB_PROXY"}}
{{section "网络分组 ice.WEB_GROUP"}}
{{section "网络标签 ice.WEB_LABEL"}}

