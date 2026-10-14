package conversion

import (
	model "soarca/pkg/models/conversion"
	"strings"
)

func GuessFormat(filename string) model.TargetFormat {
	if strings.HasSuffix(filename, "bpmn") {
		return model.FormatBpmn
	}
	return model.FormatUnknown
}
func ReadFormat(format string) model.TargetFormat {
	switch format {
	case "bpmn":
		return model.FormatBpmn
	}
	return model.FormatUnknown
}
