package cacao

import "time"

type Metadata struct{}

type Target struct{}
type DataMarkings struct{}
type Extensions struct{}

type Playbook struct {
	Type               string               `json:"type" validate:"required"`
	SpecVersion        string               `json:"spec_version"  validate:"required"`
	ID                 string               `json:"id"  validate:"required"`
	Name               string               `json:"name"  validate:"required"`
	Description        string               `json:"description"  validate:"required"`
	PlaybookTypes      []string             `json:"playbook_types"  validate:"required"`
	CreatedBy          string               `json:"created_by"  validate:"required"`
	Created            time.Time            `json:"created"  validate:"required"`  //date time is already validate by the field type!
	Modified           time.Time            `json:"modified"  validate:"required"` //,datetime=2006-01-02T15:04:05Z07:00"`
	ValidFrom          time.Time            `json:"valid_from"  validate:"required,ltecsfield=ValidUntil"`
	ValidUntil         time.Time            `json:"valid_until"  validate:"required,gtcsfield=ValidFrom"`
	Priority           int                  `json:"priority"  validate:"required"`
	Severity           int                  `json:"severity"  validate:"required"`
	Impact             int                  `json:"impact"  validate:"required"`
	Labels             []string             `json:"labels"  validate:"required,dive"`
	ExternalReferences []ExternalReferences `json:"external_references"  validate:"required,dive"`
	WorkflowStart      string               `json:"workflow_start"  validate:"required"`
	WorkflowException  string               `json:"workflow_exception"  validate:"required"`
	Workflow           []Step               `json:"workflow"  validate:"required"`
}
type ExternalReferences struct {
	Name        string 	`json:"name"  validate:"required"`
	Description string 	`json:"description"  validate:"required"`
	Source      string 	`json:"source"  validate:"required"`
	URL         string 	`json:"url"  validate:"required,url"`
}
type Commands struct {
	Type    string `json:"type"  validate:"required"`
	Command string `json:"command"  validate:"required"`
}

type Step struct {
	UUID         string     `json:"step_uuid"  validate:"required"`
	Type         string     `json:"type"  validate:"required"`
	Name         string     `json:"name"  validate:"required"`
	Description  string     `json:"description"  validate:"required"`
	OnCompletion string     `json:"on_completion"  validate:"required"`
	Commands     []Commands `json:"commands"  validate:"required"`
}
