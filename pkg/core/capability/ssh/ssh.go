package ssh

import (
	"errors"
	"reflect"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"strings"
	"time"

	"soarca/internal/logger"

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
	context capability.Context) (cacao.Variables, error) {

	log.Trace(metadata.ExecutionId)
	return execute(context.Command, context.Authentication, context.Target)
}

func execute(command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget) (cacao.Variables, error) {

	err := CheckSshAuthenticationInfo(authentication)

	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	config, err := getConfig(authentication)
	if err != nil {
		return cacao.NewVariables(), err
	}
	session, client, err := getSession(config, target)
	if err != nil {
		return cacao.NewVariables(), err
	}
	defer close(client)

	return executeCommand(session, command)
}

func executeCommand(session *ssh.Session,
	command cacao.Command) (cacao.Variables, error) {

	response, err := session.Output(StripSshPrepend(command.Command))

	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	results := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString,
		Name:  sshResultVariableName,
		Value: string(response)})
	log.Trace("Finished ssh execution will return the variables: ", results)
	sessionErr := session.Close()
	if sessionErr != nil {
		log.Error(sessionErr)
	}

	return results, err
}

func getConfig(authentication cacao.AuthenticationInformation) (ssh.ClientConfig, error) {
	config := ssh.ClientConfig{User: authentication.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(time.Second * 20)}

	switch authentication.Type {
	case "user-auth":
		config.Auth = []ssh.AuthMethod{
			ssh.Password(authentication.Password)}
		return config, nil

	case "private-key":
		signer, err := ssh.ParsePrivateKey([]byte(authentication.PrivateKey))
		if err != nil || authentication.Password == "" {
			log.Error("no valid authentication information: ", err)
			return config, err
		}
		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer)}
		return config, nil

	default:
		return config, errors.New("non supported authentication type")
	}

}

func getSession(config ssh.ClientConfig, target cacao.AgentTarget) (*ssh.Session, *ssh.Client, error) {
	host := CombinePortAndAddress(target.Address, target.Port)
	client, err := ssh.Dial("tcp", host, &config)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		log.Error(err)
		close(client)
		return nil, nil, err
	}
	return session, client, err
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

	switch authentication.Type {
	case "user-auth":
		if strings.TrimSpace(authentication.Password) == "" {
			return errors.New("password is empty")
		}
	case "private-key":
		if strings.TrimSpace(authentication.PrivateKey) == "" {
			return errors.New("private key is not set")
		}
	default:
		return errors.New("non supported authentication type")
	}
	return nil
}

func close(client *ssh.Client) {
	if client != nil {
		err := client.Close()
		if err != nil {
			log.Error(err)
		}
	}
}
