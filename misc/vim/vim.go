package vim

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _vim_pkg(m *ice.Message, url string) string {
	p := path.Join(m.Conf(code.INSTALL, kit.Keym(nfs.PATH)), path.Base(url))
	return kit.Path(m.Conf(code.INSTALL, kit.Keym(nfs.PATH)), strings.Split(m.Cmdx(cli.SYSTEM, "sh", "-c", kit.Format("tar tf %s| head -n1", p)), "/")[0])
}

const VIM = "vim"

var Index = &ice.Context{Name: VIM, Help: "编辑器", Configs: map[string]*ice.Config{
	VIM: {Name: VIM, Help: "编辑器", Value: kit.Data(
		nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/vim/vim-8.2.2681.tar.gz",
	)},
}, Commands: map[string]*ice.Command{
	VIM: {Name: "vim path auto order build download", Help: "编辑器", Action: ice.MergeAction(map[string]*ice.Action{
		cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.BUILD, _vim_pkg(m, m.Config(nfs.SOURCE)), "--enable-multibyte=yes",
				"--enable-pythoninterp=yes", "--enable-luainterp=yes", "--enable-cscope=yes")
		}},
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(code.INSTALL, cli.ORDER, _vim_pkg(m, m.Config(nfs.SOURCE)), "_install/bin")
		}},
	}, code.InstallAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmdy(code.INSTALL, nfs.SOURCE, _vim_pkg(m, m.Config(nfs.SOURCE)), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
