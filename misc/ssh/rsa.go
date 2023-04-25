package ssh

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"path"

	"golang.org/x/crypto/ssh"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	PUBLIC  = "public"
	PRIVATE = "private"
	VERIFY  = "verify"
	SIGN    = "sign"
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
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case TITLE:
					m.Push(arg[0], kit.Format("%s@%s", m.Option(ice.MSG_USERNAME), ice.Info.Hostname))
				}
			}},
			mdb.CREATE: {Name: "create bits=2048,4096 title=some", Hand: func(m *ice.Message, arg ...string) {
				if key, err := rsa.GenerateKey(rand.Reader, kit.Int(m.Option(BITS))); !m.Warn(err, ice.ErrNotValid) {
					if pub, err := ssh.NewPublicKey(key.Public()); !m.Warn(err, ice.ErrNotValid) {
						mdb.HashCreate(m, m.OptionSimple(TITLE), PUBLIC, string(ssh.MarshalAuthorizedKey(pub))+lex.SP+m.Option(TITLE),
							PRIVATE, string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
						)
					}
				}
			}},
			mdb.EXPORT: {Name: "export key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m, m.Option(mdb.HASH)).Table(func(value ice.Maps) {
					m.Cmd(nfs.SAVE, kit.HomePath(m.Option(KEY)), value[PRIVATE])
					m.Cmd(nfs.SAVE, kit.HomePath(m.Option(PUB)), value[PUBLIC])
				})
			}},
			mdb.IMPORT: {Name: "import key=.ssh/id_rsa pub=.ssh/id_rsa.pub", Hand: func(m *ice.Message, arg ...string) {
				mdb.Conf(m, "", kit.Keys(mdb.HASH, path.Base(m.Option(KEY))), kit.Data(mdb.TIME, m.Time(),
					TITLE, kit.Format("%s@%s", ice.Info.Username, ice.Info.Hostname),
					PRIVATE, m.Cmdx(nfs.CAT, kit.HomePath(m.Option(KEY))),
					PUBLIC, m.Cmdx(nfs.CAT, kit.HomePath(m.Option(PUB))),
				))
			}},
			SIGN: {Hand: func(m *ice.Message, arg ...string) {
				if !nfs.Exists(m, "etc/id_rsa") {
					if key, err := rsa.GenerateKey(rand.Reader, kit.Int("2048")); !m.Warn(err, ice.ErrNotValid) {
						m.Cmd(nfs.SAVE, "etc/id_rsa", string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})))
						m.Cmd(nfs.SAVE, "etc/id_rsa.pub", string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(key.Public().(*rsa.PublicKey))})))
					}
				}
				block, _ := pem.Decode([]byte(m.Cmdx(nfs.CAT, "etc/id_rsa")))
				key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				if m.Warn(err) {
					return
				}
				hash := sha256.New()
				if _, err := hash.Write([]byte(arg[0])); m.Warn(err) {
					return
				}
				signature, err := rsa.SignPSS(rand.Reader, key, crypto.SHA256, hash.Sum(nil), nil)
				if m.Warn(err) {
					return
				}
				m.Echo(hex.EncodeToString(signature))
			}},
			VERIFY: {Hand: func(m *ice.Message, arg ...string) {
				block, _ := pem.Decode([]byte(m.Cmdx(nfs.CAT, "etc/id_rsa.pub")))
				pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
				if m.Warn(err) {
					return
				}
				signature, err := hex.DecodeString(arg[1])
				if m.Warn(err) {
					return
				}
				hash := sha256.New()
				if _, err := hash.Write([]byte(arg[0])); m.Warn(err) {
					return
				}
				if !m.Warn(rsa.VerifyPSS(pub, crypto.SHA256, hash.Sum(nil), signature, nil)) {
					m.Echo(ice.OK)
				}
			}},
		}, mdb.HashAction(mdb.SHORT, PRIVATE, mdb.FIELD, "time,hash,title,public,private")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(mdb.EXPORT, mdb.REMOVE); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.IMPORT)
			}
		}},
	})
}
