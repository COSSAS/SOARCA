package conversion

import (
	"soarca/pkg/models/cacao"
)

type IConverter interface {
	Convert(input []byte, filename string) (*cacao.Playbook, error)
}
