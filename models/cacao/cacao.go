package cacao

import (
	//"strconv"
	"encoding/json"
	"time"
)

type Metadata struct{}

type (
	Target       struct{}
	DataMarkings struct{}
	Extensions   struct{}
)

type Playbook struct {
	ID                            string                               `bson:"_id" json:"id" validate:"required"`
	Type                          string                               `bson:"type" json:"type" validate:"required" `
	SpecVersion                   string                               `bson:"spec_version" json:"spec_version" validate:"required"`
	Name                          string                               `bson:"name" json:"name" validate:"required"`
	Description                   string                               `bson:"description" json:"description" validate:"required"`
	PlaybookTypes                 []string                             `bson:"playbook_types" json:"playbook_types" validate:"required"`
	CreatedBy                     string                               `bson:"created_by" json:"created_by"  validate:"required"`
	Created                       time.Time                            `bson:"created" json:"created"  validate:"required"`  // date time is already validate by the field type!
	Modified                      time.Time                            `bson:"modified" json:"modified" validate:"required"` //,datetime=2006-01-02T15:04:05Z07:00"`
	ValidFrom                     time.Time                            `bson:"valid_from" json:"valid_from" validate:"required,ltecsfield=ValidUntil"`
	ValidUntil                    time.Time                            `bson:"valid_until" json:"valid_until" validate:"required,gtcsfield=ValidFrom"`
	DerivedFrom                   []string                             `bson:"derived_form" json:"derived_from"`
	Priority                      int                                  `bson:"priority" json:"priority" validate:"required"`
	Severity                      int                                  `bson:"severity" json:"severity" validate:"required"`
	Impact                        int                                  `bson:"impact" json:"impact" validate:"required"`
	Labels                        []string                             `bson:"labels" json:"labels" validate:"required,dive"`
	ExternalReferences            []ExternalReferences                 `bson:"external_references" json:"external_references" validate:"required,dive"`
	Markings                      []string                             `bson:"markings" json:"markings"`
	WorkflowStart                 string                               `bson:"workflow_start" json:"workflow_start" validate:"required"`
	WorkflowException             string                               `bson:"workflow_exception" json:"workflow_exception" validate:"required"`
	Workflow                      map[string]Step                      `bson:"workflow"  json:"workflow" validate:"required"`
	DataMarkingDefs               DataMarking                          `bson:"data_markings" json:"data_marking_definitions" validate:"omitempty"`
	AuthenticationInfoDefinitions map[string]AuthenticationInformation `bson:"authentication_information" json:"authentication_info_definitions" validate:"omitempty"`
}

type AuthenticationInformation struct {
	Type             string `bson:"type"  json:"type" validate:"required"`
	Name             string `bson:"name" json:"name" validate:"omitempty"`
	Description      string `bson:"description" json:"description" validate:"omitempty"`
	Username         string `bson:"username" json:"username" validate:"omitempty"`
	UserId           string `bson:"user_id" json:"user_id" validate:"omitempty"`
	Password         string `bson:"password" json:"password" validate:"omitempty"`
	PrivateKey       string `bson:"private_key" json:"private_key" validate:"omitempty"`
	Kms              bool   `bson:"kms" json:"kms" validate:"omitempty"`
	KmsKeyIdentifier string `bson:"kms_key_identifier" json:"kms_key_identifier" validate:"omitempty"`
	Token            string `bson:"token" json:"token" validate:"omitempty"`
	OauthHeader      string `bson:"oauth_header" json:"oauth_header" validate:"omitempty"`
}

type ExternalReferences struct {
	Name        string `bson:"name" json:"name" validate:"required"`
	Description string `bson:"description" json:"description" validate:"required"`
	Source      string `bson:"source" json:"source" validate:"required"`
	URL         string `bson:"url"  json:"url" validate:"required,url"`
}
type Command struct {
	Type             string `bson:"type"  json:"type" validate:"required"`
	Command          string `bson:"command" json:"command" validate:"required"`
	Description      string `bson:"description" json:"description" validate:"omitempty"`
	CommandB64       string `bson:"base-64-command-string" json:"command_b64" validate:"omitempty"`
	Version          string `bson:"version" json:"version" validate:"omitempty"`
	PlaybookActivity string `bson:"playbook-activity" json:"playbook_activity" validate:"omitempty"`
}

type Variables struct {
	ObjectType  string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value,omitempty"`
	Constant    bool   `json:"constant,omitempty"`
	External    bool   `json:"external,omitempty"`
}

type Step struct {
	ObjectType         string               `json:"type,omitempty"`
	ID                 string               `json:"id,omitempty"`
	Name               string               `json:"name,omitempty"`
	Description        string               `json:"description,omitempty"`
	ExternalReferences []ExternalReferences `json:"external_references,omitempty"`
	Delay              int                  `json:"delay,omitempty"`
	Timeout            int                  `json:"timeout,omitempty"`
	StepVariables      map[string]Variables `json:"playbook_variables,omitempty"`
	Owner              string               `json:"owner,omitempty"`
	OnCompletion       string               `json:"on_completion,omitempty"`
	OnSuccess          string               `json:"on_success,omitempty"`
	OnFailure          string               `json:"on_failure,omitempty"`
	Commands           []Command            `json:"commands,omitempty"`
	Agent              string               `json:"agent,omitempty"`
	Targets            []string             `json:"targets,omitempty"`
	InArgs             []string             `json:"in_args,omitempty"`
	OutArgs            []string             `json:"out_args,omitempty"`
	PlaybookID         string               `json:"playbook_id,omitempty"`
	PlaybookVersion    string               `json:"playbook_version,omitempty"`
	NextSteps          []string             `json:"next_steps,omitempty"`
	Condition          string               `json:"condition,omitempty"`
	OnTrue             []string             `json:"on_true,omitempty"`
	OnFalse            []string             `json:"on_false,omitempty"`
	Switch             string               `json:"switch,omitempty"`
	Cases              map[string][]string  `json:"cases,omitempty"`
}

type DataMarking struct {
	ObjectType                 string               `json:"type,omitempty"`
	ID                         string               `json:"id,omitempty"`
	Name                       string               `json:"name,omitempty"`
	Description                string               `json:"description,omitempty"`
	CreatedBy                  string               `json:"created_by,omitempty"`
	Created                    string               `json:"created,omitempty"`
	Revoked                    bool                 `json:"revoked,omitempty"`
	ValidFrom                  string               `json:"valid_from,omitempty"`
	ValidUntil                 string               `json:"valid_until,omitempty"`
	Labels                     []string             `json:"labels,omitempty"`
	ExternalReferences         []ExternalReferences `json:"external_references,omitempty"`
	TLPv2Level                 string               `json:"tlpv2_level,omitempty"`
	Statement                  string               `json:"statement,omitempty"`
	TLP                        string               `json:"tlp,omitempty"`
	IEPVersion                 string               `json:"iep_version,omitempty"`
	StartDate                  string               `json:"start_date,omitempty"`
	EndDate                    string               `json:"end_date,omitempty"`
	EncryptInTransit           string               `json:"encrypt_in_transit,omitempty"`
	PermittedActions           string               `json:"permitted_actions,omitempty"`
	AffectedPartyNotifications string               `json:"affected_party_notifications,omitempty"`
	Attribution                string               `json:"attribution,omitempty"`
	UnmodifiedResale           string               `json:"unmodified_resale,omitempty"`
	//marking_extensions
}

func Decode(data []byte) *Playbook {
	var playbook Playbook

	if err := json.Unmarshal(data, &playbook); err != nil {
		return nil
	}

	for key, workflow := range playbook.Workflow {
		workflow.ID = key
		playbook.Workflow[key] = workflow
	}

	return &playbook
}
