package conversion

import "strings"

type TargetFormat int

const (
	FormatBpmn TargetFormat = iota
	FormatUnknown
)

func guess_format(filename string) TargetFormat {
	if strings.HasSuffix(filename, "bpmn") {
		return FormatBpmn
	}
	return FormatUnknown
}
func read_format(format string) TargetFormat {
	switch format {
	case "bpmn":
		return FormatBpmn
	}
	return FormatUnknown
}
