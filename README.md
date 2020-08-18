# icebergs

icebergs是一个应用框架，通过模块化、集群化、自动化，快速搭建起完整的个人云计算平台。

- 使用icebergs可以将各种模块或项目集成到一起，快速开发出集中式的服务器。
- 使用icebergs可以将各种设备自由的组合在一起，快速搭建起分布式的服务器。

## 0. 搭建服务
### 0.1 一键部署
```sh
mkdir miss; cd miss && curl -s https://shylinux.com/publish/ice.sh | sh
```

脚本会根据当前系统类型，自动下载程序文件ice.bin，并自动启动服务。

### 0.2 使用方式
**终端交互**

启动后的进程，像bash一样是一个可交互的shell，可以执行各种模块命令或系统命令。

**网页交互**

默认还会启动一个web服务，访问地址 http://localhost:9020 ，就可以通过网页进行操作。

**重启服务**

在终端按Ctrl+C，就可以重新启动服务。

**结束服务**

在终端按Ctrl+\，就可以停止服务。

### 0.3 使用示例

## 1. 项目开发
icebergs是一个应用框架，如果官方模块无法满足使用需求，还可以搜集第三方模块，自行编译程序。

如果第三方模块也无法满足使用需求，还可以自己开发模块，
icebergs提供了模板，可以一键创建新模块，快速添加自己的功能模块。

### 1.1 部署环境
*开发环境，需要提前安装好git和golang*
```sh
mkdir miss; cd miss && curl -s https://shylinux.com/publish/template.sh | sh
```
template.sh会自动创建出项目模板，并自动编译生成程序，然后启动服务。

为了方便以后创建项目与模块。
可以将辅助脚本template.sh下载，并添加到可执行目录中。

### 1.2 添加第三方模块

在src/main.go文件中，就可以import任意的第三方模块，
执行一下make命令，就会重新生成ice.bin。
重新启动服务，就可以使用第三方模块了。

### 1.3 开发模块
```sh
template.sh tutor hello
```
使用之前下载的template.sh，调用tutor命令，并指定模块名称hello，就可以一键创建模块了。

在src/main.go 中import新加的模块，
执行make命令，程序编译完成后，
重启服务，就可以使用新模块了。

### 1.4 开发框架
如果现有的框架，无法满足需求，还可以下载框架源码自行更改。

```sh
git clone https://github.com/shylinux/icebergs usr/icebergs
```
修改go.mod文件，引用本地框架。
```go
replace github.com/shylinux/icebergs => ./usr/icebergs
```

## 2 原型 type.go
### 2.1 msg.Detail
### 2.2 msg.Option
### 2.3 msg.Append
### 2.4 msg.Result
### 2.5 msg.Travel
### 2.6 msg.Search
### 2.7 msg.Conf
### 2.8 msg.Cmd
### 2.9 msg.Cap

## 3 框架 base.go
### 3.1 注册模块 Register
### 3.2 创建资源 Begin
### 3.3 加载配置 _init
### 3.4 启动服务 Start
### 3.5 保存配置 _exit
### 3.6 释放资源 Close

## 4 基础模块 base/
### 4.1 模块中心 base/ctx/
### 4.2 命令中心 base/cli/
### 4.3 认证中心 base/aaa/
### 4.4 网页中心 base/web/

### 4.5 词法中心 base/lex/
### 4.6 语法中心 base/yac/
### 4.7 事件中心 base/gdb/
### 4.8 日志中心 base/log/

### 4.9 网络中心 base/tcp/
### 4.10 文件中心 base/nfs/
### 4.11 终端中心 base/ssh/
### 4.12 数据中心 base/mdb/

## 5 核心模块 core/
### 5.1 编程中心 core/code/
### 5.2 文档中心 core/wiki/
### 5.3 聊天中心 core/chat/
### 5.4 团队中心 core/team/
### 5.5 贸易中心 core/mall/

## 6 其它模块 misc/
### 6.1 终端管理 misc/zsh/
### 6.1 终端管理 misc/tmux/
### 6.1 代码管理 misc/git/
### 6.1 代码管理 misc/vim/
### 6.1 公众号 misc/mp/
### 6.1 小程序 misc/wx/
### 6.1 浏览器 misc/chrome/
### 6.1 机器人 misc/lark/
### 6.1 开发板 misc/pi/
