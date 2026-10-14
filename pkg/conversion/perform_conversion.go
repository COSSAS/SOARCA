package conversion

import (
	"errors"
	bpmn_conversion "soarca/pkg/conversion/bpmn"
	"soarca/pkg/models/cacao"
	model "soarca/pkg/models/conversion"
	util "soarca/pkg/utils/conversion"
)

func PerformConversion(input_filename string, input []byte, format_string string) (*cacao.Playbook, error) {
	var format model.TargetFormat
	if format_string == "" {
		format = util.GuessFormat(input_filename)
	} else {
		format = util.ReadFormat(format_string)
	}
	if format == model.FormatUnknown {
		return nil, errors.New("could not deduce input file type")
	}
	var converter IConverter
	switch format {
	case model.FormatBpmn:
		converter = bpmn_conversion.NewBpmnConverter()
	}
	return converter.Convert(input, input_filename)
}
