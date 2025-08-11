package conversion

import (
	"errors"
	"soarca/pkg/models/cacao"
)

type SplunkConverter struct {
}

func (SplunkConverter) Convert(input []byte, filename string) (*cacao.Playbook, error) {
	return nil, errors.New("Unimplemented")

}
func NewSplunkConverter() SplunkConverter {
	return SplunkConverter{}
}
