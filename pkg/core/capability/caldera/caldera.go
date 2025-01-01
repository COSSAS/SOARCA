package caldera

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"slices"
	"time"
	"sigs.k8s.io/yaml"
	"strings"
	
	"soarca/pkg/core/capability/caldera/api/models"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

type calderaCapability struct {
	Factory ICalderaConnectionFactory
}

type Empty struct{}

const (
	calderaResult  = "__soarca_caldera_cmd_result__"
	calderaError   = "__soarca_caldera_cmd_error__"
	capabilityName = "soarca-caldera-cmd"
)

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(factory ICalderaConnectionFactory) *calderaCapability {
	if factory == nil {
		factory = calderaConnectionFactory{}
	}
	return &calderaCapability{factory}
}

func (capability *calderaCapability) GetType() string {
	return capabilityName
}

func (c *calderaCapability) Execute(
	metadata execution.Metadata,
	context capability.Context) (cacao.Variables, error) {

	command := context.Command
	target := context.Target

	connection, err := c.Factory.Create()
	if err != nil {
		log.Error("Could not create a connection to caldera")
		return cacao.NewVariables(), err
	}

	abilityId := ""
	groupName := ""
	createdAbility := false

	// parse the command and create the ability if neccesary
	if command.CommandB64 != "" {
		bytes, err := base64.StdEncoding.DecodeString(command.CommandB64)
		if err != nil {
			return cacao.NewVariables(), err
		}
		ability := ParseYamlAbility(bytes)
		abilityId, err = connection.CreateAbility(ability)
		if err != nil {
			log.Error("Could not create custom Ability")
			return cacao.NewVariables(), err
		}
		createdAbility = true
	} else {
		abilityId = strings.Replace(command.Command, "id: ", "", -1)
	}

	// parse the specific target group
	if target.Type == "security-category" && slices.Contains(target.Category, "caldera") {
		groupName = target.Name
	}

	// create an adversary
	adversaryId, err := connection.CreateAdversary(abilityId)
	if err != nil {
		log.Error("Could not create the Adversary", err)
		return cacao.NewVariables(), err
	}

	// start the operation
	operationId, err := connection.CreateOperation(groupName, adversaryId)
	if err != nil {
		log.Error("Could not start the Operation", err)
		return cacao.NewVariables(), err
	}

	// poll for operation status
	for finished, err := connection.IsOperationFinished(operationId); true; {
		if err != nil {
			log.Warn("Could not poll for operation status, retrying in 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}
		if finished {
			break
		}
		time.Sleep(3 * time.Second)
	}
	factsResponse, err := connection.RequestFacts(operationId)
	if err != nil {
		log.Error("Could not fetch Facts from Operation")
		return cacao.NewVariables(), err
	}

	// process the facts
	var facts = make(CalderaFacts)
	for _, link := range factsResponse {
		for _, fact := range link.Facts {
			facts[fmt.Sprint(link.Paw, "-", fact.Name)] = fmt.Sprint(fact.Value)
		}
	}

	// remove any artifacts
	if createdAbility {
		cleanup(connection, abilityId)
	} else {
		cleanup(connection, "")
	}

	return parseFacts(facts), nil
}

func cleanup(cc ICalderaConnection, abilityId string) {
	if abilityId != "" {
		err := cc.DeleteAbility(abilityId)
		if err != nil {
			log.Warn("Could not cleanup artifacts from command")
		}
	}
}

func parseFacts(facts CalderaFacts) cacao.Variables {
	variables := make(cacao.Variables, len(facts))
	for name, value := range facts {
		variables[name] = cacao.Variable{
			Name:  fmt.Sprintf("__%s__", name),
			Type:  cacao.VariableTypeString,
			Value: value,
		}
	}
	return variables
}

func ParseYamlAbility(bytes []byte) *models.Ability {
	var ability models.Ability
	err := yaml.Unmarshal(bytes, &ability)
	if err != nil {
		log.Error("Could not convert ability YAML to a valid Ability object")
		return &models.Ability{}
	}
	return &ability
}
