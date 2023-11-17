package ssh

import (
	"errors"
	"reflect"
	"soarca/models/cacao"
	"strings"

	"soarca/logger"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type SshCapability struct {
}

var component = reflect.TypeOf(SshCapability{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Trace, "", logger.Json)
}

func (sshCapability *SshCapability) GetType() string {
	return "ssh"
}

func (sshCapability *SshCapability) Execute(executionId uuid.UUID,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.Target,
	variables map[string]cacao.Variables) (map[string]cacao.Variables, error) {

	host := CombinePortAndAddress(target.Address, target.Port)

	errauth := CheckSshAuthenticationInfo(authentication)

	if errauth != nil {
		log.Error("No valid authentication information")
		return map[string]cacao.Variables{}, errauth
	}

	var config ssh.ClientConfig

	if authentication.Type == "user-auth" {
		config = ssh.ClientConfig{
			User: authentication.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(authentication.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else if authentication.Type == "private-key" {
		signer, errkey := ssh.ParsePrivateKey([]byte(authentication.PrivateKey))
		if errkey != nil || authentication.Password == "" {
			log.Error("No valid authentication information")
			return map[string]cacao.Variables{}, errkey
		}
		config = ssh.ClientConfig{
			User: authentication.Username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

	}

	var conn *ssh.Client

	conn, err := ssh.Dial("tcp", host, &config)
	if err != nil {
		return map[string]cacao.Variables{}, err
	}
	var session *ssh.Session
	session, err = conn.NewSession()
	if err != nil {
		return map[string]cacao.Variables{}, err
	}

	response, err := session.Output(StripSshPrepend(command.Command))
	defer session.Close()

	if err != nil {
		return map[string]cacao.Variables{"result": {Name: "result", Value: string(response)}}, err
	}
	return map[string]cacao.Variables{"result": {Name: "result", Value: string(response)}}, err
}

func CombinePortAndAddress(addresses map[string][]string, port string) string {
	if port == "" {
		port = "22"
	}
	if len(addresses) > 0 {
		if len(addresses["ipv4"]) > 0 {
			base := addresses["ipv4"][0] + ":" + port
			return base
		}
	}
	return ""

}

func StripSshPrepend(command string) string {
	split := strings.Split(command, "ssh ")
	if len(split) == 1 {
		return split[0]
	}
	return split[1]
}

func CheckSshAuthenticationInfo(authentication cacao.AuthenticationInformation) error {
	if strings.TrimSpace(authentication.Username) == "" {
		return errors.New("username is empty")
	}
	if authentication.Type == "user-auth" {
		if strings.TrimSpace(authentication.Password) == "" {
			return errors.New("password is empty")
		}
		return nil
	} else if authentication.Type == "private-key" {
		if strings.TrimSpace(authentication.PrivateKey) == "" {
			return errors.New("private key is not set")
		}
		return nil
	} else {
		return errors.New("non supported authentication type")
	}
}
