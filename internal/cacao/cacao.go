package cacao

import "time"

type Metadata struct{}

type Target struct{}
type DataMarkings struct{}
type Extensions struct{}

type Playbook struct {
	Type               string               `json:"type"`
	SpecVersion        string               `json:"spec_version"`
	ID                 string               `json:"id"`
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	PlaybookTypes      []string             `json:"playbook_types"`
	CreatedBy          string               `json:"created_by"`
	Created            time.Time            `json:"created"`
	Modified           time.Time            `json:"modified"`
	ValidFrom          time.Time            `json:"valid_from"`
	ValidUntil         time.Time            `json:"valid_until"`
	Priority           int                  `json:"priority"`
	Severity           int                  `json:"severity"`
	Impact             int                  `json:"impact"`
	Labels             []string             `json:"labels"`
	ExternalReferences []ExternalReferences `json:"external_references"`
	WorkflowStart      string               `json:"workflow_start"`
	WorkflowException  string               `json:"workflow_exception"`
	Workflow           []Step               `json:"workflow"`
}
type ExternalReferences struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	URL         string `json:"url"`
}
type Commands struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type Step struct {
	UUID         string     `json:"step_uuid"`
	Type         string     `json:"type"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	OnCompletion string     `json:"on_completion"`
	Commands     []Commands `json:"commands"`
}
