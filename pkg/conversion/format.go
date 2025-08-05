package conversion

import "strings"

type TargetFormat int

const (
	FormatBpmn TargetFormat = iota
	FormatSplunk
	FormatMisp
	FormatStix
	FormatOpenC2
	FormatTaxii
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
	case "misp":
		return FormatMisp
	case "splunk":
		return FormatSplunk
	}
	return FormatUnknown
}
