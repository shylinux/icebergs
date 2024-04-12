package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"path"
	"strings"

	"golang.org/x/crypto/ssh"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	PRIVATE = "private"
	PUBLIC  = "public"
	AUTHS   = "auths"
	PUSHS   = "pushs"
)
const RSA = "rsa"

func init() {
	const (
		SSH_AUTH_KEYS = ".ssh/authorized_keys"
		SSH_RSA_PUB   = ".ssh/id_rsa.pub"

		TITLE = "title"
		BITS  = "bits"
		KEY   = "key"
		PUB   = "pub"
	)
	aaa.Index.MergeCommands(ice.Commands{
		RSA: {Name: "rsa hash auto", Help: "密钥", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd("", PUBLIC) }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case TITLE:
					m.Push(arg[0], kit.Format("%s@%s", m.Option(ice.MSG_USERNAME), ice.Info.Hostname))
				}
			}},
			mdb.CREATE: {Name: "create bits=2048,4096 title", Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(TITLE, kit.Format("%s@%s", m.Option(ice.MSG_USERNAME), ice.Info.Hostname))
				if key, err := rsa.GenerateKey(rand.Reader, kit.Int(m.Option(BITS))); !m.WarnNotValid(err) {
					if pub, err := ssh.NewPublicKey(key.Public()); !m.WarnNotValid(err) {
						mdb.HashCreate(m, m.OptionSimple(TITLE),
							PUBLIC, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))+lex.SP+strings.TrimSpace(m.Option(TITLE)),
							PRIVATE, string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
						)
					}
				}
			}},
			mdb.EXPORT: {Name: "export key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.SAVE, kit.HomePath(m.Option(KEY)), m.Option(PRIVATE))
				m.Cmd(nfs.SAVE, kit.HomePath(m.Option(PUB)), m.Option(PUBLIC))
			}},
			mdb.IMPORT: {Name: "import key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Hand: func(m *ice.Message, arg ...string) {
				mdb.Conf(m, "", kit.Keys(mdb.HASH, path.Base(m.Option(KEY))), kit.Data(mdb.TIME, m.Time(),
					TITLE, kit.Format("%s@%s", ice.Info.Username, ice.Info.Hostname),
					PRIVATE, m.Cmdx(nfs.CAT, kit.HomePath(m.Option(KEY))),
					PUBLIC, m.Cmdx(nfs.CAT, kit.HomePath(m.Option(PUB))),
				))
			}},
			AUTHS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.CAT, kit.HomePath(SSH_AUTH_KEYS))
				kit.For(strings.Split(strings.TrimSpace(m.Results()), lex.NL), func(text string) {
					if ls := kit.Split(text, " ", " "); len(ls) > 2 {
						m.Push(mdb.TYPE, ls[0])
						m.Push(mdb.NAME, ls[2])
						m.Push(mdb.TEXT, ls[1])
					}
				})
			}},
			PUSHS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.PUSH, kit.HomePath(SSH_AUTH_KEYS), arg[0])
			}},
			PUBLIC: {Hand: func(m *ice.Message, arg ...string) {
				if !nfs.Exists(m, kit.HomePath(SSH_RSA_PUB)) {
					m.Cmd("", mdb.CREATE).Options(m.Cmd("").AppendSimple()).Cmd("", mdb.EXPORT)
				}
				m.Cmdy(nfs.CAT, kit.HomePath(SSH_RSA_PUB))
			}},
		}, mdb.HashAction(mdb.SHORT, PRIVATE, mdb.FIELD, "time,hash,title,public,private")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(mdb.EXPORT, mdb.REMOVE); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.IMPORT)
			}
		}},
	})
}
