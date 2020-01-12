# Icebergs.go

icebergs是一个后端框架，通过模块化、集群化实现资源的无限的扩展与自由的组合。

- 使用icebergs可以将各种模块或项目集成到一起，快速开发出集中式的服务器。
- 使用icebergs可以将各种设备自由的组合在一起，快速搭建起分布式的服务器。

所以通过icebergs开发出来的各种模块，无需任何多余代码，就可以独立运行，可以成为系统命令，可以远程调用，可以成为前端插件，可以成为小程序页面。
从而提高代码的复用性与灵活性。

## 1. 项目开发

### 1.1 一键创建项目
*开发环境，需要安装git和golang*
```sh
mkdir miss; cd miss && curl -s https://shylinux.com/publish/template.sh | sh
```
### 1.2 一键部署项目
*运行环境，如需通过前端页面访问服务，则需要安装git*
```sh
export ctx_dev=http://127.0.0.1:9020 && curl -s $ctx_dev/publish/ice.sh | sh
```
*ctx_dev是开发机地址，不必是本机地址，可以是任意一台先前创建过项目的机器地址。*

## 2 原型设计 type.go
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
### 4.1 模块中心 base/ctx
### 4.2 命令中心 base/cli
### 4.3 认证中心 base/aaa
### 4.4 网页中心 base/web

### 4.5 词法中心 base/lex
### 4.6 语法中心 base/yac
### 4.7 事件中心 base/gdb
### 4.8 日志中心 base/log

### 4.9 网络中心 base/tcp
### 4.10 文件中心 base/nfs
### 4.11 终端中心 base/ssh
### 4.12 数据中心 base/mdb

## 5 核心模块 core/
### 5.1 编程中心 core/code
### 5.2 文档中心 core/wiki
### 5.3 聊天中心 core/chat
### 5.4 团队中心 core/team
### 5.5 贸易中心 core/mall

## 5 配置 conf.go
### 5.1 环境 init
### 5.2 主机 host
### 5.3 启动 boot
### 5.4 节点 node
### 5.5 用户 user
### 5.6 群组 work
### 5.7 认证 auth
### 5.8 数据 data
### 5.9 文件 file

