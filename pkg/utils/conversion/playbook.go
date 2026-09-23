package conversion

import (
	"fmt"
	"soarca/pkg/models/cacao"
	"time"

	"github.com/google/uuid"
)

func NewSoarcaPlaybook(name string, types []string) (*cacao.Playbook, string, string) {
	soarca_name := fmt.Sprintf("soarca--%s", uuid.New())
	playbook := cacao.NewPlaybook()
	playbook.SpecVersion = cacao.CACAO_VERSION_2
	playbook.Type = "playbook"
	playbook.ID = fmt.Sprintf("playbook--%s", uuid.New())
	playbook.CreatedBy = soarca_name
	playbook.Name = name
	playbook.Created = time.Now().UTC()
	playbook.Modified = time.Now().UTC()
	playbook.ValidFrom = time.Now().UTC()
	playbook.ValidUntil = time.Now().UTC()
	playbook.PlaybookTypes = types
	soarca_manual_name := fmt.Sprintf("soarca-manual--%s", uuid.New())
	playbook.AgentDefinitions = cacao.NewAgentTargets(
		cacao.AgentTarget{
			ID:   soarca_name,
			Type: "soarca",
			Name: "soarca-playbook"},
		cacao.AgentTarget{
			ID:          soarca_manual_name,
			Type:        "soarca-manual",
			Name:        "soarca-manual",
			Description: "SOARCAs manual command handler"})
	return playbook, soarca_name, soarca_manual_name
}
