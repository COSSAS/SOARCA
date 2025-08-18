package memory

import (
	"errors"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/decoder"
)

type InMemoryPlaybookDatabase struct {
	playbooks map[string]cacao.Playbook
}

func NewPlaybookDatabase() *InMemoryPlaybookDatabase {
	return &InMemoryPlaybookDatabase{playbooks: make(map[string]cacao.Playbook)}
}

func (memory *InMemoryPlaybookDatabase) GetPlaybooks() ([]cacao.Playbook, error) {
	size := len(memory.playbooks)
	playbookList := make([]cacao.Playbook, 0, size)
	for _, playbook := range memory.playbooks {
		playbookList = append(playbookList, playbook)
	}

	return playbookList, nil
}

func (memory *InMemoryPlaybookDatabase) GetPlaybookMetas() ([]api.PlaybookMeta, error) {
	size := len(memory.playbooks)
	playbookList := make([]api.PlaybookMeta, 0, size)
	for _, playbook := range memory.playbooks {
		meta := api.PlaybookMeta{ID: playbook.ID, Name: playbook.Name,
			Description: playbook.Description,
			ValidFrom:   playbook.ValidFrom,
			ValidUntil:  playbook.ValidUntil,
			Labels:      playbook.Labels}
		playbookList = append(playbookList, meta)
	}

	return playbookList, nil
}

func (memory *InMemoryPlaybookDatabase) Create(json *[]byte) (cacao.Playbook, error) {

	if json == nil {
		return cacao.Playbook{}, errors.New("empty input")
	}
	result := decoder.DecodeValidate(*json)
	if result == nil {
		return cacao.Playbook{}, errors.New("failed to decode")
	}
	_, ok := memory.playbooks[result.ID]
	if ok {
		return cacao.Playbook{}, errors.New("playbook already exists")
	}
	memory.playbooks[result.ID] = *result
	return memory.playbooks[result.ID], nil
}

func (memory *InMemoryPlaybookDatabase) Read(id string) (cacao.Playbook, error) {
	playbook, ok := memory.playbooks[id]
	if !ok {
		return cacao.Playbook{}, errors.New("playbook is not in repository")
	}
	return playbook, nil
}

func (memory *InMemoryPlaybookDatabase) Update(id string, json *[]byte) (cacao.Playbook, error) {
	playbook, err := memory.Read(id)
	if err != nil {
		return playbook, err
	}
	updatePlaybook := decoder.DecodeValidate(*json)
	if updatePlaybook == nil {
		return cacao.Playbook{}, errors.New("failed to decode")
	}
	memory.playbooks[id] = *updatePlaybook
	return *updatePlaybook, nil
}

func (memory *InMemoryPlaybookDatabase) Delete(id string) error {
	delete(memory.playbooks, id)
	return nil
}
