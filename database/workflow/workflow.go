package workflowrepository

import (
	"errors"

	database "soarca/database"
	"soarca/database/projections"
	validator "soarca/internal/validators"
	"soarca/models/api"
	"soarca/models/cacao"
)

type IWorkflowRepository interface {
	GetWorkflows() ([]cacao.Playbook, error)
	GetWorkflowMetas() ([]api.PlaybookMeta, error)
	Create(jsonData *[]byte) (cacao.Playbook, error)
	Read(id string) (cacao.Playbook, error)
	Update(id string, jsonData *[]byte) (cacao.Playbook, error)
	Delete(id string) error
}

type WorkflowRepository struct {
	db      database.Database
	options database.FindOptions
}

func SetupWorkflowRepository(db database.Database, options database.FindOptions) *WorkflowRepository {
	return &WorkflowRepository{db: db, options: options}
}

func (workflowrepo *WorkflowRepository) GetWorkflowMetas() ([]api.PlaybookMeta, error) {
	playbookMetas, err := workflowrepo.db.Find(nil, workflowrepo.options.GetProjectionByType(projections.Meta))
	if err != nil {
		return nil, err
	}

	var returnPlaybookMetas []api.PlaybookMeta

	for _, playbookMeta := range playbookMetas {
		playbookMeta, ok := playbookMeta.(cacao.Playbook)
		if !ok {
			return nil, errors.New("type assertion failed for cacao.Playbook to cacao.WorkflowMeta type")
		}
		returnPlaybookMetas = append(returnPlaybookMetas, api.PlaybookMeta{
			ID:          playbookMeta.ID,
			Name:        playbookMeta.Name,
			Description: playbookMeta.Description,
			ValidFrom:   playbookMeta.ValidFrom,
			ValidUntil:  playbookMeta.ValidUntil,
			Labels:      playbookMeta.Labels,
		})
	}
	return returnPlaybookMetas, nil
}

func (workflowrepo *WorkflowRepository) GetWorkflows() ([]cacao.Playbook, error) {
	return_workflows, err := workflowrepo.db.Find(nil)
	if err != nil {
		return nil, err
	}

	var returnListWorkflows []cacao.Playbook
	for _, workflow := range return_workflows {
		// get the cacao playbook id and add to the return list
		workflow, ok := workflow.(cacao.Playbook)
		if !ok {
			return nil, errors.New("type assertion failed for workflow to cacao.playbook type")
		}
		returnListWorkflows = append(returnListWorkflows, workflow)
	}
	return returnListWorkflows, nil
}

func (workflowrepo *WorkflowRepository) Create(jsonData *[]byte) (cacao.Playbook, error) {
	// validate the input object to required type and unmarshal
	client_data, err := validator.UnmarshalJson[cacao.Playbook](jsonData)
	if err != nil {
		return cacao.Playbook{}, err
	}
	playbook, ok := client_data.(cacao.Playbook)
	if !ok {
		// handle incorrect casting
		return cacao.Playbook{}, errors.New("failed to cast playbook object")
	}

	return playbook, workflowrepo.db.Create(client_data)
}

func (workflowrepo *WorkflowRepository) Read(id string) (cacao.Playbook, error) {
	returnedObject, err := workflowrepo.db.Read(id)
	if err != nil {
		return cacao.Playbook{}, err
	}

	cacaoPlaybook, ok := returnedObject.(cacao.Playbook)
	if !ok {
		err = errors.New("Could not cast lookup object to cacao.Playbook type")
		return cacao.Playbook{}, err
	}

	return cacaoPlaybook, nil
}

func (workflowrepo *WorkflowRepository) Update(id string, jsonData *[]byte) (cacao.Playbook, error) {
	// validate the input object to required type and unmarshal
	client_data, err := validator.UnmarshalJson[cacao.Playbook](jsonData)
	if err != nil {
		return cacao.Playbook{}, err
	}
	cacaoPlaybook, ok := client_data.(cacao.Playbook)
	if !ok {
		err = errors.New("Could not cast lookup object to cacao.Playbook type")
		return cacao.Playbook{}, err
	}
	return cacaoPlaybook, workflowrepo.db.Update(id, client_data)
}

func (workflowrepo *WorkflowRepository) Delete(id string) error {
	// validate the input object to required type and unmarshal
	return workflowrepo.db.Delete(id)
}
