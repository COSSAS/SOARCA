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
	StepTypePlaybook        = "playbook-action"
	StepTypeParallel        = "parallel"
	StepTypeIfCondition     = "if-condition"
	StepTypeWhileCondition  = "while-condition"
	StepTypeSwitchCondition = "switch-condition"

	AuthInfoOAuth2Type    = "oauth2"
	AuthInfoHTTPBasicType = "http-basic"
	AuthInfoNotSet        = ""
	CACAO_VERSION_1       = "cacao-1.0"
	CACAO_VERSION_2       = "cacao-2.0"
)

type (
	Extensions map[string]interface{}
	Workflow   map[string]Step
)

type Playbook struct {
	ID                            string                               `bson:"_id" json:"id" validate:"required"`
	Type                          string                               `bson:"type" json:"type" validate:"required"`
	SpecVersion                   string                               `bson:"spec_version" json:"spec_version" validate:"required"`
	Name                          string                               `bson:"name" json:"name" validate:"required"`
	Description                   string                               `bson:"description,omitempty" json:"description,omitempty"`
	PlaybookTypes                 []string                             `bson:"playbook_types,omitempty" json:"playbook_types,omitempty"`
	CreatedBy                     string                               `bson:"created_by" json:"created_by"  validate:"required"`
	Created                       time.Time                            `bson:"created" json:"created"  validate:"required"`
	Modified                      time.Time                            `bson:"modified" json:"modified" validate:"required"`
	ValidFrom                     time.Time                            `bson:"valid_from,omitempty" json:"valid_from,omitempty"`
	ValidUntil                    time.Time                            `bson:"valid_until,omitempty" json:"valid_until,omitempty" validate:"omitempty,gtecsfield=ValidFrom"`
	DerivedFrom                   []string                             `bson:"derived_from,omitempty" json:"derived_from,omitempty"`
	Priority                      int                                  `bson:"priority,omitempty" json:"priority,omitempty"`
	Severity                      int                                  `bson:"severity,omitempty" json:"severity,omitempty"`
	Impact                        int                                  `bson:"impact,omitempty" json:"impact,omitempty"`
	Labels                        []string                             `bson:"labels,omitempty" json:"labels,omitempty"`
	ExternalReferences            []ExternalReferences                 `bson:"external_references,omitempty" json:"external_references,omitempty"`
	Markings                      []string                             `bson:"markings,omitempty" json:"markings,omitempty"`
	WorkflowStart                 string                               `bson:"workflow_start" json:"workflow_start" validate:"required"`
	WorkflowException             string                               `bson:"workflow_exception,omitempty" json:"workflow_exception,omitempty"`
	Workflow                      Workflow                             `bson:"workflow" json:"workflow" validate:"required"`
	DataMarkingDefinitions        map[string]DataMarking               `bson:"data_marking_definitions,omitempty" json:"data_marking_definitions,omitempty"`
	AuthenticationInfoDefinitions map[string]AuthenticationInformation `bson:"authentication_info_definitions,omitempty" json:"authentication_info_definitions,omitempty"`
	AgentDefinitions              map[string]AgentTarget               `bson:"agent_definitions,omitempty" json:"agent_definitions,omitempty"`
	TargetDefinitions             map[string]AgentTarget               `bson:"target_definitions,omitempty" json:"target_definitions,omitempty"`
	ExtensionDefinitions          map[string]ExtensionDefinition       `bson:"extension_definitions,omitempty" json:"extension_definitions,omitempty"`
	PlaybookVariables             VariableMap                          `bson:"playbook_variables,omitempty" json:"playbook_variables,omitempty"`
	PlaybookExtensions            Extensions                           `bson:"playbook_extensions,omitempty" json:"playbook_extensions,omitempty"`
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
	ID                    string              `bson:"id,omitempty" json:"id,omitempty"`
	Type                  string              `bson:"type" json:"type" validate:"required"`
	Name                  string              `bson:"name" json:"name" validate:"required"`
	Description           string              `bson:"description,omitempty" json:"description,omitempty"`
	Location              CivicLocation       `bson:"location,omitempty" json:"location,omitempty"`
	AgentTargetExtensions Extensions          `bson:"agent_target_extensions,omitempty" json:"agent_target_extensions,omitempty"`
	Contact               Contact             `bson:"contact,omitempty" json:"contact,omitempty"`
	Logical               []string            `bson:"logical,omitempty" json:"logical,omitempty"`
	Sector                string              `bson:"sector,omitempty" json:"sector,omitempty"`
	HttpUrl               string              `bson:"http_url,omitempty" json:"http_url,omitempty"`
	AuthInfoIdentifier    string              `bson:"authentication_info,omitempty" json:"authentication_info,omitempty"`
	Category              []string            `bson:"category,omitempty" json:"category,omitempty"`
	Address               map[string][]string `bson:"address,omitempty" json:"address,omitempty"`
	Port                  string              `bson:"port,omitempty" json:"port,omitempty"`
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
	Description string `bson:"description" json:"description" validate:"required"`
	Source      string `bson:"source" json:"source" validate:"required"`
	URL         string `bson:"url" json:"url" validate:"required,url"`
}

type ExtensionDefinition struct {
	Type               string               `bson:"type" json:"type" validate:"required"`
	Name               string               `bson:"name" json:"name" validation:"required"`
	Description        string               `bson:"description,omitempty" json:"description,omitempty"`
	CreatedBy          string               `bson:"created_by" json:"created_by" validate:"required"`
	Schema             string               `bson:"schema" json:"schema" validate:"required"`
	Version            string               `bson:"version" json:"version" validate:"required"`
	ExternalReferences []ExternalReferences `bson:"external_references,omitempty" json:"external_references,omitempty"`
}

type Command struct {
	Type             string            `bson:"type"  json:"type" validate:"required"`
	Command          string            `bson:"command" json:"command" validate:"required"`
	Description      string            `bson:"description,omitempty" json:"description,omitempty"`
	CommandB64       string            `bson:"command_b64,omitempty" json:"command_b64,omitempty"`
	Version          string            `bson:"version,omitempty" json:"version,omitempty"`
	PlaybookActivity string            `bson:"playbook_activity,omitempty" json:"playbook_activity,omitempty"`
	Headers          map[string]string `bson:"headers,omitempty" json:"headers,omitempty"`
	Content          string            `bson:"content,omitempty" json:"content,omitempty"`
	ContentB64       string            `bson:"content_b64,omitempty" json:"content_b64,omitempty"`
}

type Step struct {
	Type               string               `bson:"type" json:"type" validate:"required"`
	ID                 string               `bson:"id,omitempty" json:"id,omitempty"`
	Name               string               `bson:"name,omitempty" json:"name,omitempty"`
	Description        string               `bson:"description,omitempty" json:"description,omitempty"`
	ExternalReferences []ExternalReferences `bson:"external_references,omitempty" json:"external_references,omitempty"`
	Delay              int                  `bson:"delay,omitempty" json:"delay,omitempty"`
	Timeout            int                  `bson:"timeout,omitempty" json:"timeout,omitempty"`
	StepVariables      VariableMap          `bson:"step_variables,omitempty" json:"step_variables,omitempty"`
	Owner              string               `bson:"owner,omitempty" json:"owner,omitempty"`
	OnCompletion       string               `bson:"on_completion,omitempty" json:"on_completion,omitempty"`
	OnSuccess          string               `bson:"on_success,omitempty" json:"on_success,omitempty"`
	OnFailure          string               `bson:"on_failure,omitempty" json:"on_failure,omitempty"`
	Commands           []Command            `bson:"commands,omitempty" json:"commands,omitempty"`
	Agent              string               `bson:"agent,omitempty" json:"agent,omitempty"`
	Targets            []string             `bson:"targets,omitempty" json:"targets,omitempty"`
	InArgs             []string             `bson:"in_args,omitempty" json:"in_args,omitempty"`
	OutArgs            []string             `bson:"out_args,omitempty" json:"out_args,omitempty"`
	PlaybookID         string               `bson:"playbook_id,omitempty" json:"playbook_id,omitempty"`
	PlaybookVersion    string               `bson:"playbook_version,omitempty" json:"playbook_version,omitempty"`
	NextSteps          []string             `bson:"next_steps,omitempty" json:"next_steps,omitempty"`
	Condition          string               `bson:"condition,omitempty" json:"condition,omitempty"`
	OnTrue             string               `bson:"on_true,omitempty" json:"on_true,omitempty"`
	OnFalse            string               `bson:"on_false,omitempty" json:"on_false,omitempty"`
	Switch             string               `bson:"switch,omitempty" json:"switch,omitempty"`
	Cases              map[string]string    `bson:"cases,omitempty" json:"cases,omitempty"`
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
	var playbook Playbook

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

	return &playbook
}
