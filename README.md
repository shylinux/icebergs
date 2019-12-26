# Icebergs.go

icebergs是一个后端框架，通过模块化、集群化实现资源的无限的扩展与自由的组合。

一键创建项目
```
curl -s https://raw.githubusercontent.com/shylinux/icebergs/master/demo/build.sh | sh miss
```

## 1 原型 type.go
### 1.1 msg.Detail
### 1.2 msg.Option
### 1.3 msg.Append
### 1.4 msg.Result
### 1.5 msg.Travel
### 1.6 msg.Search
### 1.7 msg.Conf
### 1.8 msg.Cmd
### 1.9 msg.Cap

## 2 框架 base.go
### 2.1 注册模块 Register
### 2.2 创建资源 Begin
### 2.3 加载配置 _init
### 2.4 启动服务 Start
### 2.5 保存配置 _exit
### 2.6 释放资源 Close

## 3 基础模块 base/
### 3.1 模块中心 base/ctx
### 3.2 命令中心 base/cli
### 3.3 认证中心 base/aaa
### 3.4 网页中心 base/web

### 3.5 词法中心 base/lex
### 3.6 语法中心 base/yac
### 3.7 事件中心 base/gdb
### 3.8 日志中心 base/log

### 3.9 网络中心 base/tcp
### 3.10 文件中心 base/nfs
### 3.11 终端中心 base/ssh
### 3.12 数据中心 base/mdb

## 4 核心模块 core/
### 4.1 编程中心 core/code
### 4.2 文档中心 core/wiki
### 4.3 聊天中心 core/chat
### 4.4 团队中心 core/team
### 4.5 贸易中心 core/mall

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

