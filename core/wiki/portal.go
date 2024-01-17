package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _portal_commands(m *ice.Message, arg ...string) {
	const (
		MAIN = "main"
		BASE = "base"
		CORE = "core"
		MISC = "misc"
	)
	help := map[string]string{}
	list := map[string][]string{}
	m.Travel(func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
		if p := kit.ExtChange(cmd.FileLine(), nfs.SHY); nfs.Exists(m, p) {
			help[strings.TrimPrefix(m.PrefixKey(), "web.")] = p
		}
		if strings.Contains(cmd.FileLine(), ice.Info.Make.Module) || strings.HasPrefix(cmd.FileLine(), ice.Info.Make.Path+nfs.SRC) {
			list[MAIN] = append(list[MAIN], m.PrefixKey())
		} else if strings.Contains(cmd.FileLine(), nfs.PS+ice.ICEBERGS) {
			for _, mod := range []string{BASE, CORE, MISC} {
				if strings.Contains(cmd.FileLine(), nfs.PS+mod) {
					list[mod] = append(list[mod], m.PrefixKey())
				}
			}
		}
	})
	text := []string{"navmenu `"}
	for _, mod := range []string{BASE, CORE, MISC} {
		text = append(text, kit.Format("%s %s/", map[string]string{MAIN: "业务模块", BASE: "基础模块", CORE: "核心模块", MISC: "其它模块"}[mod], mod))
		last := "ice"
		for _, cmd := range list[mod] {
			kit.If(mod != BASE, func() { cmd = strings.TrimPrefix(cmd, "web.") })
			if !strings.HasPrefix(cmd, last) {
				last = strings.Split(cmd, nfs.PT)[0]
				if p := path.Join(nfs.USR_LEARNING_PORTAL, path.Join(arg...), mod, last); nfs.Exists(m, p) {
					text = append(text, kit.Format("	%s %s/", last, last))
				}
			}
			cmd = strings.TrimPrefix(cmd, last+nfs.PT)
			if p := path.Join(nfs.USR_LEARNING_PORTAL, path.Join(arg...), mod, last, strings.Replace(cmd, nfs.PT, nfs.PS, -1)+".shy"); nfs.Exists(m, p) {
				text = append(text, kit.Format("		%s %s.shy", cmd, cmd))
			} else if p, ok := help[last+nfs.PT+cmd]; ok {
				text = append(text, kit.Format("		%s %s", cmd, strings.TrimPrefix(ctx.FileURI(p), "/require/")))
			}
		}
	}
	text = append(text, "`")
	m.Cmd(nfs.SAVE, path.Join(nfs.USR_LEARNING_PORTAL, path.Join(arg...), INDEX_SHY), strings.Join(text, lex.NL))
}

const (
	HEADER    = "header"
	NAV       = "nav"
	INDEX_SHY = "index.shy"
)

const PORTAL = "portal"

func init() {
	Index.MergeCommands(ice.Commands{
		PORTAL: {Name: "portal path auto", Help: "网站门户", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) { web.RenderCmd(m, "", arg) }},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if p := path.Join(ice.USR_PORTAL, path.Join(arg...)); (m.Option(ice.DEBUG) == ice.TRUE || !nfs.ExistsFile(m, p)) && aaa.Right(m.Spawn(), arg) {
					ctx.Run(m, arg...)
					m.Cmd(nfs.SAVE, p, ice.Maps{nfs.CONTENT: m.FormatMeta(), nfs.DIR_ROOT: ""})
				} else {
					m.Copy(m.Spawn([]byte(m.Cmdx(nfs.CAT, p))))
				}
			}},
			ice.CTX_INIT:     {Hand: web.DreamWhiteHandle},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) { m.PushButton(kit.Dict(PORTAL, "官网")) }},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, nil, arg...) }},
		}, aaa.WhiteAction("")), Hand: func(m *ice.Message, arg ...string) {
			if m.Push(HEADER, m.Cmdx(WORD, path.Join(nfs.USR_LEARNING_PORTAL, INDEX_SHY))); len(arg) > 0 {
				kit.If(path.Join(arg...) == "commands", func() { _portal_commands(m, arg...) })
				m.Push(NAV, m.Cmdx(WORD, path.Join(nfs.USR_LEARNING_PORTAL, path.Join(arg...), INDEX_SHY)))
			}
			m.Display("")
		}},
	})
}
