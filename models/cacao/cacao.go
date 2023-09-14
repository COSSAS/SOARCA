package cacao

import "time"

type Metadata struct{}

type (
	Target       struct{}
	DataMarkings struct{}
	Extensions   struct{}
)

type Playbook struct {
	ID                 string               `bson:"_id" json:"id" validate:"required"`
	Type               string               `bson:"type" json:"type" validate:"required" `
	SpecVersion        string               `bson:"spec_version" json:"spec_version" validate:"required"`
	Name               string               `bson:"name" json:"name" validate:"required"`
	Description        string               `bson:"description" json:"description" validate:"required"`
	PlaybookTypes      []string             `bson:"playbook_types" json:"playbook_types" validate:"required"`
	CreatedBy          string               `bson:"created_by" json:"created_by"  validate:"required"`
	Created            time.Time            `bson:"created" json:"created"  validate:"required"`  // date time is already validate by the field type!
	Modified           time.Time            `bson:"modified" json:"modified" validate:"required"` //,datetime=2006-01-02T15:04:05Z07:00"`
	ValidFrom          time.Time            `bson:"valid_from" json:"valid_from" validate:"required,ltecsfield=ValidUntil"`
	ValidUntil         time.Time            `bson:"valid_until" json:"valid_until" validate:"required,gtcsfield=ValidFrom"`
	Priority           int                  `bson:"priority" json:"priority" validate:"required"`
	Severity           int                  `bson:"severity" json:"severity" validate:"required"`
	Impact             int                  `bson:"impact" json:"impact" validate:"required"`
	Labels             []string             `bson:"labels" json:"labels" validate:"required,dive"`
	ExternalReferences []ExternalReferences `bson:"external_references" json:"external_references" validate:"required,dive"`
	WorkflowStart      string               `bson:"workflow_start" json:"workflow_start" validate:"required"`
	WorkflowException  string               `bson:"workflow_exception" json:"workflow_exception" validate:"required"`
	Workflow           []Step               `bson:"workflow"  json:"workflow" validate:"required"`
}
type ExternalReferences struct {
	Name        string `bson:"name" json:"name" validate:"required"`
	Description string `bson:"description" json:"description" validate:"required"`
	Source      string `bson:"source" json:"source" validate:"required"`
	URL         string `bson:"url"  json:"url" validate:"required,url"`
}
type Commands struct {
	Type    string `bson:"type"  json:"type" validate:"required"`
	Command string `bson:"command" json:"command" validate:"required"`
}

type Step struct {
	UUID         string     `bson:"step_uuid" json:"step_uuid" validate:"required"`
	Type         string     `bson:"type" json:"type" validate:"required"`
	Name         string     `bson:"name" json:"name" validate:"required"`
	Description  string     `bson:"description" json:"description" validate:"required"`
	OnCompletion string     `bson:"on_completion" json:"on_completion" validate:"required"`
	Commands     []Commands `bson:"commands" json:"commands" validate:"required"`
}
