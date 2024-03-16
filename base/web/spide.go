package web

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _spide_create(m *ice.Message, name, link, types, icons, token string) {
	if u, e := url.Parse(link); !m.WarnNotValid(e != nil || link == "", link) {
		dir, file := path.Split(u.EscapedPath())
		m.Logs(mdb.INSERT, SPIDE, name, LINK, link)
		mdb.HashSelectUpdate(m, mdb.HashCreate(m, CLIENT_NAME, name), func(value ice.Map) {
			value[mdb.ICONS], value[TOKEN] = icons, kit.Select(kit.Format(value[TOKEN]), token)
			value[SPIDE_CLIENT] = kit.Dict(mdb.NAME, name, mdb.TYPE, types,
				SPIDE_METHOD, http.MethodGet, URL, link, ORIGIN, u.Scheme+"://"+u.Host,
				tcp.PROTOCOL, u.Scheme, tcp.HOSTNAME, u.Hostname(), tcp.HOST, u.Host, nfs.PATH, dir, nfs.FILE, file, cli.TIMEOUT, "300s",
			)
		})
	}
}
func _spide_show(m *ice.Message, name string, arg ...string) {
	file := ""
	action, arg := _spide_args(m, arg, SPIDE_RAW, SPIDE_MSG, SPIDE_SAVE, SPIDE_CACHE)
	kit.If(action == SPIDE_SAVE, func() { file, arg = arg[0], arg[1:] })
	msg := mdb.HashSelects(m.Spawn(), name)
	method, arg := _spide_args(m, arg, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
	method = kit.Select(http.MethodGet, msg.Append(CLIENT_METHOD), method)
	uri, arg := arg[0], arg[1:]
	body, head, arg := _spide_body(m, method, arg...)
	if m.Option("spide.break") == ice.TRUE {
		return
	}
	if c, ok := body.(io.Closer); ok {
		defer c.Close()
	}
	_uri := kit.MergeURL2(msg.Append(CLIENT_URL), uri, arg)
	req, e := http.NewRequest(method, _uri, body)
	if m.WarnNotValid(e, uri) {
		return
	}
	mdb.HashSelectDetail(m, name, func(value ice.Map) { _spide_head(m, req, head, value) })
	if m.Option(log.DEBUG) == ice.TRUE {
		kit.For(req.Header, func(k string, v []string) { m.Logs(REQUEST, k, v) })
	}
	res, e := _spide_send(m, name, req, kit.Format(m.OptionDefault(CLIENT_TIMEOUT, msg.Append(CLIENT_TIMEOUT))))
	if m.WarnNotFound(e, SPIDE, uri) {
		return
	}
	defer res.Body.Close()
	m.Cost(cli.STATUS, res.Status, nfs.SIZE, kit.FmtSize(kit.Int64(res.Header.Get(html.ContentLength))), mdb.TYPE, res.Header.Get(html.ContentType))
	if action != SPIDE_RAW {
		m.Push(mdb.TYPE, STATUS).Push(mdb.NAME, res.StatusCode).Push(mdb.VALUE, res.Status)
	}
	m.Options(STATUS, res.Status)
	kit.For(res.Header, func(k string, v []string) {
		if m.Option(log.DEBUG) == ice.TRUE {
			m.Logs(RESPONSE, k, v)
		}
		if m.Options(k, v); action != SPIDE_RAW {
			m.Push(mdb.TYPE, SPIDE_HEADER).Push(mdb.NAME, k).Push(mdb.VALUE, v[0])
		}
	})
	mdb.HashSelectUpdate(m, name, func(value ice.Map) {
		kit.For(res.Cookies(), func(v *http.Cookie) {
			kit.Value(value, kit.Keys(SPIDE_COOKIE, v.Name), v.Value)
			if m.Option(log.DEBUG) == ice.TRUE {
				m.Logs(RESPONSE, v.Name, v.Value)
			}
			if action != SPIDE_RAW {
				m.Push(mdb.TYPE, COOKIE).Push(mdb.NAME, v.Name).Push(mdb.VALUE, v.Value)
			}
		})
	})
	if m.WarnNotValid(res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated, uri, cli.STATUS, res.Status) {
		switch res.StatusCode {
		case http.StatusNotFound, http.StatusUnauthorized:
			return
		}
	}
	_spide_save(m, action, file, uri, res)
}
func _spide_args(m *ice.Message, arg []string, val ...string) (string, []string) {
	if kit.IndexOf(val, arg[0]) > -1 {
		return arg[0], arg[1:]
	}
	return "", arg
}
func _spide_body(m *ice.Message, method string, arg ...string) (io.Reader, ice.Maps, []string) {
	body, ok := m.Optionv(SPIDE_BODY).(io.Reader)
	if ok || method == http.MethodGet || len(arg) == 0 {
		return body, nil, arg
	}
	head := ice.Maps{}
	switch kit.If(len(arg) == 1, func() { arg = []string{SPIDE_DATA, arg[0]} }); arg[0] {
	case SPIDE_FORM:
		arg = kit.Simple(arg, func(v string) string { return url.QueryEscape(v) })
		head[html.ContentType], body = html.ApplicationForm, bytes.NewBufferString(kit.JoinQuery(arg[1:]...))
	case SPIDE_PART:
		head[html.ContentType], body = _spide_part(m, arg...)
	case SPIDE_FILE:
		if f, e := nfs.OpenFile(m, arg[1]); m.Assert(e) {
			m.Logs(nfs.LOAD, nfs.FILE, arg[1])
			body = f
		}
	case SPIDE_DATA:
		head[html.ContentType], body = html.ApplicationJSON, bytes.NewBufferString(kit.Select("{}", arg, 1))
	case SPIDE_JSON:
		arg = arg[1:]
		fallthrough
	default:
		data := ice.Map{}
		kit.For(arg, func(k, v string) { kit.Value(data, k, v) })
		head[html.ContentType], body = html.ApplicationJSON, bytes.NewBufferString(kit.Format(data))
	}
	return body, head, arg[:0]
}
func _spide_part(m *ice.Message, arg ...string) (string, io.Reader) {
	buf := &bytes.Buffer{}
	mp := multipart.NewWriter(buf)
	defer mp.Close()
	size, cache := int64(0), time.Now().Add(-time.Hour*240000)
	for i := 1; i < len(arg)-1; i += 2 {
		if arg[i] == nfs.SIZE {
			size = kit.Int64(arg[i+1])
		} else if arg[i] == SPIDE_CACHE {
			if t, e := time.ParseInLocation(ice.MOD_TIME, arg[i+1], time.Local); !m.WarnNotValid(e) {
				cache = t
			}
		} else if strings.HasPrefix(arg[i+1], mdb.AT) {
			p := arg[i+1][1:]
			if s, e := nfs.StatFile(m, p); !m.WarnNotValid(e) {
				if s.Size() == size && s.ModTime().Before(cache) {
					m.Option("spide.break", ice.TRUE)
					continue
				} else if s.Size() == size && !nfs.Exists(m.Spawn(kit.Dict(ice.MSG_FILES, nfs.DiskFile)), p) {
					m.Option("spide.break", ice.TRUE)
					continue
				}
				m.Logs(nfs.FIND, LOCAL, s.ModTime(), nfs.SIZE, s.Size(), CACHE, cache, nfs.SIZE, size)
			}
			if f, e := nfs.OpenFile(m, p); !m.WarnNotValid(e, arg[i+1]) {
				defer f.Close()
				if p, e := mp.CreateFormFile(arg[i], path.Base(p)); !m.WarnNotValid(e, arg[i+1]) {
					if n, e := io.Copy(p, f); !m.WarnNotValid(e, arg[i+1]) {
						m.Logs(nfs.LOAD, nfs.FILE, arg[i+1], nfs.SIZE, n)
					}
				}
			}
		} else {
			mp.WriteField(arg[i], arg[i+1])
		}
	}
	return mp.FormDataContentType(), buf
}
func _spide_head(m *ice.Message, req *http.Request, head ice.Maps, value ice.Map) {
	m.Logs(req.Method, req.URL.String())
	kit.For(head, func(k, v string) { req.Header.Set(k, v) })
	kit.For(value[SPIDE_HEADER], func(k string, v string) { req.Header.Set(k, v) })
	kit.For(value[SPIDE_COOKIE], func(k string, v string) { req.AddCookie(&http.Cookie{Name: k, Value: v}) })
	kit.For(kit.Simple(m.Optionv(SPIDE_COOKIE)), func(k, v string) { req.AddCookie(&http.Cookie{Name: k, Value: v}) })
	kit.For(kit.Simple(m.Optionv(SPIDE_HEADER)), func(k, v string) { req.Header.Set(k, v) })
	kit.If(req.Method == http.MethodPost, func() {
		m.Logs(kit.Select(ice.AUTO, req.Header.Get(html.ContentLength)), req.Header.Get(html.ContentType))
	})
}
func _spide_send(m *ice.Message, name string, req *http.Request, timeout string) (*http.Response, error) {
	client := mdb.HashSelectTarget(m, name, func() ice.Any { return &http.Client{Timeout: kit.Duration(timeout)} }).(*http.Client)
	return client.Do(req)
}
func _spide_save(m *ice.Message, action, file, uri string, res *http.Response) {
	if action == SPIDE_RAW {
		m.SetResult()
	} else {
		m.SetResult().SetAppend()
	}
	switch action {
	case SPIDE_RAW:
		b, _ := ioutil.ReadAll(res.Body)
		m.Echo(string(b))
	case SPIDE_MSG:
		var data map[string][]string
		m.Assert(json.NewDecoder(res.Body).Decode(&data))
		kit.For(data[ice.MSG_APPEND], func(k string) { kit.For(data[k], func(v string) { m.Push(k, v) }) })
		m.Resultv(data[ice.MSG_RESULT])
	case SPIDE_SAVE:
		_cache_download(m, res, file, m.OptionCB(SPIDE))
	case SPIDE_CACHE:
		m.Cmdy(CACHE, DOWNLOAD, res.Header.Get(html.ContentType), uri, kit.Dict(RESPONSE, res), m.OptionCB(SPIDE))
		m.Echo(m.Append(mdb.HASH))
	default:
		var data ice.Any
		if b, e := ioutil.ReadAll(res.Body); !m.WarnNotFound(e) {
			if json.Unmarshal(b, &data) == nil {
				m.Push("", kit.KeyValue(ice.Map{}, "", m.Optionv(SPIDE_RES, data)))
			} else {
				m.Echo(string(b))
			}
		}
	}
}

const (
	SPIDE_RAW   = "raw"
	SPIDE_MSG   = "msg"
	SPIDE_SAVE  = "save"
	SPIDE_CACHE = "cache"

	SPIDE_BODY = "body"
	SPIDE_FORM = "form"
	SPIDE_PART = "part"
	SPIDE_FILE = "file"
	SPIDE_DATA = "data"
	SPIDE_JSON = "json"
	SPIDE_RES  = "content_data"

	IMAGE_JPEG = "image/jpeg"
	IMAGE_PNG  = "image/png"
	TEXT_HTML  = "text/html"
	TEXT_CSS   = "text/css"
)
const (
	SPIDE_CLIENT = "client"
	SPIDE_METHOD = "method"
	SPIDE_COOKIE = "cookie"
	SPIDE_HEADER = "header"

	CLIENT_URL      = "client.url"
	CLIENT_NAME     = "client.name"
	CLIENT_TYPE     = "client.type"
	CLIENT_METHOD   = "client.method"
	CLIENT_ORIGIN   = "client.origin"
	CLIENT_TIMEOUT  = "client.timeout"
	CLIENT_PROTOCOL = "client.protocol"
	CLIENT_HOSTNAME = "client.hostname"
	CLIENT_HOST     = "client.host"

	OPEN   = "open"
	MAIN   = "main"
	FULL   = "full"
	LINK   = "link"
	MERGE  = "merge"
	VENDOR = "vendor"

	QS = "?"
)

var agentIcons = map[string]string{
	html.Safari:         "usr/icons/Safari.png",
	html.Chrome:         "usr/icons/Chrome.png",
	html.Edg:            "usr/icons/Edg.png",
	html.MicroMessenger: "usr/icons/wechat.png",
	"Go-http-client":    "usr/icons/go.png",
}

const SPIDE = "spide"

func init() {
	Index.MergeCommands(ice.Commands{
		// SPIDE: {Name: "spide client.name action=raw,msg,save,cache method=GET,PUT,POST,DELETE url format=form,part,json,data,file arg run create", Help: "蜘蛛侠", Actions: ice.MergeActions(ice.Actions{
		SPIDE: {Help: "蜘蛛侠", Meta: kit.Dict(ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
			CLIENT_TYPE, "类型", CLIENT_NAME, "名称", CLIENT_URL, "地址",
			CLIENT_METHOD, "方法", CLIENT_ORIGIN, "服务", CLIENT_TIMEOUT, "超时",
			CLIENT_PROTOCOL, "协议", CLIENT_HOST, "主机", CLIENT_HOSTNAME, "机器",
		))), Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				conf := mdb.Confm(m, cli.RUNTIME, cli.CONF)
				dev := kit.Select("https://2021.shylinux.com", ice.Info.Make.Domain, conf[cli.CTX_DEV])
				m.Cmd("", mdb.CREATE, ice.SHY, kit.Select("https://shylinux.com", conf[cli.CTX_SHY]), nfs.REPOS)
				m.Cmd("", mdb.CREATE, ice.DEV, dev, nfs.REPOS, ice.SRC_MAIN_ICO)
				m.Cmd("", mdb.CREATE, ice.DEV_IP, kit.Select(dev, os.Getenv("ctx_dev_ip")))
				m.Cmd("", mdb.CREATE, ice.OPS, kit.Select("http://localhost:9020", conf[cli.CTX_OPS]), nfs.REPOS, nfs.USR_ICONS_CONTEXTS)
				m.Cmd("", mdb.CREATE, ice.DEMO, kit.Select("http://localhost:20000", conf[cli.CTX_DEMO]), "", nfs.USR_ICONS_VOLCANOS)
				m.Cmd("", mdb.CREATE, ice.MAIL, kit.Select("https://mail.shylinux.com", conf[cli.CTX_MAIL]), "", "usr/icons/Mail.png")
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.IsSearchPreview(m, arg) {
					mdb.HashSelectValue(m.Spawn(), func(value ice.Map) {
						m.PushSearch(mdb.TYPE, LINK, mdb.NAME, kit.Value(value, CLIENT_NAME), mdb.TEXT, kit.Value(value, CLIENT_ORIGIN), value)
					})
				}
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(ctx.ACTION) {
				case COOKIE:
					switch arg[0] {
					case mdb.KEY:
						m.Push(arg[0], ice.MSG_SESSID)
					}
				case HEADER:
					switch arg[0] {
					case mdb.KEY:
						m.Push(arg[0], html.Authorization)
					}
				case CLIENT_NAME:
					mdb.HashSelectValue(m.Spawn(), func(value ice.Map) {
						m.Push(arg[0], kit.Value(value, CLIENT_NAME))
						m.Push(mdb.TYPE, kit.Value(value, CLIENT_TYPE))
					})
					m.Sort(arg[0])
				default:
					switch arg[0] {
					case mdb.ICON, mdb.ICONS:
						mdb.HashInputs(m, arg)
					default:
						mdb.HashSelectValue(m.Spawn(), func(value ice.Map) {
							m.Push(kit.Select(ORIGIN, arg, 0), kit.Value(value, kit.Keys("client", arg[0])))
						})
						kit.If(arg[0] == mdb.TYPE, func() { m.Push(arg[0], nfs.REPOS) })
					}
				}
			}},
			mdb.CREATE: {Name: "create name origin* type icons token", Hand: func(m *ice.Message, arg ...string) {
				_spide_create(m, m.Option(mdb.NAME), m.Option(ORIGIN), m.Option(mdb.TYPE), m.OptionDefault(mdb.ICONS, nfs.USR_ICONS_VOLCANOS), m.Option(TOKEN))
			}},
			COOKIE: {Name: "cookie key* value", Help: "状态量", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(CLIENT_NAME), kit.Keys(COOKIE, m.Option(mdb.KEY)), m.Option(mdb.VALUE))
			}},
			HEADER: {Name: "header key* value", Help: "请求头", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(CLIENT_NAME), kit.Keys(HEADER, m.Option(mdb.KEY)), m.Option(mdb.VALUE))
			}},
			MERGE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.MergeURL2(m.Cmdv("", arg[0], CLIENT_URL), arg[1], arg[2:]))
			}},
			PROXY: {Name: "proxy url size cache upload", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPIDE, ice.DEV, SPIDE_RAW, http.MethodPost, m.Option(URL), SPIDE_PART, arg[2:])
			}},
			"disconn": {Help: "断连", Icon: "bi bi-person-x", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, cli.CLOSE, kit.Dict(mdb.NAME, m.Option(CLIENT_NAME)))
				mdb.HashModify(m, mdb.NAME, m.Option(CLIENT_NAME), TOKEN, "")
			}},
			DEV_REQUEST_TEXT: {Hand: func(m *ice.Message, arg ...string) { m.Echo(SpaceName(ice.Info.NodeName)) }},
			DEV_CREATE_TOKEN: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, tcp.DIAL, ice.DEV, m.Option(CLIENT_NAME), m.OptionSimple(TOKEN)).Sleep300ms()
			}},
		}, DevTokenAction(CLIENT_NAME, CLIENT_URL), mdb.ImportantHashAction(mdb.SHORT, CLIENT_NAME, mdb.FIELD, "time,icons,client.name,client.url,client.type,token")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 || arg[0] == "" || (len(arg) > 3 && arg[3] == "") {
				list := m.CmdMap(SPACE, mdb.NAME)
				mdb.HashSelect(m, kit.Slice(arg, 0, 1)...).Sort("client.type,client.name", []string{nfs.REPOS, ""})
				m.RewriteAppend(func(value, key string, index int) string {
					kit.If(key == CLIENT_URL, func() { value = kit.MergeURL(value, m.OptionSimple(ice.MSG_DEBUG)) })
					return value
				})
				m.Table(func(value ice.Maps) {
					if value[CLIENT_TYPE] == nfs.REPOS {
						if _, ok := list[value[CLIENT_NAME]]; ok {
							m.Push(mdb.STATUS, ONLINE).PushButton("disconn", mdb.DEV_REQUEST, mdb.REMOVE)
						} else {
							m.Push(mdb.STATUS, "").PushButton(mdb.DEV_REQUEST, mdb.REMOVE)
						}
					} else {
						m.Push(mdb.STATUS, "").PushButton(mdb.REMOVE)
					}
				})
				kit.If(len(arg) > 0 && arg[0] != "", func() { m.Action(COOKIE, HEADER) })
			} else {
				_spide_show(m, arg[0], arg[1:]...)
			}
		}},
		http.MethodGet: {Name: "GET url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodGet, arg[0], arg[1:]))))
		}},
		http.MethodPut: {Name: "PUT url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodPut, arg[0], arg[1:]))))
		}},
		http.MethodPost: {Name: "POST url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodPost, arg[0], arg[1:]))))
		}},
		http.MethodDelete: {Name: "DELETE url key value run", Help: "蜘蛛侠", Hand: func(m *ice.Message, arg ...string) {
			m.Echo(kit.Formats(kit.UnMarshal(m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodDelete, arg[0], arg[1:]))))
		}},
	})
	nfs.TemplateText = func(m *ice.Message, p string) string {
		if p := kit.Select(nfs.TemplatePath(m, path.Base(p)), m.Option("_template")); kit.HasPrefix(p, "/require/", ice.HTTP) {
			return m.Cmdx(SPIDE, ice.OPS, SPIDE_RAW, http.MethodGet, p)
		} else if p == "" {
			return ""
		} else {
			return m.Cmdx(nfs.CAT, p)
		}
	}
	nfs.TemplatePath = func(m *ice.Message, arg ...string) string {
		if p := path.Join(ice.SRC_TEMPLATE, m.PrefixKey(), path.Join(arg...)); nfs.Exists(m, p) {
			return p + kit.Select("", nfs.PS, len(arg) == 0)
		} else {
			p := m.FileURI(ctx.GetCmdFile(m, m.PrefixKey()))
			if p := strings.TrimPrefix(path.Join(path.Dir(p), path.Join(arg...)), "/require/"); nfs.Exists(m, p) {
				return p
			}
			if ice.Info.Important {
				return kit.MergeURL2(SpideOrigin(m, ice.OPS)+p, path.Join(arg...))
			}
			return ""
		}
	}
	nfs.DocumentPath = func(m *ice.Message, arg ...string) string {
		if p := path.Join(nfs.USR_LEARNING_PORTAL, m.PrefixKey(), path.Join(arg...)); nfs.Exists(m, p) {
			return p + kit.Select("", nfs.PS, len(arg) == 0)
		} else {
			return kit.MergeURL2(UserHost(m)+ctx.GetCmdFile(m, m.PrefixKey()), path.Join(arg...))
		}
	}
	nfs.DocumentText = func(m *ice.Message, p string) string {
		if p := nfs.DocumentPath(m, path.Base(p)); kit.HasPrefix(p, "/require/", ice.HTTP) {
			return m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodGet, p)
		} else {
			return m.Cmdx(nfs.CAT, p)
		}
	}
}

func HostPort(m *ice.Message, host, port string, arg ...string) string {
	p := ""
	if len(arg) > 0 {
		kit.If(kit.Select("", arg, 0), func(pod string) { p += S(pod) })
		kit.If(kit.Select("", arg, 1), func(cmd string) { p += C(cmd) })
	}
	kit.If(host == "", func() { host = kit.ParseURL(UserHost(m)).Hostname() })
	if port == tcp.PORT_443 {
		return kit.Format("https://%s", host) + p
	} else if port == tcp.PORT_80 {
		return kit.Format("http://%s", host) + p
	} else if port == "" {
		return kit.Format("%s://%s", UserWeb(m).Scheme, host) + p
	} else {
		return kit.Format("http://%s:%s", host, port) + p
	}
}
func PublicIP(m *ice.Message) ice.Any {
	return SpideGet(m, "http://ip-api.com/json")
}
func SpideGet(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodGet, arg))
}
func SpidePut(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodPut, arg))
}
func SpidePost(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodPost, arg))
}
func SpideDelete(m *ice.Message, arg ...ice.Any) ice.Any {
	return kit.UnMarshal(m.Cmdx(http.MethodDelete, arg))
}
func SpideSave(m *ice.Message, file, link string, cb func(count, total, value int)) *ice.Message {
	for _, p := range []string{ice.DEV_IP, ice.DEV} {
		msg := m.Cmd(Prefix(SPIDE), p, SPIDE_SAVE, file, http.MethodGet, link, cb)
		if !msg.IsErr() {
			return msg
		}
	}
	return m
}
func SpideCache(m *ice.Message, link string) *ice.Message {
	return m.Cmd(Prefix(SPIDE), ice.DEV_IP, SPIDE_CACHE, http.MethodGet, link)
}
func SpideOrigin(m *ice.Message, name string) string { return m.Cmdv(SPIDE, name, CLIENT_ORIGIN) }
func SpideURL(m *ice.Message, name string) string    { return m.Cmdv(SPIDE, name, CLIENT_URL) }
