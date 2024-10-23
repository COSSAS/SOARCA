package catalyst

type ICatalystConnector interface {
	Hello() string
}

// The Catalyst connector itself

type CatalystConnector struct {
}

func (catalystConnector *CatalystConnector) Hello() string {
	return "hello"
}
