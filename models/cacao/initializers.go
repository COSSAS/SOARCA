package cacao

func NewAgentTargets(new ...AgentTarget) AgentTargets {
	agents := make(AgentTargets)
	for _, agent := range new {
		agents[agent.ID] = agent
	}
	return agents
}

func NewAuthenticationInfoDefinitions(new ...AuthenticationInformation) AuthenticationInformations {
	info := make(AuthenticationInformations)
	for _, item := range new {
		info[item.ID] = item
	}
	return info
}

func NewExtensionDefinitions(new ...ExtensionDefinition) ExtensionDefinitions {
	info := make(ExtensionDefinitions)
	for _, item := range new {
		info[item.ID] = item
	}
	return info
}

func NewDataMarkings(new ...DataMarking) DataMarkings {
	info := make(DataMarkings)
	for _, item := range new {
		info[item.ID] = item
	}
	return info
}
