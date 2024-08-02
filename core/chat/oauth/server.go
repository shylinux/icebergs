package oauth

import "shylinux.com/x/ice"

type Server struct{ ice.Hash }

func init() { ice.Cmd("web.chat.oauth.server", Server{}) }
