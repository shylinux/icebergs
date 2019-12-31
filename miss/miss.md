# icebergs

icebergs是一个后端框架，通过模块化、集群化实现资源的无限的扩展与自由的组合。

一键创建项目
```
mkdir miss; cd miss && curl -s https://shylinux.com/publish/build.sh | sh
```

## 命令模块 base/cli

cli模块用于与系统进行交互。

- 系统信息 ice.CLI_RUNTIME
- 系统命令 ice.CLI_SYSTEM

## 文件模块 base/nfs

nfs模块用于管理文件的读写。

## 终端模块 base/ssh

ssh模块用于与终端交互。

## 数据模块 base/mdb

mdb模块用于管理数据的读写。

## 日志模块 base/log

log模块负责输出日志。

## 事件模块 base/gdb

gdb模块会根据各种触发条件，择机执行各种命令。

- 信号器 ice.SIGNAL
- 定时器 ice.TIMER
- 触发器 ice.EVENT

## 认证模块 base/aaa

aaa模块用于各种权限管理与身份认证。

- 角色 ice.AAA_ROLE
- 用户 ice.AAA_USER
- 会话 ice.AAA_SESS

## 网页模块 base/web

web模块用于组织网络节点，与生成前端网页，

- 网络爬虫 ice.WEB_SPIDE
- 网络服务 ice.WEB_SERVE
- 网络节点 ice.WEB_SPACE
- 网络任务 ice.WEB_DREAM

- 网络收藏 ice.WEB_FAVOR
- 网络缓存 ice.WEB_CACHE
- 网络存储 ice.WEB_STORY
- 网络共享 ice.WEB_SHARE

- 网络路由 ice.WEB_ROUTE
- 网络代理 ice.WEB_PROXY
- 网络分组 ice.WEB_GROUP
- 网络标签 ice.WEB_LABEL

