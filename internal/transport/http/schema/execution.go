package api

import "github.com/google/uuid"

// RunStarted is returned when a playbook run has been accepted and started.
type RunStarted struct {
	RunId      uuid.UUID `json:"run_id" validate:"required" example:"2c855cd6-bbce-402f-a143-3d6eec346c08"`
	PlaybookId string    `json:"playbook_id" validate:"required" example:"playbook--0cec398c-db69-4f17-bde4-8ecbcc4a8879"`
}
