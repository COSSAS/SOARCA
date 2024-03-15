package ssh

import (
	"errors"
	"reflect"
	"soarca/models/cacao"
	"soarca/models/execution"
	"strings"
	"time"

	"soarca/logger"

	"golang.org/x/crypto/ssh"
)

const (
	sshResultVariableName = "__soarca_ssh_result__"
	sshCapabilityName     = "soarca-ssh"
)

type SshCapability struct {
}

var component = reflect.TypeOf(SshCapability{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func (sshCapability *SshCapability) GetType() string {
	return sshCapabilityName
}

func (sshCapability *SshCapability) Execute(metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.Variables) (cacao.Variables, error) {
	log.Trace(metadata.ExecutionId)

	host := CombinePortAndAddress(target.Address, target.Port)

	errAuth := CheckSshAuthenticationInfo(authentication)

	if errAuth != nil {
		log.Error(errAuth)
		return cacao.NewVariables(), errAuth
	} else {
		log.Trace(host)
	}

	var config ssh.ClientConfig

	if authentication.Type == "user-auth" {
		config = ssh.ClientConfig{
			User: authentication.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(authentication.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Duration(time.Second * 20),
		}
	} else if authentication.Type == "private-key" {
		signer, errKey := ssh.ParsePrivateKey([]byte(authentication.PrivateKey))
		if errKey != nil || authentication.Password == "" {
			log.Error("no valid authentication information: ", errKey)
			return cacao.NewVariables(), errKey
		}
		config = ssh.ClientConfig{
			User: authentication.Username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Duration(time.Second * 20),
		}

	}

	var conn *ssh.Client

	conn, err := ssh.Dial("tcp", host, &config)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	var session *ssh.Session
	session, err = conn.NewSession()
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}

	response, err := session.Output(StripSshPrepend(command.Command))
	defer session.Close()

	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	results := cacao.NewVariables(cacao.Variable{Name: sshResultVariableName, Value: string(response)})
	log.Trace("Finished ssh execution will return the variables: ", results)
	return results, err
}

func CombinePortAndAddress(addresses map[cacao.NetAddressType][]string, port string) string {
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
