package cacao

import (
	"encoding/json"
	"time"
)

type Metadata struct{}

const (
	StepTypeEnd             = "end"
	StepTypeStart           = "start"
	StepTypeAction          = "action"
	StepTypePlaybookAction  = "playbook-action"
	StepTypeParallel        = "parallel"
	StepTypeIfCondition     = "if-condition"
	StepTypeWhileCondition  = "while-condition"
	StepTypeSwitchCondition = "switch-condition"

	CommandTypeManual     = "manual"
	CommandTypeBash       = "bash"
	CommandTypeCalderaCmd = "caldera-cmd"
	CommandTypeElastic    = "elastic"
	CommandTypeHttpApi    = "http-api"
	CommandTypeJupyter    = "jupyter"
	CommandTypeKestrel    = "kestrel"
	CommandTypeOpenC2Http = "openc2-http"
	CommandTypePowershell = "powershell"
	CommandTypeSigma      = "sigma"
	CommandTypeSsh        = "ssh"
	CommandTypeYara       = "yara"

	AuthInfoOAuth2Type    = "oauth2"
	AuthInfoHTTPBasicType = "http-basic"
	AuthInfoNotSet        = ""
	CACAO_VERSION_1       = "cacao-1.0"
	CACAO_VERSION_2       = "cacao-2.0"
)

// Custom type intended for AgentTarget.Address dict keys
// not used at the moment as it would break a few things
type NetAddressType string

const (
	DName NetAddressType = "dname"
	IPv4  NetAddressType = "ipv4"
	IPv6  NetAddressType = "ipv6"
	L2Mac NetAddressType = "l2mac"
	VLan  NetAddressType = "vlan"
	Url   NetAddressType = "url"
)

type (
	Extensions                 map[string]interface{}
	Workflow                   map[string]Step
	Variables                  map[string]Variable
	Addresses                  map[NetAddressType][]string
	AgentTargets               map[string]AgentTarget
	AuthenticationInformations map[string]AuthenticationInformation
	ExtensionDefinitions       map[string]ExtensionDefinition
	Headers                    map[string][]string
	Cases                      map[string]string
	DataMarkings               map[string]DataMarking
)

// CACAO Variable
type Variable struct {
	Type        string `bson:"type" json:"type" validate:"required" example:"string"`                    // Type of the variable should be OASIS  variable-type-ov
	Name        string `bson:"name,omitempty" json:"name,omitempty" example:"__example_string__"`        // The name of the variable in the style __variable_name__ (not part of CACAO spec, but included in object for utility)
	Description string `bson:"description,omitempty" json:"description,omitempty" example:"some string"` // A description of the variable
	Value       string `bson:"value,omitempty" json:"value,omitempty" example:"this is a value"`         // The value of the that the variable will evaluate to
	Constant    bool   `bson:"constant,omitempty" json:"constant,omitempty" example:"false"`             // Indicate if it's a constant
	External    bool   `bson:"external,omitempty" json:"external,omitempty" example:"false"`             // Indicate if it's external
}

const (
	VariableTypeBool        = "bool"
	VariableTypeDictionary  = "dictionary"
	VariableTypeFloat       = "float"
	VariableTypeHexString   = "hexstring"
	VariableTypeInt         = "integer"
	VariableTypeIpv4Address = "ipv4-addr"
	VariableTypeIpv6Address = "ipv6-addr"
	VariableTypeLong        = "long"
	VariableTypeMacAddress  = "mac-addr"
	VariableTypeHash        = "hash"
	VariableTypeMd5Has      = "md5-hash"
	VariableTypeSha256      = "sha256-hash"
	VariableTypeString      = "string"
	VariableTypeUri         = "uri"
	VariableTypeUuid        = "uuid"
)

type Playbook struct {
	ID                            string                     `bson:"_id" json:"id" validate:"required" example:"playbook--77c4c428-6304-4950-93ff-83c5fd4cb67a"`                                      // Used by SOARCA so refer to the object while loading it from the database
	Type                          string                     `bson:"type" json:"type" validate:"required" example:"playbook"`                                                                         // Must be playbook
	SpecVersion                   string                     `bson:"spec_version" json:"spec_version" validate:"required" example:"cacao-2.0"`                                                        // Indicate the specification version cacao-2.0 is the only supported version at this time
	Name                          string                     `bson:"name" json:"name" validate:"required" example:"Investigation playbook"`                                                           // An indicative name of the playbook
	Description                   string                     `bson:"description,omitempty" json:"description,omitempty" example:"This is an example investigation playbook"`                          // A descriptive text to indicate what your playbook does
	PlaybookTypes                 []string                   `bson:"playbook_types,omitempty" json:"playbook_types,omitempty" example:"investigation"`                                                // Should be of the CACAO playbook-type-ov
	CreatedBy                     string                     `bson:"created_by" json:"created_by"  validate:"required" example:"identity--96abab60-238a-44ff-8962-5806aa60cbce"`                      // UUID referring to identity
	Created                       time.Time                  `bson:"created" json:"created"  validate:"required" example:"2024-01-01T09:00:00.000Z"`                                                  // Timestamp of the creation of the playbook
	Modified                      time.Time                  `bson:"modified" json:"modified" validate:"required" example:"2024-01-01T09:00:00.000Z"`                                                 // Timestamp of the last modification of the playbook
	ValidFrom                     time.Time                  `bson:"valid_from,omitempty" json:"valid_from,omitempty" example:"2024-01-01T09:00:00.000Z"`                                             // Timestamp from when the playbook is valid
	ValidUntil                    time.Time                  `bson:"valid_until,omitempty" json:"valid_until,omitempty" validate:"omitempty,gtecsfield=ValidFrom" example:"2124-01-01T09:00:00.000Z"` // Timestamp until when the playbook is valid
	DerivedFrom                   []string                   `bson:"derived_from,omitempty" json:"derived_from,omitempty" example:"[\"playbook--77c4c428-6304-4950-93ff-83c5224cb67a\"]"`             // Playbook id that this playbook is derived from
	Priority                      int                        `bson:"priority,omitempty" json:"priority,omitempty" example:"100"`                                                                      // A priority number ranging 0 - 100
	Severity                      int                        `bson:"severity,omitempty" json:"severity,omitempty" example:"100"`                                                                      // A priority number ranging 0 - 100
	Impact                        int                        `bson:"impact,omitempty" json:"impact,omitempty" example:"100"`                                                                          // A priority number ranging 0 - 100
	Labels                        []string                   `bson:"labels,omitempty" json:"labels,omitempty"`                                                                                        // List of labels to label playbook
	ExternalReferences            []ExternalReferences       `bson:"external_references,omitempty" json:"external_references,omitempty"`                                                              // List of external reference objects
	Markings                      []string                   `bson:"markings,omitempty" json:"markings,omitempty" example:"[marking-statement--6424867b-0440-4885-bd0b-604d51786d06]"`                // List of datamarking identifiers
	WorkflowStart                 string                     `bson:"workflow_start" json:"workflow_start" validate:"required" example:"start--07bea005-4a36-4a77-bd1f-79a6e4682a13"`                  // Start step of the playbook MUST be of step type START
	WorkflowException             string                     `bson:"workflow_exception,omitempty" json:"workflow_exception,omitempty" example:"end--37bea005-4a36-4a77-bd1f-79a6e4682a13"`            // Step that marks the actions that need to be taken when an exception occurs
	Workflow                      Workflow                   `bson:"workflow" json:"workflow" validate:"required"`                                                                                    // Map of workflow steps keyed by the step id
	DataMarkingDefinitions        DataMarkings               `bson:"data_marking_definitions,omitempty" json:"data_marking_definitions,omitempty"`                                                    // Map of datamarking definitions
	AuthenticationInfoDefinitions AuthenticationInformations `bson:"authentication_info_definitions,omitempty" json:"authentication_info_definitions,omitempty"`                                      // Map of authentication information objects
	AgentDefinitions              AgentTargets               `bson:"agent_definitions,omitempty" json:"agent_definitions,omitempty"`                                                                  // Map of agent definitions used by the workflow steps
	TargetDefinitions             AgentTargets               `bson:"target_definitions,omitempty" json:"target_definitions,omitempty"`                                                                // Map of target definitions used by the workflow steps
	ExtensionDefinitions          ExtensionDefinitions       `bson:"extension_definitions,omitempty" json:"extension_definitions,omitempty"`                                                          // Map of extension definitions used by the workflow steps
	PlaybookVariables             Variables                  `bson:"playbook_variables,omitempty" json:"playbook_variables,omitempty"`                                                                // Map of variables that are global to the playbook
	PlaybookExtensions            Extensions                 `bson:"playbook_extensions,omitempty" json:"playbook_extensions,omitempty"`                                                              // Map of extensions used by the playbook
}

type CivicLocation struct {
	Name               string `bson:"name,omitempty" json:"name,omitempty"`
	Description        string `bson:"description,omitempty" json:"description,omitempty"`
	BuildingDetails    string `bson:"building_details,omitempty" json:"building_details,omitempty"`
	NetworkDetails     string `bson:"network_details,omitempty" json:"network_details,omitempty"`
	Region             string `bson:"region,omitempty" json:"region,omitempty"`
	Country            string `bson:"country,omitempty" json:"country,omitempty"`
	AdministrativeArea string `bson:"administrative_area,omitempty" json:"administrative_area,omitempty"`
	City               string `bson:"city,omitempty" json:"city,omitempty"`
	StreetAddress      string `bson:"street_address,omitempty" json:"street_address,omitempty"`
	PostalCode         string `bson:"postal_code,omitempty" json:"postal_code,omitempty"`
	Latitude           string `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Longitude          string `bson:"longitude,omitempty" json:"longitude,omitempty"`
	Precision          string `bson:"precision,omitempty" json:"precision,omitempty"`
}

type Contact struct {
	Email          map[string]string `bson:"email,omitempty" json:"email,omitempty"`
	Phone          map[string]string `bson:"phone,omitempty" json:"phone,omitempty"`
	ContactDetails string            `bson:"contact_details,omitempty" json:"contact_details,omitempty"`
}

type AgentTarget struct {
	ID                    string        `bson:"id,omitempty" json:"id,omitempty"`
	Type                  string        `bson:"type" json:"type" validate:"required"`
	Name                  string        `bson:"name" json:"name" validate:"required"`
	Description           string        `bson:"description,omitempty" json:"description,omitempty"`
	Location              CivicLocation `bson:"location,omitempty" json:"location,omitempty"`
	AgentTargetExtensions Extensions    `bson:"agent_target_extensions,omitempty" json:"agent_target_extensions,omitempty"`
	Contact               Contact       `bson:"contact,omitempty" json:"contact,omitempty"`
	Logical               []string      `bson:"logical,omitempty" json:"logical,omitempty"`
	Sector                string        `bson:"sector,omitempty" json:"sector,omitempty"`
	AuthInfoIdentifier    string        `bson:"authentication_info,omitempty" json:"authentication_info,omitempty"`
	Category              []string      `bson:"category,omitempty" json:"category,omitempty"`
	Address               Addresses     `bson:"address,omitempty" json:"address,omitempty"`
	Port                  string        `bson:"port,omitempty" json:"port,omitempty"`
}

type AuthenticationInformation struct {
	ID               string `bson:"id,omitempty" json:"id,omitempty"`
	Type             string `bson:"type" json:"type" validate:"required"`
	Name             string `bson:"name,omitempty" json:"name,omitempty"`
	Description      string `bson:"description,omitempty" json:"description,omitempty"`
	Username         string `bson:"username,omitempty" json:"username,omitempty"`
	UserId           string `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Password         string `bson:"password,omitempty" json:"password,omitempty"`
	PrivateKey       string `bson:"private_key,omitempty" json:"private_key,omitempty"`
	Kms              bool   `bson:"kms" json:"kms"`
	KmsKeyIdentifier string `bson:"kms_key_identifier,omitempty" json:"kms_key_identifier,omitempty"`
	Token            string `bson:"token,omitempty" json:"token,omitempty"`
	OauthHeader      string `bson:"oauth_header,omitempty" json:"oauth_header,omitempty"`
}

type ExternalReferences struct {
	Name        string `bson:"name" json:"name" validate:"required"`
	Description string `bson:"description" json:"description,omitempty"`
	Source      string `bson:"source" json:"source,omitempty"`
	URL         string `bson:"url" json:"url,omitempty"`
	ExternalId  string `bson:"external_id" json:"external_id,omitempty"`
	ReferenceId string `bson:"reference_id" json:"reference_id,omitempty"`
}

type ExtensionDefinition struct {
	ID                 string               `bson:"id,omitempty" json:"id,omitempty"`
	Type               string               `bson:"type" json:"type" validate:"required"`
	Name               string               `bson:"name" json:"name" validation:"required"`
	Description        string               `bson:"description,omitempty" json:"description,omitempty"`
	CreatedBy          string               `bson:"created_by" json:"created_by" validate:"required"`
	Schema             string               `bson:"schema" json:"schema" validate:"required"`
	Version            string               `bson:"version" json:"version" validate:"required"`
	ExternalReferences []ExternalReferences `bson:"external_references,omitempty" json:"external_references,omitempty"`
}

type Command struct {
	Type             string  `bson:"type"  json:"type" validate:"required"`
	Command          string  `bson:"command" json:"command" validate:"required"`
	Description      string  `bson:"description,omitempty" json:"description,omitempty"`
	CommandB64       string  `bson:"command_b64,omitempty" json:"command_b64,omitempty"`
	Version          string  `bson:"version,omitempty" json:"version,omitempty"`
	PlaybookActivity string  `bson:"playbook_activity,omitempty" json:"playbook_activity,omitempty"`
	Headers          Headers `bson:"headers,omitempty" json:"headers,omitempty"`
	Content          string  `bson:"content,omitempty" json:"content,omitempty"`
	ContentB64       string  `bson:"content_b64,omitempty" json:"content_b64,omitempty"`
}

type OutArgs []string

type Step struct {
	Type               string               `bson:"type" json:"type" validate:"required"` // The uuid of the step (not part of CACAO spec, but included in object for utility)
	ID                 string               `bson:"id,omitempty" json:"id,omitempty"`
	Name               string               `bson:"name,omitempty" json:"name,omitempty"`
	Description        string               `bson:"description,omitempty" json:"description,omitempty"`
	ExternalReferences []ExternalReferences `bson:"external_references,omitempty" json:"external_references,omitempty"`
	Delay              int                  `bson:"delay,omitempty" json:"delay,omitempty"`
	Timeout            int                  `bson:"timeout,omitempty" json:"timeout,omitempty"`
	StepVariables      Variables            `bson:"step_variables,omitempty" json:"step_variables,omitempty"`
	Owner              string               `bson:"owner,omitempty" json:"owner,omitempty"`
	OnCompletion       string               `bson:"on_completion,omitempty" json:"on_completion,omitempty"`
	OnSuccess          string               `bson:"on_success,omitempty" json:"on_success,omitempty"`
	OnFailure          string               `bson:"on_failure,omitempty" json:"on_failure,omitempty"`
	Commands           []Command            `bson:"commands,omitempty" json:"commands,omitempty"`
	Agent              string               `bson:"agent,omitempty" json:"agent,omitempty"`
	Targets            []string             `bson:"targets,omitempty" json:"targets,omitempty"`
	InArgs             []string             `bson:"in_args,omitempty" json:"in_args,omitempty"`
	OutArgs            OutArgs              `bson:"out_args,omitempty" json:"out_args,omitempty"`
	PlaybookID         string               `bson:"playbook_id,omitempty" json:"playbook_id,omitempty"`
	PlaybookVersion    string               `bson:"playbook_version,omitempty" json:"playbook_version,omitempty"`
	NextSteps          []string             `bson:"next_steps,omitempty" json:"next_steps,omitempty"`
	Condition          string               `bson:"condition,omitempty" json:"condition,omitempty"`
	OnTrue             string               `bson:"on_true,omitempty" json:"on_true,omitempty"`
	OnFalse            string               `bson:"on_false,omitempty" json:"on_false,omitempty"`
	Switch             string               `bson:"switch,omitempty" json:"switch,omitempty"`
	Cases              Cases                `bson:"cases,omitempty" json:"cases,omitempty"`
	AuthenticationInfo string               `bson:"authentication_info,omitempty" json:"authentication_info,omitempty"`
	StepExtensions     Extensions           `bson:"step_extensions,omitempty" json:"step_extensions,omitempty"`
}

type DataMarking struct {
	Type                       string               `bson:"type" json:"type" validate:"required"`
	ID                         string               `bson:"id" json:"id" validate:"required"`
	Name                       string               `bson:"name,omitempty" json:"name,omitempty"`
	Description                string               `bson:"description,omitempty" json:"description,omitempty"`
	CreatedBy                  string               `bson:"created_by" json:"created_by" validate:"required"`
	Created                    time.Time            `bson:"created" json:"created" validate:"required"`
	Revoked                    bool                 `bson:"revoked,omitempty" json:"revoked,omitempty"`
	ValidFrom                  time.Time            `bson:"valid_from,omitempty" json:"valid_from,omitempty"`
	ValidUntil                 time.Time            `bson:"valid_until,omitempty" json:"valid_until,omitempty" validate:"gtecsfield=ValidFrom"`
	Labels                     []string             `bson:"labels,omitempty" json:"labels,omitempty"`
	ExternalReferences         []ExternalReferences `bson:"external_references,omitempty" json:"external_references,omitempty"`
	TLPv2Level                 string               `bson:"tlpv2_level,omitempty" json:"tlpv2_level,omitempty"`
	Statement                  string               `bson:"statement,omitempty" json:"statement,omitempty"`
	TLP                        string               `bson:"tlp,omitempty" json:"tlp,omitempty"`
	IEPVersion                 string               `bson:"iep_version,omitempty" json:"iep_version,omitempty"`
	StartDate                  time.Time            `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate                    time.Time            `bson:"end_date,omitempty" json:"end_date,omitempty" validate:"gtecsfield=StartDate"`
	EncryptInTransit           string               `bson:"encrypt_in_transit,omitempty" json:"encrypt_in_transit,omitempty"`
	PermittedActions           string               `bson:"permitted_actions,omitempty" json:"permitted_actions,omitempty"`
	AffectedPartyNotifications string               `bson:"affected_party_notifications,omitempty" json:"affected_party_notifications,omitempty"`
	Attribution                string               `bson:"attribution,omitempty" json:"attribution,omitempty"`
	UnmodifiedResale           string               `bson:"unmodified_resale,omitempty" json:"unmodified_resale,omitempty"`
	MarkingExtensions          Extensions           `bson:"marking_extensions,omitempty" json:"marking_extensions,omitempty"`
}

// Deprecated
func Decode(data []byte) *Playbook {
	var playbook = NewPlaybook()

	if err := json.Unmarshal(data, &playbook); err != nil {
		return nil
	}

	for key, workflow := range playbook.Workflow {
		workflow.ID = key
		playbook.Workflow[key] = workflow
	}

	for key, target := range playbook.TargetDefinitions {
		target.ID = key
		playbook.TargetDefinitions[key] = target
	}

	for key, agent := range playbook.AgentDefinitions {
		agent.ID = key
		playbook.AgentDefinitions[key] = agent
	}

	for key, auth := range playbook.AuthenticationInfoDefinitions {
		auth.ID = key
		playbook.AuthenticationInfoDefinitions[key] = auth
	}

	return playbook
}

func NewPlaybook() *Playbook {
	playbook := Playbook{}
	playbook.AgentDefinitions = NewAgentTargets()
	playbook.TargetDefinitions = NewAgentTargets()
	playbook.PlaybookVariables = NewVariables()
	playbook.AuthenticationInfoDefinitions = NewAuthenticationInfoDefinitions()
	playbook.ExtensionDefinitions = NewExtensionDefinitions()
	playbook.DataMarkingDefinitions = NewDataMarkings()
	return &playbook
}
