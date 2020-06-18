package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
	"github.com/skip2/go-qrcode"

	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

type Frame struct {
	*http.Client
	*http.Server
	*http.ServeMux
	m *ice.Message

	send map[string]*ice.Message
}

func Count(m *ice.Message, cmd, key, name string) int {
	count := kit.Int(m.Conf(cmd, kit.Keys(key, name)))
	m.Conf(cmd, kit.Keys(key, name), count+1)
	return count
}
func Format(key string, arg ...interface{}) string {
	switch args := kit.Simple(arg); key {
	case "a":
		return fmt.Sprintf("<a href='%s' target='_blank'>%s</a>", kit.Format(args[0]), kit.Select(kit.Format(args[0]), args, 1))
	}
	return ""
}
func Render(msg *ice.Message, cmd string, args ...interface{}) {
	if cmd != "" {
		defer func() { msg.Log(ice.LOG_EXPORT, "%s: %v", cmd, args) }()
	}
	switch arg := kit.Simple(args...); cmd {
	case ice.RENDER_VOID:
	case ice.RENDER_OUTPUT:
	case "redirect":
		http.Redirect(msg.W, msg.R, kit.MergeURL(arg[0], arg[1:]), 307)

	case "refresh":
		arg = []string{"200", fmt.Sprintf(`<!DOCTYPE html><head><meta charset="utf-8"><meta http-equiv="Refresh" content="%d"></head><body>%s</body>`,
			kit.Int(kit.Select("3", arg, 0)), kit.Select("请稍后，系统初始化中...", arg, 1),
		)}
		fallthrough

	case "status":
		msg.W.WriteHeader(kit.Int(kit.Select("200", arg, 0)))
		msg.W.Write([]byte(kit.Select("", arg, 1)))

	case "cookie":
		expire := time.Now().Add(kit.Duration(msg.Conf(ice.AAA_SESS, "meta.expire")))
		http.SetCookie(msg.W, &http.Cookie{Value: arg[0], Name: kit.Select(ice.MSG_SESSID, arg, 1), Path: "/", Expires: expire})

	case ice.RENDER_DOWNLOAD:
		msg.W.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(arg[0]), arg, 2)))
		msg.W.Header().Set("Content-Type", kit.Select("text/html", arg, 1))
		http.ServeFile(msg.W, msg.R, arg[0])

	case ice.RENDER_RESULT:
		if len(arg) > 0 {
			msg.W.Write([]byte(kit.Format(arg[0], args[1:]...)))
		} else {
			args = append(args, "length:", len(msg.Result()))
			msg.W.Write([]byte(msg.Result()))
		}

	case ice.RENDER_QRCODE:
		if qr, e := qrcode.New(arg[0], qrcode.Medium); msg.Assert(e) {
			msg.W.Header().Set("Content-Type", "image/png")
			msg.Assert(qr.Write(kit.Int(kit.Select("256", arg, 1)), msg.W))
		}

	default:
		if cmd != "" {
			msg.Echo(kit.Format(cmd, args...))
		}
		msg.W.Header().Set("Content-Type", "application/json")
		fmt.Fprint(msg.W, msg.Formats("meta"))
	}
	msg.Append(ice.MSG_OUTPUT, ice.RENDER_OUTPUT)
}

func (web *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (web *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	web.send = map[string]*ice.Message{}
	return web
}
func (web *Frame) Start(m *ice.Message, arg ...string) bool {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		if w, ok := s.Server().(*Frame); ok {
			if w.ServeMux != nil {
				return
			}
			w.ServeMux = http.NewServeMux()

			// 静态路由
			msg := m.Spawns(s)
			m.Confm(SERVE, "meta.static", func(key string, value string) {
				m.Log("route", "%s <- %s <- %s", s.Name, key, value)
				w.Handle(key, http.StripPrefix(key, http.FileServer(http.Dir(value))))
			})

			// 级联路由
			route := "/" + s.Name + "/"
			if n, ok := p.Server().(*Frame); ok && n.ServeMux != nil {
				msg.Log("route", "%s <= %s", p.Name, route)
				n.Handle(route, http.StripPrefix(path.Dir(route), w))
			}

			// 命令路由
			m.Travel(func(p *ice.Context, sub *ice.Context, k string, x *ice.Command) {
				if s == sub && k[0] == '/' {
					msg.Log("route", "%s <- %s", s.Name, k)
					Trans(w, msg, k, x)
				}
			})
		}
	})

	// TODO simple
	m.Richs(ice.WEB_SPIDE, nil, arg[0], func(key string, value map[string]interface{}) {
		client := value["client"].(map[string]interface{})

		// 服务地址
		port := m.Cap(ice.CTX_STREAM, client["hostname"])
		m.Log("serve", "listen %s %s %v", arg[0], port, m.Conf(ice.CLI_RUNTIME, "node"))

		// 启动服务
		web.m, web.Server = m, &http.Server{Addr: port, Handler: web}
		m.Event(ice.SERVE_START, arg[0])
		m.Warn(true, "listen %s", web.Server.ListenAndServe())
		m.Event(ice.SERVE_CLOSE, arg[0])
	})
	return true
}
func (web *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

var Index = &ice.Context{Name: "web", Help: "网络模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()

			SpideCreate(m, "self", kit.Select("http://:9020", m.Conf(ice.CLI_RUNTIME, "conf.ctx_self")))
			SpideCreate(m, "dev", kit.Select("http://:9020", m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev")))
			SpideCreate(m, "shy", kit.Select("https://shylinux.com:443", m.Conf(ice.CLI_RUNTIME, "conf.ctx_shy")))

			m.Cmd(ice.APP_SEARCH, "add", "favor", "base", m.AddCmd(&ice.Command{Name: "search word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Richs(ice.WEB_FAVOR, nil, arg[1], func(key string, value map[string]interface{}) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "id", arg[2], func(index int, value map[string]interface{}) {
							if cmd := m.Conf(ice.WEB_FAVOR, kit.Keys("meta.render", value["type"])); cmd != "" {
								m.Optionv("value", value)
								m.Cmdy(cmd, arg[1:])
							} else {
								m.Push("detail", value)
							}
						})
					})
					return
				}

				m.Option("cache.limit", -2)
				wg := &sync.WaitGroup{}
				m.Richs(ice.WEB_FAVOR, nil, "*", func(key string, val map[string]interface{}) {
					favor := kit.Format(kit.Value(val, "meta.name"))
					wg.Add(1)
					m.Gos(m, func(m *ice.Message) {
						m.Grows(ice.WEB_FAVOR, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
							if favor == arg[0] || value["type"] == arg[0] ||
								strings.Contains(kit.Format(value["name"]), arg[0]) || strings.Contains(kit.Format(value["text"]), arg[0]) {
								m.Push("pod", m.Option(ice.MSG_USERPOD))
								m.Push("engine", "favor")
								m.Push("favor", favor)
								m.Push("", value, []string{"id", "time", "type", "name", "text"})
							}
						})
						wg.Done()
					})
				})
				wg.Wait()
			}}))

			m.Cmd(ice.APP_SEARCH, "add", "story", "base", m.AddCmd(&ice.Command{Name: "search word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Cmdy(ice.WEB_STORY, "index", arg[2])
					return
				}

				m.Richs(ice.WEB_STORY, "head", "*", func(key string, val map[string]interface{}) {
					if val["story"] == arg[0] {
						m.Push("pod", m.Option(ice.MSG_USERPOD))
						m.Push("engine", "story")
						m.Push("favor", val["story"])
						m.Push("id", val["list"])

						m.Push("time", val["time"])
						m.Push("type", val["scene"])
						m.Push("name", val["story"])
						m.Push("text", val["count"])
					}
				})
			}}))

			m.Cmd(ice.APP_SEARCH, "add", "share", "base", m.AddCmd(&ice.Command{Name: "search word", Help: "搜索引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch arg[0] {
				case "set":
					m.Cmdy(ice.WEB_SHARE, arg[2])
					return
				}

				m.Option("cache.limit", -2)
				m.Grows(ice.WEB_SHARE, nil, "", "", func(index int, value map[string]interface{}) {
					if value["share"] == arg[0] || value["type"] == arg[0] ||
						strings.Contains(kit.Format(value["name"]), arg[0]) || strings.Contains(kit.Format(value["text"]), arg[0]) {
						m.Push("pod", m.Option(ice.MSG_USERPOD))
						m.Push("engine", "share")
						m.Push("favor", value["type"])
						m.Push("id", value["share"])

						m.Push("time", value["time"])
						m.Push("type", value["type"])
						m.Push("name", value["name"])
						m.Push("text", value["text"])
					}
				})
			}}))

			m.Conf(ice.WEB_FAVOR, "meta.render.bench", m.AddCmd(&ice.Command{Name: "render type name text arg...", Help: "渲染引擎", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy("web.code.bench", "action", "show", arg)
			}}))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(ice.WEB_SPIDE, ice.WEB_SERVE, ice.WEB_GROUP, ice.WEB_LABEL,
				ice.WEB_FAVOR, ice.WEB_CACHE, ice.WEB_STORY, ice.WEB_SHARE)

			m.Done()
			m.Richs(ice.WEB_SPACE, nil, "*", func(key string, value map[string]interface{}) {
				if kit.Format(value["type"]) == "master" {
					m.Done()
				}
			})
		}},
	},
}

func init() {
	ice.Index.Register(Index, &Frame{},
		SPIDE, SERVE, SPACE, DREAM,
	)
}
