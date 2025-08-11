package conversion

import (
	"errors"
	"soarca/pkg/models/cacao"
)

func PerformConversion(input_filename string, input []byte, format_string string) (*cacao.Playbook, error) {
	format := FormatUnknown
	if format_string == "" {
		format = guess_format(input_filename)
	} else {
		format = read_format(format_string)
	}
	if format == FormatUnknown {
		return nil, errors.New("Could not deduce input file type")
	}
	var converter IConverter
	switch format {
	case FormatBpmn:
		converter = NewBpmnConverter()
	case FormatMisp:
		converter = NewMispConverter()
	case FormatSplunk:
		converter = NewSplunkConverter()
	}
	return converter.Convert(input, input_filename)
}

type IConverter interface {
	Convert(input []byte, filename string) (*cacao.Playbook, error)
}
