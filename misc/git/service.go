package git

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"

	git "shylinux.com/x/go-git/v5"
	"shylinux.com/x/go-git/v5/plumbing"
	"shylinux.com/x/go-git/v5/plumbing/protocol/packp"
	"shylinux.com/x/go-git/v5/plumbing/transport"
	"shylinux.com/x/go-git/v5/plumbing/transport/server"
)

func _service_path(m *ice.Message, p string, arg ...string) string {
	return kit.Path(ice.USR_LOCAL_REPOS, kit.TrimExt(p, GIT), path.Join(arg...))
}
func _service_link(m *ice.Message, p string, arg ...string) string {
	return kit.MergeURL2(web.UserHost(m), web.X(p)+_GIT)
}
func _service_param(m *ice.Message, arg ...string) (string, string) {
	repos, service := arg[0], kit.Select(arg[len(arg)-1], m.Option(SERVICE))
	return _service_path(m, repos), strings.TrimPrefix(service, "git-")
}
func _service_repos(m *ice.Message, arg ...string) error {
	repos, service := _service_param(m, arg...)
	m.Logs(m.R.Method, service, repos)
	if service == RECEIVE_PACK && m.R.Method == http.MethodPost {
		defer m.Cmd(Prefix(SERVICE), mdb.CREATE, mdb.NAME, path.Base(repos))
	}
	info := false
	if strings.HasSuffix(path.Join(arg...), INFO_REFS) {
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-advertisement", service))
		_service_writer(m, "# service=git-"+service+lex.NL)
		info = true
	} else {
		web.RenderType(m.W, "", kit.Format("application/x-git-%s-result", service))
	}
	web.Count(m, service, repos)
	if mdb.Conf(m, web.CODE_GIT_SERVICE, kit.Keym(ice.CMD)) == GIT {
		return _service_repos0(m, arg...)
	}
	// if service == RECEIVE_PACK && m.R.Method == http.MethodPost {
	// 	return _service_repos0(m, arg...)
	// }
	reader, err := _service_reader(m)
	if err != nil {
		return err
	}
	defer reader.Close()
	stream := ServerCommand{Stdin: reader, Stdout: m.W, Stderr: m.W}
	if service == RECEIVE_PACK {
		return ServeReceivePack(info, stream, repos)
	} else {
		return ServeUploadPack(info, stream, repos)
	}
}
func _service_repos0(m *ice.Message, arg ...string) error {
	repos, service := _service_param(m, arg...)
	if m.Options(cli.CMD_DIR, repos, cli.CMD_OUTPUT, m.W); strings.HasSuffix(path.Join(arg...), INFO_REFS) {
		_git_cmd(m, service, "--stateless-rpc", "--advertise-refs", nfs.PT)
		return nil
	}
	reader, err := _service_reader(m)
	if err != nil {
		return err
	}
	defer reader.Close()
	_git_cmd(m.Options(cli.CMD_INPUT, reader), service, "--stateless-rpc", nfs.PT)
	return nil
}
func _service_writer(m *ice.Message, cmd string, str ...string) {
	s := strconv.FormatInt(int64(len(cmd)+4), 16)
	kit.If(len(s)%4 != 0, func() { s = strings.Repeat("0", 4-len(s)%4) + s })
	m.W.Write([]byte(s + cmd + "0000" + strings.Join(str, "")))
}
func _service_reader(m *ice.Message) (io.ReadCloser, error) {
	switch m.R.Header.Get(html.ContentEncoding) {
	case "deflate":
		return flate.NewReader(m.R.Body), nil
	case "gzip":
		return gzip.NewReader(m.R.Body)
	}
	return m.R.Body, nil
}

const (
	INFO_REFS    = "info/refs"
	UPLOAD_PACK  = "upload-pack"
	RECEIVE_PACK = "receive-pack"
)
const SERVICE = "service"

func init() {
	web.Index.MergeCommands(ice.Commands{"/x/": {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
		if !m.IsCliUA() {
			web.RenderCmd(m, web.CODE_GIT_SERVICE, arg)
			return
		} else if len(arg) == 0 {
			return
		} else if arg[0] == ice.LIST {
			m.Cmd(Prefix(SERVICE), func(value ice.Maps) { m.Push(nfs.REPOS, _service_link(m, value[nfs.REPOS])) })
			m.Sort(nfs.REPOS)
			return
		} else if m.RenderVoid(); m.Option("go-get") == "1" {
			p := _service_link(m, path.Join(arg...))
			m.RenderResult(kit.Format(`<meta name="go-import" content="%s">`, kit.Format(`%s git %s`, strings.TrimSuffix(strings.Split(p, "://")[1], nfs.PT+GIT), p)))
			return
		}
		switch repos, service := _service_param(m, arg...); service {
		case RECEIVE_PACK:
			if !web.BasicCheck(m, "git server", func(msg *ice.Message) bool {
				return msg.Append(mdb.TYPE) == STATUS || (msg.Append(mdb.TYPE) == web.SERVER && msg.Append(mdb.TEXT) == path.Base(repos))
			}) {
				return
			} else if !nfs.Exists(m, repos) {
				m.Cmd(Prefix(SERVICE), mdb.CREATE, mdb.NAME, path.Base(repos))
			}
		case UPLOAD_PACK:
			if mdb.Conf(m, Prefix(SERVICE), kit.Keym(aaa.AUTH)) == aaa.PRIVATE {
				if !web.BasicCheck(m, "git server", func(msg *ice.Message) bool {
					return msg.Append(mdb.TYPE) == STATUS || msg.Append(mdb.TYPE) == web.SERVER && msg.Append(mdb.TEXT) == path.Base(repos)
				}) {
					return
				}
			}
			if !nfs.Exists(m, repos) {
				list := m.CmdMap(Prefix(REPOS), REPOS)
				name := path.Base(repos)
				if _repos, ok := list[name]; ok {
					m.Cmd(Prefix(SERVICE), CLONE, name, _repos[ORIGIN])
				}
			}
			if m.WarnNotFound(!nfs.Exists(m, repos), arg[0]) {
				return
			}
		}
		m.WarnNotValid(_service_repos(m, arg...))
	}}})
	Index.MergeCommands(ice.Commands{
		SERVICE: {Name: "service repos branch commit file auto", Help: "代码源", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.DIR, ice.USR_LOCAL_REPOS, func(value ice.Maps) { _repos_insert(m, value[nfs.PATH]) })
			}},
			CLONE: {Name: "clone name*=demo origin", Hand: func(m *ice.Message, arg ...string) {
				git.PlainClone(_service_path(m, m.Option(mdb.NAME)), true, &git.CloneOptions{URL: m.Option(ORIGIN), Auth: _repos_auth(m, m.Option(ORIGIN))})
				_repos_insert(m, _service_path(m, m.Option(mdb.NAME)))
			}},
			mdb.CREATE: {Name: "create name*=demo", Hand: func(m *ice.Message, arg ...string) {
				git.PlainInit(_service_path(m, m.Option(mdb.NAME)), true)
				_repos_insert(m, _service_path(m, m.Option(mdb.NAME)))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				m.Assert(m.Option(REPOS) != "")
				mdb.HashRemove(m, m.Option(REPOS))
				nfs.Trash(m, _service_path(m, m.Option(REPOS)))
			}},
			code.INNER: {Hand: func(m *ice.Message, arg ...string) { _repos_inner(m, _service_path, arg...) }},
			web.DREAM_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[0] == REPOS, func() {
					mdb.HashSelect(m).Sort(REPOS).Cut("repos,version,time")
					web.DreamListSpide(m, []string{ice.DEV}, web.ORIGIN, func(dev, origin string) {
						m.Spawn().SplitIndex(m.Cmdx(web.SPIDE, dev, web.SPIDE_RAW, http.MethodGet, web.C(web.CODE_GIT_SERVICE))).Table(func(value ice.Maps) {
							value[nfs.REPOS] = origin + web.X(value[nfs.REPOS])
							m.Push("", value, kit.Split("repos,version,time"))
						})
					})
				})
			}},
		}, web.DreamAction(), mdb.HashAction(mdb.SHORT, REPOS, mdb.FIELD, "time,repos,branch,version,message", ice.CMD, GIT), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m, arg...).Table(func(value ice.Maps) {
					m.Push(nfs.SIZE, kit.Split(m.Cmdx(cli.SYSTEM, "du", "-sh", path.Join(ice.USR_LOCAL_REPOS, value[REPOS])))[0])
					m.PushScript(kit.Format("git clone %s", _service_link(m, value[REPOS])))
				}).Sort(REPOS)
				kit.If(!m.IsCliUA(), func() { m.Cmdy(web.CODE_PUBLISH, ice.CONTEXTS, ice.DEV) })
				kit.If(mdb.Config(m, aaa.AUTH) == aaa.PRIVATE, func() { m.StatusTimeCount(aaa.AUTH, aaa.PRIVATE) })
			} else if repos := _repos_open(m, arg[0]); len(arg) == 1 {
				defer m.PushScript(kit.Format("git clone %s", _service_link(m, arg[0])))
				_repos_branch(m, repos)
			} else if len(arg) == 2 {
				if iter, err := repos.Branches(); err == nil {
					iter.ForEach(func(refer *plumbing.Reference) error {
						kit.If(refer.Name().Short() == arg[1], func() { _repos_log(m, refer.Hash(), repos) })
						return nil
					})
				}
			} else if len(arg) == 3 {
				if arg[2] == INDEX {
					_repos_status(m, arg[0], repos)
				} else {
					_repos_stats(m, repos, arg[2])
				}
			} else {
				m.Cmdy("", code.INNER, arg)
			}
		}},
	})
}

type ServerCommand struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func ServeReceivePack(info bool, srvCmd ServerCommand, path string) error {
	if ep, err := transport.NewEndpoint(path); err != nil {
		return err
	} else if s, err := server.DefaultServer.NewReceivePackSession(ep, nil); err != nil {
		return err
	} else {
		return serveReceivePack(info, srvCmd, s)
	}
}
func ServeUploadPack(info bool, srvCmd ServerCommand, path string) error {
	if ep, err := transport.NewEndpoint(path); err != nil {
		return err
	} else if s, err := server.DefaultServer.NewUploadPackSession(ep, nil); err != nil {
		return err
	} else {
		return serveUploadPack(info, srvCmd, s)
	}
}
func serveReceivePack(info bool, cmd ServerCommand, s transport.ReceivePackSession) error {
	if info {
		return serveAdvertisedRefer(cmd, s)
	}
	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(cmd.Stdin); err != nil {
		return err
	} else if resp, err := s.ReceivePack(context.TODO(), req); err != nil {
		return err
	} else {
		return resp.Encode(cmd.Stdout)
	}
}
func serveUploadPack(info bool, cmd ServerCommand, s transport.UploadPackSession) (err error) {
	if info {
		return serveAdvertisedRefer(cmd, s)
	}
	req := packp.NewUploadPackRequest()
	if err := req.Decode(cmd.Stdin); err != nil {
		return err
	} else if resp, err := s.UploadPack(context.TODO(), req); err != nil {
		return err
	} else {
		return resp.Encode(cmd.Stdout)
	}
}
func serveAdvertisedRefer(cmd ServerCommand, s transport.Session) (err error) {
	if resp, err := s.AdvertisedReferences(); err != nil {
		return err
	} else {
		return resp.Encode(cmd.Stdout)
	}
}
