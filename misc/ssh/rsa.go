package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"path"

	"golang.org/x/crypto/ssh"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	PUBLIC  = "public"
	PRIVATE = "private"
)
const RSA = "rsa"

func init() {
	const (
		TITLE = "title"
		BITS  = "bits"
		KEY   = "key"
		PUB   = "pub"
	)
	aaa.Index.MergeCommands(ice.Commands{
		RSA: {Name: "rsa hash auto", Help: "密钥", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create bits=2048,4096 title=some", Hand: func(m *ice.Message, arg ...string) {
				if key, err := rsa.GenerateKey(rand.Reader, kit.Int(m.Option(BITS))); !m.Warn(err, ice.ErrNotValid) {
					if pub, err := ssh.NewPublicKey(key.Public()); !m.Warn(err, ice.ErrNotValid) {
						mdb.HashCreate(m, m.OptionSimple(TITLE), PUBLIC, string(ssh.MarshalAuthorizedKey(pub)),
							PRIVATE, string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
						)
					}
				}
			}},
			mdb.EXPORT: {Name: "export key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m, m.Option(mdb.HASH)).Tables(func(value ice.Maps) {
					m.Cmd(nfs.SAVE, kit.HomePath(m.Option(KEY)), value[PRIVATE])
					m.Cmd(nfs.SAVE, kit.HomePath(m.Option(PUB)), value[PUBLIC])
				})
			}},
			mdb.IMPORT: {Name: "import key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Hand: func(m *ice.Message, arg ...string) {
				m.Conf("", kit.Keys(mdb.HASH, path.Base(m.Option(KEY))), kit.Data(mdb.TIME, m.Time(),
					TITLE, kit.Format("%s@%s", ice.Info.Username, ice.Info.Hostname),
					PRIVATE, m.Cmdx(nfs.CAT, kit.HomePath(m.Option(KEY))),
					PUBLIC, m.Cmdx(nfs.CAT, kit.HomePath(m.Option(PUB))),
				))
			}},
		}, mdb.HashAction(mdb.SHORT, PRIVATE, mdb.FIELD, "time,hash,title,public,private")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(mdb.EXPORT, mdb.REMOVE); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.IMPORT)
			}
		}},
	})
}
