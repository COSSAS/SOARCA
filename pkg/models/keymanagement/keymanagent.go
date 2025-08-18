package keymanagement

import "golang.org/x/crypto/ssh"

type KeyPair struct {
	Public  ssh.PublicKey
	Private ssh.Signer
}
