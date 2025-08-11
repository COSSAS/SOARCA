package conversion

import (
	"errors"
	"soarca/pkg/models/cacao"
)

type MispConverter struct {
}

func (MispConverter) Convert(input []byte, filename string) (*cacao.Playbook, error) {
	return nil, errors.New("Unimplemented")

}
func NewMispConverter() MispConverter {
	return MispConverter{}
}
