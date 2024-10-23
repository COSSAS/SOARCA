package thehive

type ITheHiveConnector interface {
	Hello() string
}

// The TheHive connector itself

type TheHiveConnector struct {
}

func (thehiveConnector *TheHiveConnector) Hello() string {
	return "hello"
}
