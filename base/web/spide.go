package web

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
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
	kit "shylinux.com/x/toolkits"
)

func _spide_create(m *ice.Message, name, link string) {
	if u, e := url.Parse(link); !m.Warn(e != nil || link == "", ice.ErrNotValid, link) {
		dir, file := path.Split(u.EscapedPath())
		m.Logs(mdb.INSERT, SPIDE, name, LINK, link)
		mdb.HashSelectUpdate(m, mdb.HashCreate(m, CLIENT_NAME, name), func(value ice.Map) {
			value[SPIDE_CLIENT] = kit.Dict(mdb.NAME, name, SPIDE_METHOD, http.MethodGet, URL, link, ORIGIN, u.Scheme+"://"+u.Host,
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
	if m.Option("_break") == ice.TRUE {
		return
	}
	if c, ok := body.(io.Closer); ok {
		defer c.Close()
	}
	req, e := http.NewRequest(method, kit.MergeURL2(msg.Append(CLIENT_URL), uri, arg), body)
	if m.Warn(e, ice.ErrNotValid, uri) {
		return
	}
	mdb.HashSelectDetail(m, name, func(value ice.Map) { _spide_head(m, req, head, value) })
	if m.Option(log.DEBUG) == ice.TRUE {
		kit.For(req.Header, func(k string, v []string) { m.Logs(REQUEST, k, v) })
	}
	res, e := _spide_send(m, name, req, kit.Format(m.OptionDefault(CLIENT_TIMEOUT, msg.Append(CLIENT_TIMEOUT))))
	if m.Warn(e, ice.ErrNotFound, uri) {
		return
	}
	defer res.Body.Close()
	m.Cost(cli.STATUS, res.Status, nfs.SIZE, kit.FmtSize(kit.Int64(res.Header.Get(ContentLength))), mdb.TYPE, res.Header.Get(ContentType))
	m.Push(mdb.TYPE, STATUS).Push(mdb.NAME, res.StatusCode).Push(mdb.VALUE, res.Status)
	kit.For(res.Header, func(k string, v []string) {
		if m.Option(log.DEBUG) == ice.TRUE {
			m.Logs(RESPONSE, k, v)
		}
		m.Push(mdb.TYPE, SPIDE_HEADER).Push(mdb.NAME, k).Push(mdb.VALUE, v[0])
	})
	mdb.HashSelectUpdate(m, name, func(value ice.Map) {
		kit.For(res.Cookies(), func(v *http.Cookie) {
			kit.Value(value, kit.Keys(SPIDE_COOKIE, v.Name), v.Value)
			if m.Option(log.DEBUG) == ice.TRUE {
				m.Logs(RESPONSE, v.Name, v.Value)
			}
			m.Push(mdb.TYPE, COOKIE).Push(mdb.NAME, v.Name).Push(mdb.VALUE, v.Value)
		})
	})
	if m.Warn(res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated, ice.ErrNotValid, uri, cli.STATUS, res.Status) {
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
		_data := kit.JoinKV("=", "&", arg[1:]...)
		head[ContentType], body = ApplicationForm, bytes.NewBufferString(_data)
	case SPIDE_PART:
		head[ContentType], body = _spide_part(m, arg...)
	case SPIDE_FILE:
		if f, e := nfs.OpenFile(m, arg[1]); m.Assert(e) {
			m.Logs(nfs.LOAD, nfs.FILE, arg[1])
			body = f
		}
	case SPIDE_DATA:
		head[ContentType], body = ApplicationJSON, bytes.NewBufferString(kit.Select("{}", arg, 1))
	case SPIDE_JSON:
		arg = arg[1:]
		fallthrough
	default:
		data := ice.Map{}
		kit.For(arg, func(k, v string) { kit.Value(data, k, v) })
		_data := kit.Format(data)
		head[ContentType], body = ApplicationJSON, bytes.NewBufferString(_data)
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
			if t, e := time.ParseInLocation(ice.MOD_TIME, arg[i+1], time.Local); !m.Warn(e, ice.ErrNotValid) {
				cache = t
			}
		} else if strings.HasPrefix(arg[i+1], mdb.AT) {
			p := arg[i+1][1:]
			if s, e := nfs.StatFile(m, p); !m.Warn(e, ice.ErrNotValid) {
				if s.Size() == size && s.ModTime().Before(cache) {
					m.Option("_break", ice.TRUE)
					continue
				} else if s.Size() == size && !nfs.Exists(m.Spawn(kit.Dict(ice.MSG_FILES, nfs.DiskFile)), p) {
					m.Option("_break", ice.TRUE)
					continue
				}
				m.Logs(nfs.FIND, LOCAL, s.ModTime(), nfs.SIZE, s.Size(), CACHE, cache, nfs.SIZE, size)
			}
			if f, e := nfs.OpenFile(m, p); !m.Warn(e, ice.ErrNotValid, arg[i+1]) {
				defer f.Close()
				if p, e := mp.CreateFormFile(arg[i], path.Base(p)); !m.Warn(e, ice.ErrNotValid, arg[i+1]) {
					if n, e := io.Copy(p, f); !m.Warn(e, ice.ErrNotValid, arg[i+1]) {
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
	kit.If(req.Method == http.MethodPost, func() { m.Logs(kit.Select(ice.AUTO, req.Header.Get(ContentLength)), req.Header.Get(ContentType)) })
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
		if b, _ := ioutil.ReadAll(res.Body); strings.HasPrefix(res.Header.Get(ContentType), ApplicationJSON) {
			// m.Echo(kit.Formats(kit.UnMarshal(string(b))))
			m.Echo(string(b))
		} else {
			m.Echo(string(b))
		}
	case SPIDE_MSG:
		var data map[string][]string
		m.Assert(json.NewDecoder(res.Body).Decode(&data))
		kit.For(data[ice.MSG_APPEND], func(k string) { kit.For(data[k], func(v string) { m.Push(k, v) }) })
		m.Resultv(data[ice.MSG_RESULT])
	case SPIDE_SAVE:
		_cache_download(m, res, file, m.OptionCB(SPIDE))
	case SPIDE_CACHE:
		m.Cmdy(CACHE, DOWNLOAD, res.Header.Get(ContentType), uri, kit.Dict(RESPONSE, res), m.OptionCB(SPIDE))
		m.Echo(m.Append(mdb.HASH))
	default:
		var data ice.Any
		if b, e := ioutil.ReadAll(res.Body); !m.Warn(e) {
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

	Basic          = "Basic"
	Bearer         = "Bearer"
	Authorization  = "Authorization"
	AcceptLanguage = "Accept-Language"
	ContentLength  = "Content-Length"
	ContentType    = "Content-Type"
	UserAgent      = "User-Agent"
	Referer        = "Referer"
	Accept         = "Accept"
	Mozilla        = "Mozilla"

	ApplicationForm  = "application/x-www-form-urlencoded"
	ApplicationOctet = "application/octet-stream"
	ApplicationJSON  = "application/json"

	IMAGE_PNG = "image/png"
	TEXT_HTML = "text/html"
	TEXT_CSS  = "text/css"
)
const (
	SPIDE_CLIENT = "client"
	SPIDE_METHOD = "method"
	SPIDE_COOKIE = "cookie"
	SPIDE_HEADER = "header"

	CLIENT_NAME     = "client.name"
	CLIENT_METHOD   = "client.method"
	CLIENT_TIMEOUT  = "client.timeout"
	CLIENT_PROTOCOL = "client.protocol"
	CLIENT_HOSTNAME = "client.hostname"
	CLIENT_ORIGIN   = "client.origin"
	CLIENT_URL      = "client.url"

	OPEN   = "open"
	MAIN   = "main"
	FULL   = "full"
	LINK   = "link"
	MERGE  = "merge"
	VENDOR = "vendor"

	QS = "?"
)
const SPIDE = "spide"

func init() {
	nfs.TemplatePath = func(m *ice.Message, arg ...string) string {
		if p := path.Join(ice.SRC_TEMPLATE, m.PrefixKey(), path.Join(arg...)); nfs.Exists(m, p) {
			return p + kit.Select("", nfs.PS, len(arg) == 0)
		} else {
			return path.Join(path.Dir(ctx.GetCmdFile(m, m.PrefixKey())), path.Join(arg...)) + kit.Select("", nfs.PS, len(arg) == 0)
		}
	}
	nfs.TemplateText = func(m *ice.Message, p string) string {
		if p := nfs.TemplatePath(m, path.Base(p)); kit.HasPrefix(p, "/require/", ice.HTTP) {
			return m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodGet, p)
		} else {
			return m.Cmdx(nfs.CAT, p)
		}
	}
	nfs.DocumentPath = func(m *ice.Message, arg ...string) string {
		if p := path.Join(ice.SRC_DOCUMENT, m.PrefixKey(), path.Join(arg...)); nfs.Exists(m, p) {
			return p + kit.Select("", nfs.PS, len(arg) == 0)
		} else {
			return path.Join(path.Dir(ctx.GetCmdFile(m, m.PrefixKey())), path.Join(arg...)) + kit.Select("", nfs.PS, len(arg) == 0)
		}
	}
	nfs.DocumentText = func(m *ice.Message, p string) string {
		if p := nfs.DocumentPath(m, path.Base(p)); kit.HasPrefix(p, "/require/", ice.HTTP) {
			return m.Cmdx(SPIDE, ice.DEV, SPIDE_RAW, http.MethodGet, p)
		} else {
			return m.Cmdx(nfs.CAT, p)
		}
	}
	Index.MergeCommands(ice.Commands{
		// SPIDE: {Name: "spide client.name action=raw,msg,save,cache method=GET,PUT,POST,DELETE url format=form,part,json,data,file arg run create", Help: "蜘蛛侠", Actions: ice.MergeActions(ice.Actions{
		SPIDE: {Help: "蜘蛛侠", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				conf := mdb.Confm(m, cli.RUNTIME, cli.CONF)
				m.Cmd("", mdb.CREATE, ice.SHY, kit.Select(kit.Select("https://shylinux.com", ice.Info.Make.Remote), conf[cli.CTX_SHY]))
				m.Cmd("", mdb.CREATE, ice.DEV, kit.Select(kit.Select("https://contexts.com.cn", ice.Info.Make.Domain), conf[cli.CTX_DEV]))
				m.Cmd("", mdb.CREATE, ice.HUB, kit.Select("https://repos.shylinux.com", conf[cli.CTX_HUB]))
				m.Cmd("", mdb.CREATE, ice.COM, kit.Select("https://2021.shylinux.com", conf[cli.CTX_COM]))
				m.Cmd("", mdb.CREATE, ice.OPS, kit.Select("http://localhost:9020", conf[cli.CTX_OPS]))
				m.Cmd("", mdb.CREATE, ice.DEMO, kit.Select("http://localhost:20000", conf[cli.CTX_DEMO]))
				m.Cmd("", mdb.CREATE, ice.MAIL, kit.Select("https://mail.shylinux.com", conf[cli.CTX_MAIL]))
				m.Cmd("", mdb.CREATE, nfs.REPOS, kit.Select("https://repos.shylinux.com", conf[cli.CTX_HUB]))
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
						m.Push(arg[0], Authorization)
					}
				default:
					mdb.HashSelectValue(m.Spawn(), func(value ice.Map) { m.Push(kit.Select(ORIGIN, arg, 0), kit.Value(value, CLIENT_ORIGIN)) })
				}
			}},
			mdb.CREATE: {Name: "create name link", Hand: func(m *ice.Message, arg ...string) { _spide_create(m, m.Option(mdb.NAME), m.Option(LINK)) }},
			COOKIE: {Name: "cookie key* value", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(CLIENT_NAME), kit.Keys(COOKIE, m.Option(mdb.KEY)), m.Option(mdb.VALUE))
			}},
			HEADER: {Name: "header key* value", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashModify(m, m.OptionSimple(CLIENT_NAME), kit.Keys(HEADER, m.Option(mdb.KEY)), m.Option(mdb.VALUE))
			}},
			MERGE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.MergeURL2(m.Cmdv("", arg[0], CLIENT_URL), arg[1], arg[2:]))
			}},
			PROXY: {Name: "proxy url size cache upload", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPIDE, ice.DEV, SPIDE_RAW, http.MethodPost, m.Option(URL), SPIDE_PART, arg[2:])
			}},
		}, mdb.HashAction(mdb.SHORT, CLIENT_NAME, mdb.FIELD, "time,client.name,client.url")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 || arg[0] == "" || (len(arg) > 3 && arg[3] == "") {
				mdb.HashSelect(m, kit.Slice(arg, 0, 1)...).Sort(CLIENT_NAME)
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
}

func HostPort(m *ice.Message, host, port string) string {
	kit.If(host == "", func() { host = kit.ParseURL(UserHost(m)).Hostname() })
	if port == "443" {
		return kit.Format("https://%s", host)
	} else if port == "80" || port == "" {
		return kit.Format("http://%s", host)
	} else {
		return kit.Format("http://%s:%s", host, port)
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
func SpideSave(m *ice.Message, file, link string, cb func(count int, total int, value int)) *ice.Message {
	return m.Cmd(Prefix(SPIDE), ice.DEV, SPIDE_SAVE, file, http.MethodGet, link, cb)
}
func SpideOrigin(m *ice.Message, name string) string {
	return m.Cmdv("web.spide", name, CLIENT_ORIGIN)
}
