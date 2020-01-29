package hub

type AgentStarter interface {
	StartAgent(hostname, stateDir string)
}

func StartAgentsSubStep(
	hostnames []string,
	stateDir string,
	agentStarter AgentStarter) {

	for _, hostname := range hostnames {
		agentStarter.StartAgent(hostname, stateDir)
	}
}
