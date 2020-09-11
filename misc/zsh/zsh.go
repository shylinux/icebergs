package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"net/url"
	"path"
)

const (
	ZSH = "zsh"
)

var Index = &ice.Context{Name: ZSH, Help: "命令行",
	Configs: map[string]*ice.Config{
		ZSH: {Name: ZSH, Help: "命令行", Value: kit.Data(
			"source", "https://sourceforge.net/projects/zsh/files/zsh/5.8/zsh-5.8.tar.xz",

			"proxy", "tmux", "history", "zsh.history", "script", []interface{}{
				".vim/syntax/sh.vim", "etc/conf/sh.vim",
				".bashrc", "etc/conf/bashrc",
				".zshrc", "etc/conf/zshrc",
			},
		)},
	},
	Commands: map[string]*ice.Command{
		ZSH: {Name: "zsh port=auto path=auto auto 启动:button 构建:button 下载:button", Help: "命令行", Action: map[string]*ice.Action{
			"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "download", m.Conf(ZSH, kit.META_SOURCE))
			}},
			"build": {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "build", m.Conf(ZSH, kit.META_SOURCE))
			}},
			"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv("prepare", func(p string) []string {
					m.Option(cli.CMD_DIR, p)
					return []string{}
				})
				m.Cmdy(code.INSTALL, "start", m.Conf(ZSH, kit.META_SOURCE), "bin/zsh")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(ZSH, kit.META_SOURCE)), arg)
		}},

		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			// m.Cmd("web.spide_rewrite", "create", "from", "https://sourceforge.net/projects/zsh/files/zsh/5.8/zsh-5.8.tar.xz", "to", "http://localhost:9020/publish/zsh-5.8.tar.gz")

			m.Conf(web.FAVOR, "meta.render.shell", m.AddCmd(&ice.Command{Name: "render type name text", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				value, _ := m.Optionv(kit.MDB_VALUE).(map[string]interface{})
				m.Option("cmd_dir", kit.Value(value, "extra.pwd"))
				m.Cmdy(cli.SYSTEM, kit.Select(kit.Format(value["text"]), arg, 2))
			}}))
			m.Conf(web.FAVOR, "meta.render.cmd", m.AddCmd(&ice.Command{Name: "render type name text", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				value, _ := m.Optionv(kit.MDB_VALUE).(map[string]interface{})
				m.Cmdy(kit.Split(kit.Format(kit.Select(kit.Format(value["text"], arg, 2)))))
			}}))
			m.Conf(web.FAVOR, "meta.render.bin", m.AddCmd(&ice.Command{Name: "render type name text", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(cli.SYSTEM, "file", arg[2])
			}}))
		}},

		"/help": {Name: "/help", Help: "帮助", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("help")
		}},
		"/login": {Name: "/login", Help: "登录", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("login", "init", c.Name)
		}},
		"/logout": {Name: "/logout", Help: "登出", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy("login", "exit")
		}},

		"/ish": {Name: "/ish", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if sub, e := url.QueryUnescape(m.Option("sub")); m.Assert(e) {
				m.Cmdy(kit.Split(sub))
				if len(m.Resultv()) == 0 {
					m.Table()
				}
			}
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
