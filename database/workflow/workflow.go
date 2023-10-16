package workflowrepository

import (
	"errors"

	database "soarca/database"
	validator "soarca/internal/validators"
	"soarca/models/cacao"
)

type IWorkflowRepository interface {
	GetWorkflowIds() ([]string, error)
	Create(jsonData *[]byte) (string, error)
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

func (workflowrepo *WorkflowRepository) GetWorkflowIds() ([]string, error) {
	return_workflows, err := workflowrepo.db.Find(nil, workflowrepo.options.GetIds())
	if err != nil {
		return nil, err
	}

	var returnListIDs []string
	for _, workflow := range return_workflows {
		// get the cacao playbook id and add to the return list
		returnListIDs = append(returnListIDs, workflow.(cacao.Playbook).ID)
	}
	return returnListIDs, nil
}

func (workflowrepo *WorkflowRepository) Create(jsonData *[]byte) (string, error) {
	// validate the input object to required type and unmarshal
	client_data, err := validator.UnmarshalJson[cacao.Playbook](jsonData)
	if err != nil {
		return "", err
	}
	playbook, ok := client_data.(cacao.Playbook)
	if !ok {
		// handle incorrect casting
		return "", errors.New("Failed to cast playbook object")
	}

	return playbook.ID, workflowrepo.db.Create(client_data)
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
