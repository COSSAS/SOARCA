package playbookrepository

import (
	"errors"

	database "soarca/database"
	"soarca/database/projections"
	validator "soarca/internal/validators"
	"soarca/models/api"
	"soarca/models/cacao"
	"soarca/models/decoder"
)

type IPlaybookRepository interface {
	GetPlaybooks() ([]cacao.Playbook, error)
	GetPlaybookMetas() ([]api.PlaybookMeta, error)
	Create(jsonData *[]byte) (cacao.Playbook, error)
	Read(id string) (cacao.Playbook, error)
	Update(id string, jsonData *[]byte) (cacao.Playbook, error)
	Delete(id string) error
}

type PlaybookRepository struct {
	db      database.Database
	options database.FindOptions
}

func SetupPlaybookRepository(db database.Database, options database.FindOptions) *PlaybookRepository {
	return &PlaybookRepository{db: db, options: options}
}

func (playbookRepo *PlaybookRepository) GetPlaybookMetas() ([]api.PlaybookMeta, error) {
	playbookMetas, err := playbookRepo.db.Find(nil, playbookRepo.options.GetProjectionByType(projections.Meta))
	if err != nil {
		return nil, err
	}

	var returnPlaybookMetas []api.PlaybookMeta

	for _, playbookMeta := range playbookMetas {
		playbookMeta, ok := playbookMeta.(cacao.Playbook)
		if !ok {
			return nil, errors.New("type assertion failed for cacao.Playbook to cacao.PlaybookMeta type")
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

func (playbookRepo *PlaybookRepository) GetPlaybooks() ([]cacao.Playbook, error) {
	playbooks, err := playbookRepo.db.Find(nil)
	if err != nil {
		return nil, err
	}

	var returnListPlaybooks []cacao.Playbook
	for _, playbook := range playbooks {
		// get the cacao playbook id and add to the return list
		playbook, ok := playbook.(cacao.Playbook)
		if !ok {
			return nil, errors.New("type assertion failed for cacao.playbook type")
		}
		decode(&playbook)
		returnListPlaybooks = append(returnListPlaybooks, playbook)
	}
	return returnListPlaybooks, nil
}

func (playbookRepo *PlaybookRepository) Create(jsonData *[]byte) (cacao.Playbook, error) {
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
	decode(&playbook)

	return playbook, playbookRepo.db.Create(client_data)
}

func (playbookRepo *PlaybookRepository) Read(id string) (cacao.Playbook, error) {
	returnedObject, err := playbookRepo.db.Read(id)
	if err != nil {
		return cacao.Playbook{}, err
	}

	cacaoPlaybook, ok := returnedObject.(cacao.Playbook)

	if !ok {
		err = errors.New("could not cast lookup object to cacao.Playbook type")
		return cacao.Playbook{}, err
	}
	decode(&cacaoPlaybook)

	return cacaoPlaybook, nil
}

func (playbookRepo *PlaybookRepository) Update(id string, jsonData *[]byte) (cacao.Playbook, error) {
	// validate the input object to required type and unmarshal
	client_data, err := validator.UnmarshalJson[cacao.Playbook](jsonData)
	if err != nil {
		return cacao.Playbook{}, err
	}
	cacaoPlaybook, ok := client_data.(cacao.Playbook)
	if !ok {
		err = errors.New("could not cast lookup object to cacao.Playbook type")
		return cacao.Playbook{}, err
	}
	decode(&cacaoPlaybook)
	return cacaoPlaybook, playbookRepo.db.Update(id, client_data)
}

func (playbookRepo *PlaybookRepository) Delete(id string) error {
	// validate the input object to required type and unmarshal
	return playbookRepo.db.Delete(id)
}

func decode(playbook *cacao.Playbook) {
	if playbook.PlaybookVariables == nil {
		playbook.PlaybookVariables = cacao.NewVariables()
	}
	decoder.SetPlaybookKeysAsId(playbook)
}
