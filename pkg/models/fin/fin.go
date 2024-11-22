package fin

import (
	"encoding/json"
	"soarca/pkg/models/cacao"
	"time"
)

// command constants
const (
	MessageTypeAck        = "ack"
	MessageTypeNack       = "nack"
	MessageTypeRegister   = "register"
	MessageTypeUnregister = "unregister"
	MessageTypeCommand    = "command"
	MessageTypeResult     = "result"
	MessageTypePause      = "pause"
	MessageTypeResume     = "resume"
	MessageTypeStop       = "stop"
)

// Ack
type Ack struct {
	Type      string `json:"type"`
	MessageId string `json:"message_id"`
}

// Nack
type Nack struct {
	Type      string `json:"type"`
	MessageId string `json:"message_id"`
}

// Register message structure
type Register struct {
	Type            string       `json:"type"`
	MessageId       string       `json:"message_id"`
	FinID           string       `json:"fin_id"`
	Name            string       `json:"fin_name"`
	ProtocolVersion string       `json:"protocol_version"`
	Security        Security     `json:"security"`
	Capabilities    []Capability `json:"capabilities"`
	Meta            Meta         `json:"meta,omitempty"`
}

// Capability register message substructure
type Capability struct {
	Id      string                       `json:"capability_id"`
	Name    string                       `json:"name"`
	Version string                       `json:"version"`
	Step    map[string]Step              `json:"step,omitempty"`
	Agent   map[string]cacao.AgentTarget `json:"agent,omitempty"`
}

// Step structure as example to the executor
type Step struct {
	Type               string                     `json:"type"`
	Name               string                     `json:"name"`
	Description        string                     `json:"description"`
	ExternalReferences []cacao.ExternalReferences `json:"external_references"`
	Command            string                     `json:"command"`
	Target             string                     `json:"target"`
}

// Unregister command structure
type Unregister struct {
	Type      string `json:"type"`
	MessageId string `json:"message_id"`
	Id        string `json:"capability_id"`
	FinID     string `json:"fin_id"`
	All       string `json:"all"`
}

// Command
type Command struct {
	Type                string              `json:"type"`
	MessageId           string              `json:"message_id"`
	CommandSubstructure CommandSubstructure `json:"command"`
	Meta                Meta                `json:"meta"`
}

// Command substructure used by the command message
type CommandSubstructure struct {
	Command        string                          `json:"command"`
	Authentication cacao.AuthenticationInformation `json:"authentication"`
	Context        Context                         `json:"context"`
	Variables      map[string]cacao.Variable       `json:"variables"`
}

// Result message structure
type Result struct {
	Type            string          `json:"type"`
	MessageId       string          `json:"message_id"`
	ResultStructure ResultStructure `json:"result"`
	Meta            Meta            `json:"meta"`
}

// Result substructure used by the result message
type ResultStructure struct {
	State     string                    `json:"state"`
	Context   Context                   `json:"context"`
	Variables map[string]cacao.Variable `json:"variables"`
}

// Control message structure
type Control struct {
	Type         string `json:"type"`
	MessageId    string `json:"message_id"`
	CapabilityId string `json:"capability_id"`
}

// Status message structure
type Status struct {
	Type         string `json:"type"`
	MessageId    string `json:"message_id"`
	CapabilityId string `json:"capability_id"`
	Progress     string `json:"progress"`
}

// Security message substructure
type Security struct {
	Version         string `json:"version"`
	ChannelSecurity string `json:"channel_security"`
}

// Context message substructure
type Context struct {
	CompletedOn time.Time `json:"completed_on"`
	GeneratedOn time.Time `json:"generated_on"`
	Timeout     int       `json:"timeout"`
	Delay       int       `json:"delay"`
	StepId      string    `json:"step_id"`
	PlaybookId  string    `json:"playbook_id"`
	ExecutionId string    `json:"execution_id"`
}

// Meta message substructure
type Meta struct {
	Timestamp time.Time `json:"timestamp"`
	SenderId  string    `json:"sender_id"`
}

type Message struct {
	Type      string `json:"type"`
	MessageId string `json:"message_id"`
}

func NewCommand() Command {
	instance := Command{}
	instance.Type = MessageTypeCommand
	instance.CommandSubstructure.Context.Timeout = 1
	//instance.CommandSubstructure.Context.GeneratedOn = time.Now()

	return instance
}

func NewAck(messageId string) Ack {
	ack := Ack{Type: MessageTypeAck, MessageId: messageId}
	return ack
}

func NewNack(messageId string) Nack {
	nack := Nack{Type: MessageTypeNack, MessageId: messageId}
	return nack
}

func Decode(data []byte, object any) error {
	return json.Unmarshal(data, object)
}

func Encode(object any) ([]byte, error) {
	return json.Marshal(object)
}
